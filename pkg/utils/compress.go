package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func Compress(s string) string {
	sum := sha256.Sum224([]byte(s))
	return hex.EncodeToString(sum[:])
}
