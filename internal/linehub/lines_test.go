package linehub

import (
	"encoding/json"
	"testing"

	"github.com/asano69/myapp/internal/lineid"
)

func TestEncodeLinesCarriesTheSignature(t *testing.T) {
	meta := lineid.Meta{Lines: []lineid.Line{
		{ID: "x", Updated: 7, Admin: "Bob", Key: "k", Signature: "s"},
	}}

	var msg linesMessage
	if err := json.Unmarshal(encodeLines(meta), &msg); err != nil {
		t.Fatal(err)
	}
	want := lineInfo{ID: "x", Updated: 7, Admin: "Bob", Key: "k", Signature: "s"}
	if len(msg.Lines) != 1 || msg.Lines[0] != want {
		t.Errorf("got %+v, want [%+v]", msg.Lines, want)
	}
}
