package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pix303/eventstore-go-v2/pkg/broker"
	"github.com/pix303/eventstore-go-v2/pkg/errors"
	"github.com/pix303/eventstore-go-v2/pkg/store"
	"github.com/pix303/localemgmt-go/api/internal/dto"
	"github.com/pix303/localemgmt-go/domain/pkg/localeitem/aggregate"
	"github.com/pix303/localemgmt-go/domain/pkg/localeitem/events"
)

var (
	ErrorVerifyRequest            = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying request parameters")
	ErrorVerifyAggregateExistence = echo.NewHTTPError(http.StatusBadRequest, "Error on verifying existence of aggregateID")
)

type LocaleItemHandler struct {
	eventStore store.EventStore
}

const projectionTopic = "projection"

func detailHandler(c chan broker.BrokerMessage, store *store.EventStore) {
	for {
		msg := <-c
		fmt.Printf("this is the event msg that i just recived %s\n\n", msg)
		evts, _, err := store.Repository.RetriveByAggregateID(msg.AggregateID)
		if err != nil {
			fmt.Println(fmt.Errorf("error on retrive aggregate events: %v", err))
		}
		fmt.Printf("ho many events: %d\n", len(evts))
		aggregate := aggregate.NewLocaleItemAggregate()
		aggregate.Reduce(evts)
		fmt.Printf("%+v\n", aggregate)
	}
}

func NewLocaleItemHandler() (LocaleItemHandler, error) {
	pms := map[string]store.ProjectionChannelHandler{
		"detail": detailHandler,
	}
	configs := []store.EventStoreConfigurator{
		store.WithInMemoryRepository,
		store.PrepareProjectionHandlersConfig(pms),
	}

	es, err := store.NewEventStore(configs)

	if err != nil {
		return LocaleItemHandler{}, err
	}

	return LocaleItemHandler{
		es,
	}, nil
}

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

	evt := events.NewCreateEvent(payload.Content, payload.Context, payload.Lang, "todo")

	result, err := h.eventStore.Add(evt)

	if err != nil && result == false {
		return err
	}

	return c.JSON(http.StatusOK, evt)
}

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

	_, err = h.eventStore.GetByID(payload.AggregateId)
	if err == errors.NotFoundAggregateID {
		return ErrorVerifyAggregateExistence
	}

	evt := events.NewUpdateEvent(payload.AggregateId, payload.Content, payload.Lang, "todo")

	result, err := h.eventStore.Add(evt)

	if err != nil && result == false {
		return err
	}

	return c.JSON(http.StatusOK, evt)
}
