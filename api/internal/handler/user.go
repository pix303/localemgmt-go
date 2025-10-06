package handler

import (
	"context"
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

const userInfoUrl = "https://www.googleapis.com/oauth2/v3/userinfo"

var oAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  "api/v1/user/auth-callback",
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
	state := uuid.New().String()
	handler.mutex.Lock()
	handler.stateRequests[state] = time.Now()
	handler.mutex.Unlock()

	baseUrl := os.Getenv("BASE_URL")
	oAuthConfig.RedirectURL = string(fmt.Append([]byte(baseUrl), oAuthConfig.RedirectURL))
	redirectUrl := oAuthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)

	err := ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)
	if err != nil {
		delete(handler.stateRequests, state)
		return err
	}
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

	_, err = actor.SendMessageWithResponse(userMsg)
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

	_, err = actor.SendMessageWithResponse(sessionMsg)
	if err != nil {
		slog.Error("fail to store user session ", slog.Any("error", err))
		return echo.NewHTTPError(ErrStoreUserSession.Code, fmt.Sprintf("%s: %s", ErrStoreUserSession.Message, err.Error()))
	}

	return nil
}

func (handler *UserHandler) GetInfo(ctx echo.Context) error {
	return nil
}
