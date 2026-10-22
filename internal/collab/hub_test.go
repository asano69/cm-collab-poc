package collab

import (
	"encoding/json"
	"testing"
)

func newTestClient() *client {
	return &client{send: make(chan []byte, 8), cancel: func() {}}
}

func TestPushRejectsInvalidChanges(t *testing.T) {
	h := NewHub()
	bad := []Update{{ClientID: "a", Changes: json.RawMessage(`[3]`)}}

	if err := h.push(newTestClient(), pushMessage{Version: 0, Updates: bad}); err == nil {
		t.Fatal("expected an error for changes that do not fit the document")
	}
	if len(h.updates) != 0 {
		t.Errorf("invalid update was recorded: %d updates", len(h.updates))
	}
}

func TestPushAppliesChanges(t *testing.T) {
	h := NewHub()
	good := []Update{{ClientID: "a", Changes: json.RawMessage(`[[0,"hi"]]`)}}

	if err := h.push(newTestClient(), pushMessage{Version: 0, Updates: good}); err != nil {
		t.Fatal(err)
	}
	if got := h.Text(); got != "hi" {
		t.Errorf("Text() = %q, want %q", got, "hi")
	}
}

func TestReplaceUpdatesTextAndNotifiesClients(t *testing.T) {
	h := NewHub()
	c := newTestClient()
	h.join(c)

	h.Replace("hello")
	if got := h.Text(); got != "hello" {
		t.Errorf("Text() = %q, want %q", got, "hello")
	}
	if len(c.send) != 1 {
		t.Errorf("client got %d messages, want 1", len(c.send))
	}

	h.Replace("hello")
	if len(h.updates) != 1 {
		t.Errorf("replacing with the same text added an update: %d", len(h.updates))
	}
}
