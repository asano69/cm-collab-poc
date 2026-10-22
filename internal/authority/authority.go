// Package authority signs the lines of the shared document with the key of
// the main server, so that meta.json acts as a detached signature of data.txt.
package authority

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"slices"
	"strings"
	"unicode/utf16"

	"github.com/asano69/myapp/internal/lineid"
	"github.com/asano69/myapp/internal/relayauth"
)

// Authority holds the signing key of the main server.
type Authority struct {
	key ed25519.PrivateKey
}

// Load reads the private key at path. A missing file is created with a new key.
func Load(path string) (*Authority, error) {
	var key ed25519.PrivateKey
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		key, err = relayauth.GenerateKey(path)
	} else {
		key, err = relayauth.LoadPrivateKey(path)
	}
	if err != nil {
		return nil, err
	}
	return &Authority{key: key}, nil
}

// PublicKey returns the public key in authorized_keys form.
func (a *Authority) PublicKey() string { return relayauth.AuthorizedKey(a.key) }

// WriteKnownHosts saves the public key for host in a known_hosts file for
// relays. The file is replaced.
func (a *Authority) WriteKnownHosts(path, host string) error {
	line := host + " " + a.PublicKey() + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// KeyPath is where the main server and the relays publish the public key, so
// that browsers can verify the line signatures.
const KeyPath = "/authority.pub"

// KeyHandler serves publicKey (authorized_keys form) as plain text.
func KeyHandler(publicKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, publicKey)
	})
}

// ReadKnownHosts returns the public key, in authorized_keys form, from a
// known_hosts file written by WriteKnownHosts.
func ReadKnownHosts(path string) (string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("%s not found; copy it from the main server", path)
	}
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	// The line is "host ssh-ed25519 AAAA...".
	fields := strings.Fields(string(data))
	if len(fields) < 3 || fields[1] != "ssh-ed25519" {
		return "", fmt.Errorf("%s is not valid; copy it from the main server", path)
	}
	return fields[1] + " " + fields[2], nil
}

// Sign returns meta where every line that is unsigned, or whose text no longer
// matches its hash, is hashed and signed as an edit by author at time now.
// Other lines keep their signature. meta is not modified.
func (a *Authority) Sign(doc []uint16, meta lineid.Meta, author lineid.Author, now int64) lineid.Meta {
	texts := lineTexts(doc)
	meta.Lines = slices.Clone(meta.Lines)
	for i := range meta.Lines {
		l := &meta.Lines[i]
		hash := hashOf(texts[i])
		if l.Signature != "" && l.Hash == hash {
			continue
		}
		*l = lineid.Line{ID: l.ID, Hash: hash, Updated: now, Admin: author.Admin, Key: author.Key}
		l.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(a.key, message(*l)))
	}
	return meta
}

// message is what gets signed. The prefix keeps the signature from being valid
// for any other purpose.
func message(l lineid.Line) []byte {
	return fmt.Appendf(nil, "myapp-line-v1\n%s\n%s\n%d\n%s\n%s",
		l.ID, l.Hash, l.Updated, l.Admin, l.Key)
}

func hashOf(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

// lineTexts splits doc into its lines. There is always at least one.
func lineTexts(doc []uint16) []string {
	var lines []string
	start := 0
	for i, u := range doc {
		if u == '\n' {
			lines = append(lines, string(utf16.Decode(doc[start:i])))
			start = i + 1
		}
	}
	return append(lines, string(utf16.Decode(doc[start:])))
}
