package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/crypt-util-go/pkg/crypt"
	"github.com/pix303/localemgmt-go/domain/pkg/user"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	userInfoUrl   = "https://www.googleapis.com/oauth2/v3/userinfo"
	cookieKey     = "token"
	subjectKey    = "subjectId"
	tokenLifetime = time.Hour
	maxRefresh    = 10
)

var oAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  fmt.Sprintf("%s/%s", os.Getenv("BASE_URL"), "api/v1/auth-callback"),
	Scopes:       []string{"openid", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

var (
	ErrStoreUserInfo        = echo.NewHTTPError(http.StatusInternalServerError, "fail to store user info")
	ErrStoreUserSession     = echo.NewHTTPError(http.StatusInternalServerError, "fail to store user session")
	ErrCreateJWTUserSession = echo.NewHTTPError(http.StatusInternalServerError, "fail to create user session cookie value")
	ErrRetriveUserInfo      = echo.NewHTTPError(http.StatusInternalServerError, "fail to retrive user info")
	ErrRevokeGoogleTokens   = echo.NewHTTPError(http.StatusInternalServerError, "fail to revoke tokens")
)

type UserHandler struct {
	mutex         sync.RWMutex
	stateRequests map[string]time.Time
}

type AccessTokenClaim struct {
	RefreshCounter uint8 `json:"refreshCounter"`
	jwt.RegisteredClaims
}

func NewUserHandler() UserHandler {
	return UserHandler{
		mutex:         sync.RWMutex{},
		stateRequests: make(map[string]time.Time),
	}
}

// Login is responsable for handle start login request
func (handler *UserHandler) Login(ctx echo.Context) error {
	// generate state and store it
	state := uuid.New().String()
	handler.mutex.Lock()
	handler.stateRequests[state] = time.Now()
	handler.mutex.Unlock()

	// request to Google Auth provider
	redirectUrl := oAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	// return to fe url to call for SSO
	err := ctx.JSON(http.StatusOK, map[string]string{"url": redirectUrl})
	if err != nil {
		delete(handler.stateRequests, state)
		return err
	}
	return nil
}

// AuthCallback is responsable for handle auth callback from google
func (handler *UserHandler) AuthCallback(ctx echo.Context) error {
	// state must match with request authorization state
	state := ctx.QueryParam("state")
	code := ctx.QueryParam("code")

	// check if state is valid and match with stateRequests
	handler.mutex.Lock()
	storedState, ok := handler.stateRequests[state]
	if ok {
		delete(handler.stateRequests, state)
	} else {
		return errors.New("invalid state")
	}
	handler.mutex.Unlock()

	// check if there is timeout between request and callback
	if time.Since(storedState) > time.Minute*10 {
		return errors.New("state expired")
	}

	// request to Google tokens with returned code from Google Auth provider
	token, err := oAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return err
	}

	// request user info with access token
	userInfoHttpClient := oAuthConfig.Client(context.Background(), token)
	resp, err := userInfoHttpClient.Get(userInfoUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read user info from response body
	userInfoResponse := user.UserInfoReponseBody{}
	err = json.NewDecoder(resp.Body).Decode(&userInfoResponse)
	if err != nil {
		return err
	}
	userInfoResponse.RefreshToken = token.RefreshToken

	// store user info
	userMsg := actor.NewMessage(
		user.UserActorAddress,
		nil,
		user.UpdateUserMessageBody{User: userInfoResponse},
		true,
	)

	_, err = actor.SendMessageWithResponse[user.UpdateUserMessageBodyResult](userMsg)
	if err != nil {
		slog.Error("fail to store user info", slog.Any("error", err))
		return ErrStoreUserInfo
	}

	jwtString, err := newJwtToken(userInfoResponse.SubjectID, 0)
	if err != nil {
		slog.Error("fail to obtain new jwt", slog.Any("error", err))
		return ErrCreateJWTUserSession
	}

	// redirect with cookie with access token
	setCookie(ctx, jwtString)
	baseUrl := os.Getenv("BASE_URL")
	err = ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/%s", baseUrl, "fe/"))
	if err != nil {
		slog.Error("fail to respond with session cookie", slog.Any("error", err))
		return echo.NewHTTPError(ErrStoreUserSession.Code, fmt.Errorf("%w: %v", ErrStoreUserSession, err))
	}

	return nil
}

func getUserInfo(ctx echo.Context) (user.User, error) {
	subId, ok := ctx.Get(subjectKey).(string)
	if !ok {
		return user.User{}, errors.New("no subject id")
	}

	userMsg := actor.NewMessage(
		user.UserActorAddress,
		nil,
		user.RetriveUserMessageBody{SubjectID: subId},
		true,
	)

	userResult, err := actor.SendMessageWithResponse[user.RetriveUserMessageBodyResult](userMsg)
	if err != nil {
		return user.User{}, errors.New("fail to retrive user info")
	}

	return userResult.User, nil
}

// GetInfo returns user info
func (handler *UserHandler) GetInfo(ctx echo.Context) error {
	user, err := getUserInfo(ctx)
	if err != nil {
		slog.Error("fail to get user info", slog.Any("error", err))
		return ErrRetriveUserInfo
	}
	user.RefreshToken = ""
	return ctx.JSON(http.StatusOK, user)
}

// Logout revokes the refresh token and clears the cookie
func (handler *UserHandler) Logout(ctx echo.Context) error {
	user, err := getUserInfo(ctx)
	if err != nil {
		slog.Error("fail to get user info", slog.Any("error", err))
		return ErrRetriveUserInfo
	}

	rt, err := crypt.Decrypt(user.RefreshToken, []byte(os.Getenv("REFRESH_TOKEN_CKEY")))
	if err != nil {
		slog.Error("fail to decrypt refresh token", slog.Any("error", err))
		return ErrRetriveUserInfo
	}

	const revokeURL = "https://oauth2.googleapis.com/revoke"
	payload := url.Values{}
	payload.Set("token", rt)
	req, err := http.NewRequestWithContext(context.Background(), "POST", revokeURL, strings.NewReader(payload.Encode()))
	if err != nil {
		slog.Error("fail to build revoke request", slog.Any("error", err))
		return ErrRetriveUserInfo
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpClient := http.Client{}
	res, err := httpClient.Do(req)
	if err != nil {
		slog.Error("failed to revoke token", slog.String("error", err.Error()))
		return ErrRevokeGoogleTokens
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		slog.Error("failed to revoke token", slog.Int("statusCode", res.StatusCode))
		return ErrRevokeGoogleTokens
	}

	slog.Info("token revoke successfully")

	ctx.SetCookie(&http.Cookie{
		Name:     cookieKey,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	return ctx.NoContent(http.StatusOK)
}

func newJwtToken(subject string, refreshCounter uint8) (string, error) {
	atc := AccessTokenClaim{
		RefreshCounter: refreshCounter + 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    "localemgmt",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenLifetime)),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, atc)
	key := os.Getenv("JWT_KEY")
	jwtString, err := jwtToken.SignedString([]byte(key))
	return jwtString, err
}

func setCookie(ctx echo.Context, jwt string) {
	isProd := os.Getenv("STAGE")
	ctx.SetCookie(&http.Cookie{
		Name:     cookieKey,
		Value:    jwt,
		Expires:  time.Now().Add(tokenLifetime).Add(time.Minute),
		HttpOnly: true,
		Secure:   isProd == "prod",
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func validateToken(tokenSource string) (*AccessTokenClaim, error) {
	atc := AccessTokenClaim{}
	token, err := jwt.ParseWithClaims(tokenSource, &atc, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			slog.Error("invalid signing method", slog.String("method", token.Method.Alg()))
			return nil, errors.New("invalid signing method")
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		slog.Error("fail parsing token", slog.String("error", err.Error()))
		return nil, err
	}

	if !token.Valid {
		slog.Error("fail validating token")
		return nil, errors.New("token is invalid")
	}

	return &atc, nil
}

func (handler *UserHandler) SessionValidator() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			accessTokenCookie, err := ctx.Cookie(cookieKey)
			if err != nil {
				if err == http.ErrNoCookie {
					return ctx.NoContent(http.StatusUnauthorized)
				} else {
					return ctx.NoContent(http.StatusBadRequest)
				}
			}

			atclaims, err := validateToken(accessTokenCookie.Value)
			if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
				slog.Error("fail to validate token", slog.Any("error", err))
				return ctx.NoContent(http.StatusUnauthorized)
			}

			if atclaims.RefreshCounter >= maxRefresh {
				return ctx.NoContent(http.StatusUnauthorized)
			}

			if atclaims.ExpiresAt.Before(time.Now()) {
				slog.Info("token expired, creating new one")
				newJwtString, err := newJwtToken(atclaims.Subject, atclaims.RefreshCounter)
				if err != nil {
					slog.Error("fail to create new jwt token", slog.Any("error", err))
					return ctx.NoContent(http.StatusUnauthorized)
				}
				setCookie(ctx, newJwtString)
			}

			ctx.Set(subjectKey, atclaims.Subject)
			return next(ctx)
		}
	}
}
