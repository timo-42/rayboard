package events

import (
	"context"
	"errors"
	"testing"
)

func TestPublishSpecificAndWildcardHandlers(t *testing.T) {
	bus := NewBus()
	var seen []string

	bus.Subscribe("ticket.created", func(_ context.Context, event Event) error {
		seen = append(seen, "specific:"+event.Type)
		return nil
	})
	bus.Subscribe("*", func(_ context.Context, event Event) error {
		seen = append(seen, "wildcard:"+event.Type)
		return nil
	})

	errs := bus.Publish(context.Background(), Event{Type: "ticket.created"})

	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if len(seen) != 2 || seen[0] != "specific:ticket.created" || seen[1] != "wildcard:ticket.created" {
		t.Fatalf("unexpected handlers: %#v", seen)
	}
}

func TestPublishCollectsErrors(t *testing.T) {
	bus := NewBus()
	want := errors.New("failed")

	bus.Subscribe("ticket.updated", func(context.Context, Event) error {
		return want
	})

	errs := bus.Publish(context.Background(), Event{Type: "ticket.updated"})

	if len(errs) != 1 || !errors.Is(errs[0], want) {
		t.Fatalf("unexpected errors: %v", errs)
	}
}
