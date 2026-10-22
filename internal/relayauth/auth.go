package relayauth

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Path is the endpoint of the main server that relays connect to.
const Path = "/connection/relay"

const (
	headerKey       = "X-Relay-Key"
	headerTime      = "X-Relay-Time"
	headerSignature = "X-Relay-Signature"

	// maxSkew limits how far a request's timestamp may be from the server's
	// clock, which bounds how long a captured request could be replayed.
	maxSkew = 30 * time.Second
)

type keyContext struct{}

// KeyFrom returns the public key that a request passed by Require was signed
// with. ok is false for requests that did not come through Require.
func KeyFrom(ctx context.Context) (key ed25519.PublicKey, ok bool) {
	key, ok = ctx.Value(keyContext{}).(ed25519.PublicKey)
	return key, ok
}

// message is what gets signed. The prefix keeps the signature from being valid
// for any other purpose.
func message(timestamp, path string) []byte {
	return []byte("myapp-relay-v1\n" + timestamp + "\n" + path)
}

// Sign adds the authentication headers for a request to path.
func Sign(h http.Header, priv ed25519.PrivateKey, path string, now time.Time) {
	ts := strconv.FormatInt(now.Unix(), 10)
	pub := priv.Public().(ed25519.PublicKey)
	h.Set(headerKey, base64.StdEncoding.EncodeToString(pub))
	h.Set(headerTime, ts)
	h.Set(headerSignature, base64.StdEncoding.EncodeToString(ed25519.Sign(priv, message(ts, path))))
}

func authorized(keys []ed25519.PublicKey, pub ed25519.PublicKey) bool {
	for _, k := range keys {
		if bytes.Equal(k, pub) {
			return true
		}
	}
	return false
}

// verify checks the headers set by Sign. The messages are shown to the user.
func verify(r *http.Request, keys []ed25519.PublicKey, now time.Time) error {
	pub, err := base64.StdEncoding.DecodeString(r.Header.Get(headerKey))
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return errors.New("missing or malformed " + headerKey + " header")
	}
	ts := r.Header.Get(headerTime)
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return errors.New("missing or malformed " + headerTime + " header")
	}
	if skew := now.Sub(time.Unix(sec, 0)); skew > maxSkew || skew < -maxSkew {
		return fmt.Errorf("relay clock differs from the server by %s (allowed: %s)",
			skew.Round(time.Second), maxSkew)
	}
	if !authorized(keys, pub) {
		return errors.New("public key is not in authorized_keys")
	}
	sig, err := base64.StdEncoding.DecodeString(r.Header.Get(headerSignature))
	if err != nil || !ed25519.Verify(pub, message(ts, r.URL.Path), sig) {
		return errors.New("invalid signature")
	}
	return nil
}

// Require lets a request through only when it is signed by a key listed in
// keysPath. The file is read on every request, so edits apply without a
// restart. Connections that are already open are not affected by them.
func Require(keysPath string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys, err := LoadAuthorized(keysPath)
		if err != nil {
			log.Printf("relayauth: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := verify(r, keys, time.Now()); err != nil {
			log.Printf("relayauth: rejected %s: %v", r.RemoteAddr, err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		// verify has checked the header, so it decodes.
		pub, _ := base64.StdEncoding.DecodeString(r.Header.Get(headerKey))
		ctx := context.WithValue(r.Context(), keyContext{}, ed25519.PublicKey(pub))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
