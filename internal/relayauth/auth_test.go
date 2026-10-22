package relayauth

import (
	"crypto/ed25519"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	priv, err := GenerateKey(filepath.Join(t.TempDir(), "key.pem"))
	if err != nil {
		t.Fatal(err)
	}
	return priv
}

func pubOf(priv ed25519.PrivateKey) ed25519.PublicKey {
	return priv.Public().(ed25519.PublicKey)
}

func signedRequest(priv ed25519.PrivateKey, now time.Time) *http.Request {
	r := httptest.NewRequest(http.MethodGet, Path, nil)
	Sign(r.Header, priv, Path, now)
	return r
}

func TestVerify(t *testing.T) {
	priv, other := newKey(t), newKey(t)
	now := time.Now()
	keys := []ed25519.PublicKey{pubOf(priv)}

	tests := []struct {
		name    string
		req     func() *http.Request
		keys    []ed25519.PublicKey
		wantErr string
	}{
		{"valid", func() *http.Request { return signedRequest(priv, now) }, keys, ""},
		{"expired", func() *http.Request { return signedRequest(priv, now.Add(-time.Minute)) }, keys, "clock differs"},
		{"future", func() *http.Request { return signedRequest(priv, now.Add(time.Minute)) }, keys, "clock differs"},
		{"unknown key", func() *http.Request { return signedRequest(other, now) }, keys, "not in authorized_keys"},
		{"no keys", func() *http.Request { return signedRequest(priv, now) }, nil, "not in authorized_keys"},
		{"other path", func() *http.Request {
			r := signedRequest(priv, now)
			r.URL.Path = "/connection/collab"
			return r
		}, keys, "invalid signature"},
		{"forged signature", func() *http.Request {
			r := signedRequest(priv, now)
			r.Header.Set(headerSignature, signedRequest(other, now).Header.Get(headerSignature))
			return r
		}, keys, "invalid signature"},
		{"no headers", func() *http.Request { return httptest.NewRequest(http.MethodGet, Path, nil) }, keys, "malformed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verify(tt.req(), tt.keys, now)
			switch {
			case tt.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadAuthorized(t *testing.T) {
	priv := newKey(t)
	path := filepath.Join(t.TempDir(), "authorized_keys")
	content := "# relays\n\n" + AuthorizedKey(priv) + " laptop\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	keys, err := LoadAuthorized(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || !keys[0].Equal(pubOf(priv)) {
		t.Errorf("keys = %v, want just the generated key", keys)
	}
}

func TestLoadAuthorizedMissingFile(t *testing.T) {
	keys, err := LoadAuthorized(filepath.Join(t.TempDir(), "none"))
	if err != nil || len(keys) != 0 {
		t.Errorf("got %v, %v; want no keys and no error", keys, err)
	}
}

func TestLoadAuthorizedRejectsBadLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "authorized_keys")
	if err := os.WriteFile(path, []byte("ssh-rsa AAAA\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAuthorized(path); err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Errorf("error = %v, want one naming line 1", err)
	}
}

func TestGenerateKeyKeepsExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "key.pem")
	if _, err := GenerateKey(path); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateKey(path); err == nil {
		t.Fatal("second GenerateKey overwrote the file")
	}
	if _, err := LoadPrivateKey(path); err != nil {
		t.Errorf("the generated key cannot be loaded: %v", err)
	}
}
