package schema

import "github.com/google/uuid"

// newUUIDv7 generates a time-ordered UUID v7, which is index-friendly
// due to its monotonically increasing timestamp prefix.
func newUUIDv7() uuid.UUID {
	id, _ := uuid.NewV7()
	return id
}
