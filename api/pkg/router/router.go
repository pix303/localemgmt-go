package router

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
