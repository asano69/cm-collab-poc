package lineid

import (
	"slices"
	"strconv"
	"testing"
	"unicode/utf16"
)

func units(s string) []uint16 { return utf16.Encode([]rune(s)) }

// edit is an Edit whose text is a string.
type edit struct {
	from, to int
	text     string
}

// useCounterIDs makes the new IDs "1", "2", ... for the rest of the test.
func useCounterIDs(t *testing.T) {
	t.Helper()
	old, n := newID, 0
	newID = func() string {
		n++
		return strconv.Itoa(n)
	}
	t.Cleanup(func() { newID = old })
}

// apply gives the lines of doc the IDs "1", "2", ... (updated at 0) and
// applies the edits at time now.
func apply(t *testing.T, doc string, now int64, edits ...edit) Meta {
	t.Helper()
	useCounterIDs(t)
	var es []Edit
	for _, e := range edits {
		es = append(es, Edit{From: e.from, To: e.to, Text: units(e.text)})
	}
	return New(units(doc), 0).Apply(units(doc), es, now)
}

func ids(m Meta) []string {
	out := make([]string, len(m.Lines))
	for i, l := range m.Lines {
		out[i] = l.ID
	}
	return out
}

func TestApply(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		edit edit
		want []string
	}{
		{"typing", "ab", edit{1, 1, "x"}, []string{"1"}},
		{"enter at line end", "ab", edit{2, 2, "\n"}, []string{"1", "2"}},
		{"enter at line start", "ab", edit{0, 0, "\n"}, []string{"2", "1"}},
		{"enter in the middle", "abcd", edit{2, 2, "\n"}, []string{"1", "2"}},
		{"enter on an empty line", "a\n\nb", edit{2, 2, "\n"}, []string{"1", "2", "4", "3"}},
		{"join lines", "ab\ncd", edit{2, 3, ""}, []string{"1"}},
		{"join with an empty upper line", "\ncd", edit{0, 1, ""}, []string{"2"}},
		{"delete a whole line", "a\nb\nc", edit{2, 4, ""}, []string{"1", "3"}},
		{"delete a whole line before an empty one", "a\nb\n\nc", edit{2, 4, ""}, []string{"1", "3", "4"}},
		{"clear a line", "ab\ncd", edit{0, 2, ""}, []string{"1", "2"}},
		{"delete across lines", "abc\ndef\nghi", edit{1, 9, ""}, []string{"1"}},
		{"paste lines at the end", "ab", edit{2, 2, "\nx\ny"}, []string{"1", "2", "3"}},
		{"paste a line at the start", "ab", edit{0, 0, "x\n"}, []string{"2", "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ids(apply(t, tt.doc, 1, tt.edit)); !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyHandlesSeveralEdits(t *testing.T) {
	got := ids(apply(t, "ab\ncd", 1, edit{1, 1, "\n"}, edit{4, 4, "\n"}))
	if want := []string{"1", "4", "2", "3"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestApplyMarksTouchedLinesOnly(t *testing.T) {
	m := apply(t, "a\nb", 5, edit{0, 0, "x"})
	if m.Lines[0].Updated != 5 || m.Lines[1].Updated != 0 {
		t.Errorf("updated = %d, %d; want 5, 0", m.Lines[0].Updated, m.Lines[1].Updated)
	}
}

func TestApplyKeepsTheOldMeta(t *testing.T) {
	old := New(units("ab"), 0)
	old.Apply(units("ab"), []Edit{{From: 2, To: 2, Text: units("\n")}}, 1)
	if len(old.Lines) != 1 {
		t.Errorf("the old meta was modified: %v", old.Lines)
	}
}
