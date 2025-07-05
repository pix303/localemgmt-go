package aggregate

import (
	"log/slog"

	"github.com/pix303/actor-lib/pkg/actor"

	"github.com/pix303/eventstore-go-v2/pkg/store"
)

var (
	FailToRetriveAggregateEvents = "fail to retrive aggregate events "
)

type LocaleItemAggregateState struct {
	aggregate LocaleItemAggregate
	store     store.EventStore
}

type LocaleItemAggregateDetailEvent struct {
	AggregateID string
	EventType   string
	Payload     string
}

func (this *LocaleItemAggregateState) Process(inbox <-chan actor.Message) {
	for {
		msg := <-inbox
		switch payload := msg.Body.(type) {
		case LocaleItemAggregateDetailEvent:
			this.detailHandler(payload)
		}
	}
}

func (this *LocaleItemAggregateState) detailHandler(msg LocaleItemAggregateDetailEvent) {
	evts, _, err := this.store.Repository.RetriveByAggregateID(msg.AggregateID)
	if err != nil {
		slog.Warn(FailToRetriveAggregateEvents, slog.String("error", err.Error()))
		return
	}
	aggregate := NewLocaleItemAggregate()
	aggregate.Reduce(evts)
	slog.Info("aggregate detail created", slog.Any("item", aggregate))
	// TODO: storage file
}
