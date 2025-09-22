package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/pix303/localemgmt-go/api/internal/dto"
)

func (reqHandler *LocaleItemHandler) GetDetail(ctx echo.Context) error {
	payload := dto.GetDetailRequest{}
	err := ctx.Bind(&payload)
	if err != nil {
		return err
	}

	return nil
}
