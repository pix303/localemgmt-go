package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/pix303/localemgmt-go/api/internal/dto"
	"net/http"
)

func WelcomeWithMessageHandler(c echo.Context) error {
	result := dto.Message{
		Content: "Welcome in Localemgmt API",
	}
	return c.JSON(http.StatusOK, result)
}
