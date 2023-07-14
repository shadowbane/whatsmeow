package helpers

import (
	cryptorand "crypto/rand"
	"github.com/oklog/ulid/v2"
)

func NewULID() string {
	reader, _ := cryptorand.Read(make([]byte, 128))

	return ulid.MustNew(uint64(reader), cryptorand.Reader).String()
}
