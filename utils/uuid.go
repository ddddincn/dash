package utils

import (
	"encoding/hex"

	"github.com/google/uuid"
)

func GenUUIDWithOutDash() string {
	uuid := uuid.New()
	var buf [32]byte
	dst := buf[:]
	hex.Encode(dst, uuid[:])
	return string(buf[:])
}
