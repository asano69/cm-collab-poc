package lineid

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// newID returns the ID for a new line. Tests replace it to get predictable IDs.
var newID = newUUIDv7

// newUUIDv7 returns a random UUID version 7 (RFC 9562): a 48-bit Unix time in
// milliseconds, the version, the variant and 74 random bits.
func newUUIDv7() string {
	var u [16]byte
	// Reading random bytes cannot fail: the runtime aborts if it would.
	rand.Read(u[:])

	ms := uint64(time.Now().UnixMilli())
	for i := 0; i < 6; i++ {
		u[i] = byte(ms >> (8 * (5 - i)))
	}
	u[6] = 0x70 | u[6]&0x0f // version 7
	u[8] = 0x80 | u[8]&0x3f // variant 10

	s := hex.EncodeToString(u[:])
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}
