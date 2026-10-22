package authority

import (
	"crypto/ed25519"
	"encoding/base64"
	"path/filepath"
	"testing"
	"unicode/utf16"

	"github.com/asano69/myapp/internal/lineid"
)

func units(s string) []uint16 { return utf16.Encode([]rune(s)) }

func newAuthority(t *testing.T) *Authority {
	t.Helper()
	a, err := Load(filepath.Join(t.TempDir(), "Authority.pem"))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestSignRecordsAuthorAndVerifies(t *testing.T) {
	a := newAuthority(t)
	doc := units("a\nb")
	author := lineid.Author{Admin: "Bob", Key: "ssh-ed25519 AAAA"}

	meta := a.Sign(doc, lineid.New(doc, 1), author, 5)
	for _, l := range meta.Lines {
		if l.Admin != "Bob" || l.Key != author.Key || l.Updated != 5 || l.Hash == "" {
			t.Fatalf("line not filled in: %+v", l)
		}
		sig, err := base64.StdEncoding.DecodeString(l.Signature)
		if err != nil {
			t.Fatal(err)
		}
		pub := a.key.Public().(ed25519.PublicKey)
		if !ed25519.Verify(pub, message(l), sig) {
			t.Errorf("signature of %q does not verify", l.ID)
		}
	}
}

func TestSignKeepsUnchangedLines(t *testing.T) {
	a := newAuthority(t)
	doc := units("a\nb")
	first := a.Sign(doc, lineid.New(doc, 1), lineid.Author{Admin: "Bob"}, 5)

	second := a.Sign(units("a\nc"), first, lineid.Author{Admin: "Alice"}, 9)
	if second.Lines[0] != first.Lines[0] {
		t.Errorf("unchanged line was signed again: %+v", second.Lines[0])
	}
	if l := second.Lines[1]; l.Admin != "Alice" || l.Updated != 9 || l.Hash == first.Lines[1].Hash {
		t.Errorf("changed line was not signed again: %+v", l)
	}
}
