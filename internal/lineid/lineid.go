// Package lineid keeps a stable ID for every line of a document. The IDs are
// updated while an edit is applied, so they always match the text.
//
// An edit replaces the lines it touches by the lines it produces. One of the
// touched lines keeps its ID on one of the produced lines; every other
// produced line gets a fresh ID, and the other touched IDs are gone for good.
// IDs are random UUIDs, so a gone ID is never handed out again.
//
//   - Normally the first touched line keeps its ID on the first produced line:
//     Enter at the end of a line, joining lines, typing.
//   - If the edit starts at a line start and leaves text after it (or touches
//     several lines), the last touched line keeps its ID on the last produced
//     line: Enter at the start of a line, deleting whole lines.
//
// Undo and moving text are ordinary edits, so they produce fresh IDs.
package lineid

import (
	"slices"

	"github.com/google/uuid"
)

const newline = '\n'

// newID returns the ID for a new line, a UUID version 7. Tests replace it to
// get predictable IDs.
var newID = func() string { return uuid.Must(uuid.NewV7()).String() }

// Line is the metadata of one line. Apply leaves the lines it produces
// unsigned; the signer fills in everything but ID and Updated.
type Line struct {
	ID        string `json:"id"`        // UUID version 7
	Hash      string `json:"hash"`      // SHA-256 of the line text, hex
	Updated   int64  `json:"updated"`   // Unix time of the last edit that touched the line
	Admin     string `json:"admin"`     // name of the administrator of the server the edit came through
	Key       string `json:"key"`       // public key of the client that made the edit; empty without one
	Signature string `json:"signature"` // of ID, Hash, Updated, Admin and Key, by the main server
}

// Author says who is behind an edit.
type Author struct {
	Admin string
	Key   string
}

// Meta holds the metadata of every line, in document order.
type Meta struct {
	Lines []Line `json:"lines"`
}

// Edit replaces doc[From:To] with Text. Positions are UTF-16 code units.
type Edit struct {
	From, To int
	Text     []uint16
}

// Count returns the number of lines of doc.
func Count(doc []uint16) int { return newlines(doc) + 1 }

func newlines(units []uint16) int {
	n := 0
	for _, u := range units {
		if u == newline {
			n++
		}
	}
	return n
}

// New gives every line of doc a fresh ID.
func New(doc []uint16, now int64) Meta {
	lines := make([]Line, Count(doc))
	for i := range lines {
		lines[i] = Line{ID: newID(), Updated: now}
	}
	return Meta{Lines: lines}
}

// Apply returns the metadata after edits were applied to doc. The edits must
// be sorted and must not overlap, with positions in doc. m is not modified.
func (m Meta) Apply(doc []uint16, edits []Edit, now int64) Meta {
	m.Lines = slices.Clone(m.Lines)
	// Last edit first, so that the line indexes of the earlier ones stay valid.
	for n := len(edits) - 1; n >= 0; n-- {
		m.applyEdit(doc, edits[n], now)
	}
	return m
}

func (m *Meta) applyEdit(doc []uint16, e Edit, now int64) {
	first := newlines(doc[:e.From])
	last := first + newlines(doc[e.From:e.To])
	atLineStart := e.From == 0 || doc[e.From-1] == newline
	atLineEnd := e.To == len(doc) || doc[e.To] == newline
	produced := newlines(e.Text) + 1

	// The ID that stays, and the produced line that keeps it.
	survivor, slot := m.Lines[first].ID, 0
	if atLineStart && (first < last || !atLineEnd) {
		survivor, slot = m.Lines[last].ID, produced-1
	}

	lines := make([]Line, produced)
	for n := range lines {
		id := survivor
		if n != slot {
			id = newID()
		}
		lines[n] = Line{ID: id, Updated: now}
	}
	m.Lines = slices.Replace(m.Lines, first, last+1, lines...)
}
