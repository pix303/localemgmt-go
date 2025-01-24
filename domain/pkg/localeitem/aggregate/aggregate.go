package aggregate

import (
	"time"

	"github.com/pix303/eventstore-go-v2/pkg/events"
	domain "github.com/pix303/localemgmt-go/domain/pkg/localeitem/events"
)

type TranslationItem struct {
	Lang      string
	Content   string
	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}

type LocaleItemAggregate struct {
	AggregateID string
	Context     string
	Translation []TranslationItem
}

func (a *LocaleItemAggregate) Apply(evt events.AggregateEvent) {
	switch evt.GetEventType() {
	case domain.CreateLocaleItemStoreEventType:
		// initLocaleItem(evt)
	case domain.UpdateTranslationStoreEventType:
		// updateLocaleItem(evt)
	}

}
