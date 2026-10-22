package linehub

import (
	"encoding/json"
	"testing"

	"github.com/asano69/myapp/internal/lineid"
)

// lineIDs decodes a lines message.
func lineIDs(t *testing.T, data []byte) []string {
	t.Helper()
	var msg linesMessage
	if err := json.Unmarshal(data, &msg); err != nil || msg.Type != "lines" {
		t.Fatalf("%s is not a lines message (%v)", data, err)
	}
	ids := make([]string, len(msg.Lines))
	for i, l := range msg.Lines {
		ids[i] = l.ID
	}
	return ids
}

func TestMessageHoldsTheLineIDs(t *testing.T) {
	h := New()
	joined := lineIDs(t, h.Message())
	if len(joined) != 1 {
		t.Fatalf("empty document: got %d IDs, want 1", len(joined))
	}

	h.Replace("a\nb")
	changed := lineIDs(t, h.Message())
	if len(changed) != 2 {
		t.Fatalf("after the change: got %d IDs, want 2", len(changed))
	}
	if changed[0] != joined[0] {
		t.Errorf("the first line lost its ID: %s, was %s", changed[0], joined[0])
	}
}

func TestChangesKeepLineIDsInStepWithTheText(t *testing.T) {
	h := New()
	h.Replace("a\nb\nc")

	if got := len(h.Meta().Lines); got != 3 {
		t.Errorf("got %d line IDs, want 3", got)
	}
}

func TestRestoreMetaNeverReusesIDs(t *testing.T) {
	h := New()
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
