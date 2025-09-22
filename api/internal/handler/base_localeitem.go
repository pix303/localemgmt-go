package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/eventstore-go-v2/pkg/store"
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
	EventStoreActor actor.Actor
}

var LocaleItemHandlerAddress = actor.NewAddress("locale", "localeitem-handler")

func NewLocaleItemHandler() (LocaleItemHandler, error) {

	es, err := store.NewEvenStoreActorWithPostgres()
	if err != nil {
		return LocaleItemHandler{}, err
	}

	err = actor.RegisterActor(&es)
	if err != nil {
		return LocaleItemHandler{}, err
	}

	return LocaleItemHandler{
		es,
	}, nil
}
