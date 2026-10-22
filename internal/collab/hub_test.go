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
	joined := len(c.send) // the line IDs that every client gets on joining

	h.Replace("hello")
	if got := h.Text(); got != "hello" {
		t.Errorf("Text() = %q, want %q", got, "hello")
	}
	// The update and the line IDs that go with it.
	if got := len(c.send) - joined; got != 2 {
		t.Errorf("client got %d messages, want 2", got)
	}

	h.Replace("hello")
	if len(h.updates) != 1 {
		t.Errorf("replacing with the same text added an update: %d", len(h.updates))
	}
}

// lineIDs decodes a lines message.
func lineIDs(t *testing.T, data []byte) []string {
	t.Helper()
	var msg linesMessage
	if err := json.Unmarshal(data, &msg); err != nil || msg.Type != "lines" {
		t.Fatalf("%s is not a lines message (%v)", data, err)
	}
	return msg.IDs
}

func TestClientsReceiveLineIDsAfterEveryChange(t *testing.T) {
	h := NewHub()
	c := newTestClient()
	h.join(c)
	joined := lineIDs(t, <-c.send)
	if len(joined) != 1 {
		t.Fatalf("on joining: got %d IDs, want 1", len(joined))
	}

	h.Replace("a\nb")
	<-c.send // the update
	changed := lineIDs(t, <-c.send)
	if len(changed) != 2 {
		t.Fatalf("after the change: got %d IDs, want 2", len(changed))
	}
	if changed[0] != joined[0] {
		t.Errorf("the first line lost its ID: %s, was %s", changed[0], joined[0])
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
	saved := lineid.Meta{Lines: []lineid.Line{{ID: "old"}}}

	h.RestoreMeta(saved)
	lines := h.Meta().Lines
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	for _, l := range lines {
		if l.ID == "old" {
			t.Errorf("line ID %q may have been handed out before", l.ID)
		}
	}
}
