package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pix303/cinecity/pkg/actor"
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

	localeHandler, err := handler.NewLocaleItemHandler()
	if err != nil {
		return nil, err
	}

	localeItemGroup := apiGroup.Group("/localeitem")
	localeItemGroup.POST("/create", localeHandler.CreateLocaleItem)
	localeItemGroup.POST("/update", localeHandler.UpdateTranslation)
	localeItemGroup.GET("/detail/:id", localeHandler.GetDetail)
	localeItemGroup.GET("/context/:id", localeHandler.GetContext)

	router := LocaleItemRouter{r}
	return &router, nil
}

type StartRouter struct{}
type LocaleItemRouterState struct {
	RouterInstance *LocaleItemRouter
	server         *http.Server
	isRunning      bool
	mutex          sync.Mutex
}

func NewLocaleItemRouterState() (*LocaleItemRouterState, error) {
	r, err := NewRouter()
	if err != nil {
		return nil, err
	}
	initState := LocaleItemRouterState{
		RouterInstance: r,
		mutex:          sync.Mutex{},
	}
	return &initState, nil
}

const Port = 8083

func (this *LocaleItemRouterState) Process(msg actor.Message) {
	switch msg.Body.(type) {
	case StartRouter:
		if !this.isRunning {
			this.mutex.Lock()
			slog.Info("Starting server", slog.Int("port", Port))

			this.server = &http.Server{
				Addr:    fmt.Sprintf("localhost:%d", Port),
				Handler: this.RouterInstance.Router,
			}

			go func() {
				err := this.server.ListenAndServe()
				if err != nil && err != http.ErrServerClosed {
					slog.Error("http server fail:", slog.String("err", err.Error()))
				}
			}()

			this.server.RegisterOnShutdown(func() {
				slog.Info("Server shutdown completely")
			})

			this.isRunning = true
			this.mutex.Unlock()
		} else {
			slog.Info("Server is already running, start message ignored")
		}

	default:
		slog.Warn("Unable to process unknown message", slog.String("msg", msg.String()))
	}
}

func (this *LocaleItemRouterState) GetState() any {
	return nil
}

func (this *LocaleItemRouterState) Shutdown() {
	if this.isRunning {
		this.mutex.Lock()

		slog.Info("Stopping server on port 8080")
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
		defer cancelFunc()

		err := this.server.Shutdown(ctx)
		if err != nil {
			//it's not a real error but a info
			if err == http.ErrServerClosed {
				slog.Info("http server stopped on port 8080 with success", slog.String("err", err.Error()))
			} else {
				slog.Error("fail on stopping http server", slog.String("err", err.Error()))
			}
			return
		}

		this.mutex.Unlock()
	} else {
		slog.Info("Server is already stopped on port 8080")
	}
}
