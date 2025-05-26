package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random ID
func GenerateID(prefix string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return prefix + "-" + hex.EncodeToString(bytes)
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}