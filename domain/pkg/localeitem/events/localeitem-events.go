package events

import (
	"github.com/pix303/eventstore-go-v2/pkg/events"
)

const LocaleItemAggregateName = "localeitem"

const CreateLocaleItemStoreEventType = "created-localeitem"

type CreateLocaleItemPayload struct {
	events.StoreEvent
	Content string
	Context string
	Lang    string
}

func NewCreateEvent(content string, context string, lang string, userID string) (events.StoreEvent, error) {
	payload := CreateLocaleItemPayload{
		Content: content,
		Context: context,
		Lang:    lang,
	}

	evt, err := events.NewStoreEvent(CreateLocaleItemStoreEventType, LocaleItemAggregateName, userID, payload, nil)
	if err != nil {
		return evt, err
	}

	return evt, nil
}

const UpdateTranslationStoreEventType = "update-translation"

type UpdateTranslationLocaleItemPayload struct {
	Content string
	Lang    string
}

func NewUpdateEvent(aggregateID string, content string, lang string, userID string) (events.StoreEvent, error) {
	payload := UpdateTranslationLocaleItemPayload{
		content,
		lang,
	}

	evt, err := events.NewStoreEvent(UpdateTranslationStoreEventType, LocaleItemAggregateName, userID, payload, &aggregateID)
	return evt, err
}
