package collab

import (
	"encoding/json"
	"testing"

	"github.com/asano69/myapp/internal/lineid"
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

func TestPushKeepsLineIDsInStepWithTheText(t *testing.T) {
	h := NewHub()
	good := []Update{{ClientID: "a", Changes: json.RawMessage(`[[0,"a","b","c"]]`)}}

	if err := h.push(newTestClient(), pushMessage{Version: 0, Updates: good}); err != nil {
		t.Fatal(err)
	}
	if got := len(h.Meta().Lines); got != 3 {
		t.Errorf("got %d line IDs, want 3", got)
	}
}

func TestRestoreMetaNeverReusesIDs(t *testing.T) {
	h := NewHub()
	h.Replace("a\nb")
	// One saved line, but the document has two.
	saved := lineid.Meta{NextID: 50, Lines: []lineid.Line{{ID: 7}}}

	h.RestoreMeta(saved)
	for _, l := range h.Meta().Lines {
		if l.ID < 50 {
			t.Errorf("line ID %d may have been handed out before", l.ID)
		}
	}
}
