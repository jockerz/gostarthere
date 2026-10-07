package service

import (
	"crypto/rand"
	"math/big"
)

const LowerAlphanumChars = "abcdefghijklmnopqrstuvwxyz0123456789"
const LengthLowerAlphanumChars = len(LowerAlphanumChars)

var LowerAlphanumRunes = []rune(LowerAlphanumChars)

// Generate random alphanumeric string with spesified length
func GenerateRandomLowerAlphanumeric(length int) string {
	result := make([]rune, length)
	for _ = range length {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(LengthLowerAlphanumChars)))
		idx := int(n.Int64())
		result = append(result, LowerAlphanumRunes[idx])
	}
	return string(result)
}
