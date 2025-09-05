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
	ErrorTimeout                  = echo.NewHTTPError(http.StatusRequestTimeout, "Error timout")
	ErrorVerifyRequest            = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying request parameters")
	ErrorVerifyAggregateExistence = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying existence of aggregateID")
	ErrorEventStore               = echo.NewHTTPError(http.StatusInternalServerError, "Error eventstore")
	ErrorStoreCreateEvent         = echo.NewHTTPError(http.StatusInternalServerError, "Error on store creating locale item event")
	ErrorStoreUpdateEvent         = echo.NewHTTPError(http.StatusInternalServerError, "Error on store updating locale item event")
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
		return ErrorVerifyRequest
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

	reqHandler.EventStoreActor.Send(msg)

	select {
	case returnMsg := <-returnChan:
		if body, ok := returnMsg.Body.(store.AddEventBodyResult); ok {
			if body.Success {
				return c.JSON(http.StatusOK, evt)
			} else {
				return ErrorStoreCreateEvent
			}
		}
	case <-ctx.Done():
		return ErrorEventStore
	}

	return ErrorStoreCreateEvent
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
		return ErrorVerifyRequest
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

	reqHandler.EventStoreActor.Send(msg)
	select {
	case returnMsg := <-returnChan:
		if body, ok := returnMsg.Body.(store.CheckExistenceByAggregateIDBodyResult); ok {
			if !body.Exists {
				return ErrorVerifyAggregateExistence
			}
		}
	case <-ctx.Done():
		return ErrorTimeout
	}

	// add update event
	evt, err := events.NewUpdateEvent(payload.AggregateId, payload.Content, payload.Lang, "todo")
	if err != nil {
		return err
	}

	msg.Body = store.AddEventBody{
		Event: evt,
	}
	reqHandler.EventStoreActor.Send(msg)

	returnMsg := <-returnChan
	if body, ok := returnMsg.Body.(store.AddEventBodyResult); ok {
		if body.Success {
			return c.JSON(http.StatusOK, evt)
		} else {
			return ErrorStoreUpdateEvent
		}
	}

	return ErrorStoreUpdateEvent
}
