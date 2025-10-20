package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/localemgmt-go/domain/pkg/user"
	"github.com/pix303/localemgmt-go/domain/pkg/usersession"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	userInfoUrl = "https://www.googleapis.com/oauth2/v3/userinfo"
	sessionKey  = "session_id"
)

var oAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  "",
	Scopes:       []string{"openid", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

var (
	ErrStoreUserInfo    = echo.NewHTTPError(http.StatusInternalServerError, "fail to store user info")
	ErrStoreUserSession = echo.NewHTTPError(http.StatusInternalServerError, "fail to store user session")
)

type UserHandler struct {
	mutex         sync.RWMutex
	stateRequests map[string]time.Time
}

func NewUserHandler() UserHandler {
	return UserHandler{
		mutex:         sync.RWMutex{},
		stateRequests: make(map[string]time.Time),
	}
}

func (handler *UserHandler) Login(ctx echo.Context) error {
	slog.Debug("start login")
	state := uuid.New().String()
	handler.mutex.Lock()
	handler.stateRequests[state] = time.Now()
	handler.mutex.Unlock()

	baseUrl := os.Getenv("BASE_URL")
	oAuthConfig.RedirectURL = fmt.Sprintf("%s/%s", baseUrl, "api/v1/auth-callback")
	redirectUrl := oAuthConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)

	err := ctx.JSON(http.StatusOK, map[string]string{"url": redirectUrl})
	if err != nil {
		delete(handler.stateRequests, state)
		return err
	}
	slog.Debug("end  login")
	return nil
}

func (handler *UserHandler) AuthCallback(ctx echo.Context) error {
	state := ctx.QueryParam("state")
	code := ctx.QueryParam("code")
	handler.mutex.Lock()
	storedState, ok := handler.stateRequests[state]
	if ok {
		delete(handler.stateRequests, state)
	} else {
		return errors.New("invalid state")
	}
	handler.mutex.Unlock()

	if time.Since(storedState) > time.Minute*10 {
		return errors.New("state expired")
	}

	token, err := oAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return err
	}

	userInfoHttpClient := oAuthConfig.Client(context.Background(), token)
	resp, err := userInfoHttpClient.Get(userInfoUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	result := user.UserInfoReponseBody{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	result.RefreshToken = token.RefreshToken
	slog.Debug("user info", slog.Any("user", result))

	userMsg := actor.NewMessage(
		user.UserActorAddress,
		nil,
		user.UpdateUserMessageBody{User: result},
		true,
	)

	_, err = actor.SendMessageWithResponse[any](userMsg)
	if err != nil {
		slog.Error("fail to store user info", slog.Any("error", err))
		return ErrStoreUserInfo
	}

	sessionMsg := actor.NewMessage(
		usersession.UserSessionActorAddress,
		nil,
		usersession.CreateSessionItemMsgBody{SubjectID: result.SubjectID, RefreshToken: token.RefreshToken},
		true,
	)

	sessionResult, err := actor.SendMessageWithResponse[usersession.CreateSessionItemMsgBodyResult](sessionMsg)
	if err != nil {
		slog.Error("fail to store user session ", slog.Any("error", err))
		return echo.NewHTTPError(ErrStoreUserSession.Code, fmt.Errorf("%w: %v", ErrStoreUserSession, err))
	}

	baseUrl := os.Getenv("BASE_URL")
	isProd := os.Getenv("STAGE")
	ctx.SetCookie(&http.Cookie{
		Name:     sessionKey,
		Value:    sessionResult.SessionID,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: true,
		Secure:   isProd == "prod",
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	err = ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/%s", baseUrl, "fe/"))
	if err != nil {
		slog.Error("fail to respond with session id", slog.Any("error", err))
		return echo.NewHTTPError(ErrStoreUserSession.Code, fmt.Errorf("%w: %v", ErrStoreUserSession, err))
	}

	return nil
}

func (handler *UserHandler) GetInfo(ctx echo.Context) error {
	start := time.Now()
	slog.Debug("get user info")
	cookie, err := ctx.Cookie(sessionKey)
	if err != nil {
		return ctx.NoContent(http.StatusUnauthorized)
	}

	sessionMsg := actor.NewMessage(
		usersession.UserSessionActorAddress,
		nil,
		usersession.RetriveSessionMsgBody{SessionId: cookie.Value},
		true,
	)

	sessionResult, err := actor.SendMessageWithResponse[usersession.RetriveSessionMsgBodyResult](sessionMsg)
	if err != nil {
		return ctx.NoContent(http.StatusUnauthorized)
	}

	userMsg := actor.NewMessage(
		user.UserActorAddress,
		nil,
		user.RetriveUserMessageBody{SubjectID: sessionResult.Session.SubjectID},
		true,
	)

	userResult, err := actor.SendMessageWithResponse[user.RetriveUserMessageBodyResult](userMsg)
	if err != nil {
		return ctx.NoContent(http.StatusInternalServerError)
	}
	slog.Debug("get user info end", slog.Float64("delta", time.Since(start).Seconds()))
	return ctx.JSON(http.StatusOK, userResult)
}

func (handler *UserHandler) SessionValidator() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			start := time.Now()
			slog.Debug("start session validator")
			sessionCookie, err := ctx.Cookie(sessionKey)
			if err != nil {
				if err == http.ErrNoCookie {
					return ctx.NoContent(http.StatusUnauthorized)
				} else {
					return ctx.NoContent(http.StatusBadRequest)
				}
			}

			msg := actor.NewMessage(
				usersession.UserSessionActorAddress,
				nil,
				usersession.RetriveSessionMsgBody{SessionId: sessionCookie.Value},
				true,
			)

			sessionResult, err := actor.SendMessageWithResponse[usersession.RetriveSessionMsgBodyResult](msg)
			if err == sql.ErrNoRows {
				return ctx.NoContent(http.StatusUnauthorized)
			}
			if err != nil {
				slog.Error("fail to retrive user session ", slog.Any("error", err))
				return ctx.NoContent(http.StatusUnauthorized)
			}

			sessionData := sessionResult.Session
			if sessionData.ExpireAt.After(time.Now()) && sessionData.RefreshCounter < 10 {
				slog.Debug("end ---session validator", slog.Float64("delta", time.Since(start).Seconds()))
				return next(ctx)
			}
			return ctx.NoContent(http.StatusUnauthorized)
		}
	}
}
