package domain_test

import (
	"strings"
	"testing"

	"github.com/pix303/eventstore-go-v2/pkg/domain"
)

func TestDefaultStoreEvent_ok(t *testing.T) {
	storeEvent := domain.NewDefaultStoreEvent()
	storeEventPrint := storeEvent.ToString()
	if !strings.Contains(storeEventPrint, "no-type") {
		t.Errorf("store event not printed correctly")
	}
}

func TestDefaultStoreEvent_fail(t *testing.T) {
	storeEvent := domain.NewDefaultStoreEvent()
	storeEvent.EventType = "add-something"
	storeEventPrint := storeEvent.ToString()
	if strings.Contains(storeEventPrint, "no-type") {
		t.Errorf("store event not printed correctly")
	}
	if !strings.Contains(storeEventPrint, "add-something") {
		t.Errorf("store event not printed correctly")
	}
}
