package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pix303/actor-lib/pkg/actor"
	"github.com/pix303/eventstore-go-v2/pkg/store"
	"github.com/pix303/localemgmt-go/api/internal/dto"
	"github.com/pix303/localemgmt-go/domain/pkg/localeitem/events"
)

var (
	ErrorVerifyRequest            = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying request parameters")
	ErrorVerifyAggregateExistence = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying existence of aggregateID")
	ErrorEventStore               = echo.NewHTTPError(http.StatusInternalServerError, "Error eventstore")
	ErrorStoreCreateEvent         = echo.NewHTTPError(http.StatusInternalServerError, "Error on store creating locale item event")
	ErrorStoreUpdateEvent         = echo.NewHTTPError(http.StatusInternalServerError, "Error on store updating locale item event")
)

func wrapError(err *echo.HTTPError, message string) *echo.HTTPError {
	return &echo.HTTPError{
		Code:    err.Code,
		Message: fmt.Sprintf("%s: %s", err.Message, message),
	}
}

type LocaleItemHandler struct {
	eventStoreActor actor.Actor
}

var LocaleItemHandlerAddress = actor.NewAddress("locale", "localeitem-handler")

func NewLocaleItemHandler() (LocaleItemHandler, error) {

	es, err := store.NewEvenStoreActorWithPostgres()
	if err != nil {
		return LocaleItemHandler{}, err
	}

	actor.RegisterActor(&es)

	return LocaleItemHandler{
		es,
	}, nil
}

// CreateLocaleItem add crate locale item event
func (h *LocaleItemHandler) CreateLocaleItem(c echo.Context) error {
	payload := dto.CreateRequest{}
	err := c.Bind(&payload)
	if err != nil {
		return err
	}

	// verify request
	if payload.Content == "" || payload.Context == "" || payload.Lang == "" {
		return ErrorVerifyRequest
	}

	evt, err := events.NewCreateEvent(payload.Content, payload.Context, payload.Lang, "todo")

	if err != nil {
		return err
	}

	msg := actor.Message{
		To:   store.EventStoreAddress(),
		From: nil,
		Body: store.AddEventBody{Event: evt},
	}

	returnMsg, err := actor.DispatchMessageWithReturn(msg)
	if err != nil {
		return ErrorEventStore
	}

	if body, ok := returnMsg.Body.(store.ResultAddEventBody); ok {
		if body.Success {
			return c.JSON(http.StatusOK, evt)
		}
	}

	return ErrorStoreCreateEvent
}

// UpdateTranslation add add or update locale item translation event
func (h *LocaleItemHandler) UpdateTranslation(c echo.Context) error {
	payload := dto.UpdateRequest{}
	err := c.Bind(&payload)
	if err != nil {
		return err
	}

	// verify request
	if payload.AggregateId == "" || payload.Content == "" || payload.Lang == "" {
		return ErrorVerifyRequest
	}

	msg := actor.NewMessage(
		store.EventStoreAddress(),
		LocaleItemHandlerAddress,
		store.CheckExistenceByAggregateIDBody{Id: payload.AggregateId},
	)

	result, err := actor.DispatchMessageWithReturn(msg)
	if err != nil {
		return wrapError(ErrorEventStore, err.Error())
	}

	if body, ok := result.Body.(store.CheckExistenceByAggregateIDBodyResult); ok {
		if !body.Exists {
			return ErrorVerifyAggregateExistence
		}
	}

	// add update event
	evt, err := events.NewUpdateEvent(payload.AggregateId, payload.Content, payload.Lang, "todo")
	if err != nil {
		return err
	}

	msg.Body = store.AddEventBody{
		Event: evt,
	}

	result, err = actor.DispatchMessageWithReturn(msg)
	if err != nil {
		return wrapError(ErrorStoreUpdateEvent, err.Error())
	}

	return c.JSON(http.StatusOK, evt)
}
