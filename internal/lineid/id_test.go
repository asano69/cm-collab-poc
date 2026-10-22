package lineid

import (
	"regexp"
	"testing"
)

var uuidV7 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewUUIDv7(t *testing.T) {
	a, b := newUUIDv7(), newUUIDv7()
	for _, id := range []string{a, b} {
		if !uuidV7.MatchString(id) {
			t.Errorf("%q is not a UUID version 7", id)
		}
	}
	if a == b {
		t.Errorf("two new IDs are equal: %q", a)
	}
}
