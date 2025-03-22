package router

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pix303/actor-lib/pkg/actor"
	"github.com/pix303/localemgmt-go/api/internal/handler"
)

var apiVersion = "api/v1"

type LocaleItemRouter struct {
	Router *echo.Echo
}

func NewRouter() (*LocaleItemRouter, error) {

	r := echo.New()
	r.Use(middleware.Logger())

	apiGroup := r.Group(apiVersion)
	apiGroup.GET("/", handler.WelcomeWithMessageHandler)

	localeHanderl, err := handler.NewLocaleItemHandler()
	if err != nil {
		return nil, err
	}

	localeItemGroup := apiGroup.Group("/localeitem")
	localeItemGroup.POST("/create", localeHanderl.CreateLocaleItem)
	localeItemGroup.POST("/update", localeHanderl.UpdateTranslation)

	router := LocaleItemRouter{r}
	return &router, nil
}

type StartRouter int
type StopRouter int
type LocaleItemRouterState struct {
	RouterInstance *LocaleItemRouter
}

func NewLocaleItemRouterState() *LocaleItemRouterState {
	r, err := NewRouter()
	if err != nil {
		slog.Error("fail to start router", slog.String("err", err.Error()))
		return nil
	}
	initState := LocaleItemRouterState{
		RouterInstance: r,
	}
	return &initState
}

func (this *LocaleItemRouterState) Process(inbox chan actor.Message) {
	for {
		msg := <-inbox
		switch msg.Body.(type) {
		case StartRouter:
			slog.Error("Error: %s", "error", http.ListenAndServe(":8080", this.RouterInstance.Router).Error())
		case StopRouter:
			err := this.RouterInstance.Router.Shutdown(context.Background())
			if err != nil {
				slog.Error("fail to stop router", slog.String("err", err.Error()))
			}
		}
	}
}
