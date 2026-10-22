package collab

import (
	"encoding/json"
	"testing"

	"github.com/asano69/myapp/internal/lineid"
)

func TestEncodeLinesCarriesTimeAndAdmin(t *testing.T) {
	meta := lineid.Meta{Lines: []lineid.Line{{ID: "x", Updated: 7, Admin: "Bob"}}}

	var msg linesMessage
	if err := json.Unmarshal(encodeLines(meta), &msg); err != nil {
		t.Fatal(err)
	}
	want := lineInfo{ID: "x", Updated: 7, Admin: "Bob"}
	if len(msg.Lines) != 1 || msg.Lines[0] != want {
		t.Errorf("got %+v, want [%+v]", msg.Lines, want)
	}
}
