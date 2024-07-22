package utils

import (
	"encoding/hex"

	"github.com/google/uuid"
)

func IsNotNilOrEmptyString(v *string) bool {
	return v != nil && *v != ""
}

func IsNilOrEmptyString(v *string) bool {
	return v == nil || *v == ""
}

func NewUUID() *string {
	var r [32]byte

	u := uuid.New()
	hex.Encode(r[:], u[:])

	return PtrTo(string(r[:]))
}
