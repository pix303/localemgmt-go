package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/eventstore-go-v2/pkg/store"
	"github.com/pix303/localemgmt-go/api/internal/dto"
	"github.com/pix303/localemgmt-go/domain/pkg/localeitem/events"
)

var (
	ErrTimeout                  = echo.NewHTTPError(http.StatusRequestTimeout, "Error timout")
	ErrVerifyRequest            = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying request parameters")
	ErrVerifyAggregateExistence = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying existence of aggregateID")
	ErrEventStore               = echo.NewHTTPError(http.StatusInternalServerError, "Error eventstore")
	ErrStoreCreateEvent         = echo.NewHTTPError(http.StatusInternalServerError, "Error on store creating locale item event")
	ErrStoreUpdateEvent         = echo.NewHTTPError(http.StatusInternalServerError, "Error on store updating locale item event")
)

// func wrapError(err *echo.HTTPError, message string) *echo.HTTPError {
// 	return &echo.HTTPError{
// 		Code:    err.Code,
// 		Message: fmt.Sprintf("%s: %s", err.Message, message),
// 	}
// }

type LocaleItemHandler struct {
	EventStoreActor actor.Actor
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
func (reqHandler *LocaleItemHandler) CreateLocaleItem(c echo.Context) error {
	payload := dto.CreateRequest{}
	err := c.Bind(&payload)
	if err != nil {
		return err
	}

	// verify request
	if payload.Content == "" || payload.Context == "" || payload.Lang == "" {
		return ErrVerifyRequest
	}

	evt, err := events.NewCreateEvent(payload.Content, payload.Context, payload.Lang, "todo")

	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFunc()
	returnChan := make(chan actor.Message)
	msg := actor.NewMessage(
		store.EventStoreAddress,
		LocaleItemHandlerAddress,
		store.AddEventBody{Event: evt},
		returnChan,
	)

	err = reqHandler.EventStoreActor.Send(msg)
	if err != nil {
		return ErrStoreCreateEvent
	}

	select {
	case returnMsg := <-returnChan:
		if body, ok := returnMsg.Body.(store.AddEventBodyResult); ok {
			if body.Success {
				return c.JSON(http.StatusOK, evt)
			} else {
				return ErrStoreCreateEvent
			}
		}
	case <-ctx.Done():
		return ErrEventStore
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

	ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFunc()
	returnChan := make(chan actor.Message)
	msg := actor.NewMessage(
		store.EventStoreAddress,
		LocaleItemHandlerAddress,
		store.CheckExistenceByAggregateIDBody{Id: payload.AggregateId},
		returnChan,
	)

	err = reqHandler.EventStoreActor.Send(msg)
	if err != nil {
		return ErrVerifyAggregateExistence
	}

	select {
	case returnMsg := <-returnChan:
		if body, ok := returnMsg.Body.(store.CheckExistenceByAggregateIDBodyResult); ok {
			if !body.Exists {
				return ErrVerifyAggregateExistence
			}
		}
	case <-ctx.Done():
		return ErrTimeout
	}

	// add update event
	evt, err := events.NewUpdateEvent(payload.AggregateId, payload.Content, payload.Lang, "todo")
	if err != nil {
		return err
	}

	msg.Body = store.AddEventBody{
		Event: evt,
	}

	err = reqHandler.EventStoreActor.Send(msg)
	if err != nil {
		return ErrStoreUpdateEvent
	}

	returnMsg := <-returnChan
	if body, ok := returnMsg.Body.(store.AddEventBodyResult); ok {
		if body.Success {
			return c.JSON(http.StatusOK, evt)
		} else {
			return ErrStoreUpdateEvent
		}
	}

	return ErrStoreUpdateEvent
}
