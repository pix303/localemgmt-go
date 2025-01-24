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

func NewTranslationItem(lang, content, userId string) TranslationItem {
	return TranslationItem{
		Lang:      lang,
		Content:   content,
		CreatedBy: userId,
		UpdatedBy: userId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (item *TranslationItem) UpdateTranslationItem(content, userId string) {
	item.Content = content
	item.UpdatedBy = userId
	item.UpdatedAt = time.Now()
}

type LocaleItemAggregate struct {
	AggregateID  string
	Context      string
	Translations []TranslationItem
}

func NewLocaleItemAggregate() LocaleItemAggregate {
	return LocaleItemAggregate{
		// todo: how to better init stirngs that will contains some valuable value
		"no-code",
		"no-context",
		// todo: ho to better init an array?
		make([]TranslationItem, 5),
	}
}

func (a *LocaleItemAggregate) Reduce(evts []events.AggregateEvent) {
	for _, evt := range evts {
		a.Apply(evt)
	}
}

func (a *LocaleItemAggregate) Apply(evt events.AggregateEvent) {
	switch evt.GetEventType() {
	case domain.CreateLocaleItemStoreEventType:
		a.init(evt.(events.StoreEvent[domain.CreateLocaleItemPayload]))
	case domain.UpdateTranslationStoreEventType:
		a.update(evt.(events.StoreEvent[domain.UpdateTranslationLocaleItemPayload]))
	}
}

func (item *LocaleItemAggregate) init(evt events.StoreEvent[domain.CreateLocaleItemPayload]) {
	item.AggregateID = evt.GetAggregateID()
	item.Context = evt.GetAggregateID()
	item.Context = evt.PayloadData.Context
	item.Translations = append(item.Translations, NewTranslationItem(evt.PayloadData.Lang, evt.PayloadData.Content, "todo"))
}

func (item *LocaleItemAggregate) update(evt events.StoreEvent[domain.UpdateTranslationLocaleItemPayload]) {
	newTranslation := NewTranslationItem(evt.PayloadData.Lang, evt.PayloadData.Content, "todo")
	updateOrInsert(item.Translations, newTranslation)
}

func updateOrInsert(items []TranslationItem, newItem TranslationItem) []TranslationItem {
	for idx, titem := range items {
		if titem.Lang == newItem.Lang {
			items[idx] = newItem
			return items
		}
	}
	return append(items, newItem)
}
