package events

import (
	"github.com/pix303/eventstore-go-v2/pkg/events"
)

const LocaleItemAggregateName = "localeitem"

const CreateLocaleItemStoreEventType = "created-localeitem"

type CreateLocaleItemPayload struct {
	Content string
	Context string
	Lang    string
}

func NewCreateEvent(content string, context string, lang string, userID string) events.StoreEvent[CreateLocaleItemPayload] {

	payload := CreateLocaleItemPayload{
		content,
		context,
		lang,
	}

	return events.NewStoreEvent(CreateLocaleItemStoreEventType, LocaleItemAggregateName, userID, payload)
}

const UpdateTranslationStoreEventType = "update-translation"

type UpdateTranslationLocaleItemPayload struct {
	Content string
	Lang    string
}

func NewUpdateEvent(aggregateID string, content string, lang string, userID string) events.StoreEvent[UpdateTranslationLocaleItemPayload] {

	payload := UpdateTranslationLocaleItemPayload{
		content,
		lang,
	}

	se := events.NewStoreEvent(UpdateTranslationStoreEventType, LocaleItemAggregateName, userID, payload)
	se.AggregateID = aggregateID
	return se
}
