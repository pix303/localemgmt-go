package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/eventstore-go-v2/pkg/store"
	"github.com/pix303/localemgmt-go/api/internal/dto"
	"github.com/pix303/localemgmt-go/domain/pkg/localeitem/aggregate"
	"github.com/pix303/localemgmt-go/domain/pkg/localeitem/events"
)

var (
	ErrTimeout                  = echo.NewHTTPError(http.StatusRequestTimeout, "Error timout")
	ErrVerifyRequest            = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying request parameters: lang and content")
	ErrVerifyAggregateExistence = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying existence of aggregateID")
	ErrEventStore               = echo.NewHTTPError(http.StatusInternalServerError, "Error eventstore")
	ErrStoreCreateEvent         = echo.NewHTTPError(http.StatusInternalServerError, "Error on store creating locale item event")
	ErrStoreUpdateEvent         = echo.NewHTTPError(http.StatusInternalServerError, "Error on store updating locale item event")
)

type LocaleItemHandler struct {
}

func NewLocaleItemHandler() (LocaleItemHandler, error) {

	es, err := store.NewEvenStoreActorWithPostgres()
	if err != nil {
		return LocaleItemHandler{}, err
	}

	err = actor.RegisterActor(&es)
	if err != nil {
		return LocaleItemHandler{}, err
	}

	return LocaleItemHandler{}, nil
}

func (handler *LocaleItemHandler) GetDetail(ctx echo.Context) error {
	aggregateId := ctx.Param("id")
	msg := actor.NewMessage(
		aggregate.LocaleItemAggregateDetailAddress,
		nil,
		aggregate.GetLocaleItemAggregateDetailBody{
			Id: aggregateId,
		},
		true,
	)
	resultMsg, err := actor.SendMessageWithResponse(msg)
	if err != nil {
		return err
	}

	err = ctx.JSON(http.StatusOK, resultMsg.Body)
	if err != nil {
		return err
	}
	return nil
}

func (handler *LocaleItemHandler) GetContext(ctx echo.Context) error {
	contextId := ctx.Param("id")
	msg := actor.NewMessage(
		aggregate.LocaleItemAggregateListAddress,
		nil,
		aggregate.GetContextBody{
			Id: contextId,
		},
		true,
	)
	resultMsg, err := actor.SendMessageWithResponse(msg)
	if err != nil {
		return err
	}

	err = ctx.JSON(http.StatusOK, resultMsg.Body)
	if err != nil {
		return err
	}
	return nil
}

// CreateLocaleItem add crate locale item event
func (handler *LocaleItemHandler) CreateLocaleItem(c echo.Context) error {
	payload := dto.CreateRequest{}
	err := c.Bind(&payload)
	if err != nil {
		return err
	}

	// verify request
	if payload.Content == "" || payload.Lang == "" {
		return ErrVerifyRequest
	}

	if payload.Context == "" {
		payload.Context = aggregate.DEFAULT_CONTEXT
	}

	// TODO: add check if for content + lang + context something exists

	evt, err := events.NewCreateEvent(payload.Content, payload.Context, payload.Lang, "todo")

	if err != nil {
		return err
	}

	msg := actor.NewMessage(
		store.EventStoreAddress,
		nil,
		store.AddEventBody{Event: evt},
		true,
	)

	responseMsg, err := actor.SendMessageWithResponse(msg)
	if err != nil {
		return ErrStoreCreateEvent
	}

	if body, ok := responseMsg.Body.(store.AddEventBodyResult); ok {
		if body.Success {
			return c.JSON(http.StatusOK, evt)
		} else {
			return ErrStoreCreateEvent
		}
	}

	return ErrStoreCreateEvent
}

// UpdateTranslation add add or update locale item translation event
func (reqHandler *LocaleItemHandler) UpdateTranslation(c echo.Context) error {
	payload := dto.UpdateRequest{}
	err := c.Bind(&payload)
	if err != nil {
		return err
	}

	// verify request
	if payload.AggregateId == "" || payload.Content == "" || payload.Lang == "" {
		return ErrVerifyRequest
	}

	// check aggregate id presence
	checkMsg := actor.NewMessage(
		store.EventStoreAddress,
		nil,
		store.CheckExistenceByAggregateIDBody{Id: payload.AggregateId},
		true,
	)

	checkResponseMsg, err := actor.SendMessageWithResponse(checkMsg)
	if responseBody, ok := checkResponseMsg.Body.(store.CheckExistenceByAggregateIDBodyResult); ok {
		if !responseBody.Exists {
			return ErrVerifyAggregateExistence
		}
	}

	// add update event
	evt, err := events.NewUpdateEvent(payload.AggregateId, payload.Content, payload.Lang, "todo")
	if err != nil {
		return err
	}

	addEventMsg := actor.NewMessage(
		store.EventStoreAddress,
		nil,
		store.AddEventBody{Event: evt},
		true,
	)

	addResponseMsg, err := actor.SendMessageWithResponse(addEventMsg)
	if err != nil {
		return ErrStoreUpdateEvent
	}

	if body, ok := addResponseMsg.Body.(store.AddEventBodyResult); ok {
		if body.Success {
			return c.JSON(http.StatusOK, evt)
		} else {
			return ErrStoreUpdateEvent
		}
	}

	return ErrStoreUpdateEvent
}
