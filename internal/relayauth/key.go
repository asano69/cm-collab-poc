// Package relayauth authenticates relay servers to the main server with
// ed25519 signatures. The relay signs each connection request with its private
// key; the main server accepts only public keys listed in authorized_keys.
package relayauth

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

const sshKeyType = "ssh-ed25519"

// sshPrefix is the OpenSSH wire encoding that precedes the raw public key:
// the key type string, then the length of the key.
var sshPrefix = binary.BigEndian.AppendUint32(
	append(binary.BigEndian.AppendUint32(nil, uint32(len(sshKeyType))), sshKeyType...),
	ed25519.PublicKeySize,
)

// GenerateKey creates a new private key at path (PKCS#8 PEM, mode 0600). An
// existing file is never overwritten.
func GenerateKey(path string) (ed25519.PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("encode key: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return nil, fmt.Errorf("%s already exists; remove it first to generate a new key", path)
	}
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	err = pem.Encode(f, &pem.Block{Type: "PRIVATE KEY", Bytes: der})
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(path)
		return nil, fmt.Errorf("write %s: %w", path, err)
	}
	return priv, nil
}

// LoadPrivateKey reads a key written by GenerateKey.
func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s not found; create it with `myapp keygen`", path)
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("%s is not a PEM file", path)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%s is not an ed25519 key", path)
	}
	return priv, nil
}

// PublicKeyLine returns the authorized_keys line for pub.
func PublicKeyLine(pub ed25519.PublicKey) string {
	blob := append(bytes.Clone(sshPrefix), pub...)
	return sshKeyType + " " + base64.StdEncoding.EncodeToString(blob)
}

// AuthorizedKey returns the authorized_keys line for the public half of priv.
func AuthorizedKey(priv ed25519.PrivateKey) string {
	return PublicKeyLine(priv.Public().(ed25519.PublicKey))
}

func parseAuthorizedKey(line string) (ed25519.PublicKey, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 || fields[0] != sshKeyType {
		return nil, fmt.Errorf("want %q followed by a base64 key", sshKeyType)
	}
	blob, err := base64.StdEncoding.DecodeString(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid base64 key: %w", err)
	}
	if len(blob) != len(sshPrefix)+ed25519.PublicKeySize || !bytes.HasPrefix(blob, sshPrefix) {
		return nil, errors.New("not a valid ssh-ed25519 public key")
	}
	return ed25519.PublicKey(blob[len(sshPrefix):]), nil
}

// LoadAuthorized reads the public keys from an authorized_keys file. A missing
// file means that no key is authorized.
func LoadAuthorized(path string) ([]ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var keys []ed25519.PublicKey
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, err := parseAuthorizedKey(line)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, i+1, err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}
