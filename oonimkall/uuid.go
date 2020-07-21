package oonimkall

import "github.com/google/uuid"

// NewUUID4 generates a new UUID4.
func NewUUID4() string {
	return uuid.Must(uuid.NewRandom()).String()
}
