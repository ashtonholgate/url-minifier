package domain

import (
	"crypto/sha256"
	"encoding/binary"
	"regexp"
	"strings"
)

const (
	// Base62Chars contains all characters used for base62 encoding
	Base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// ShortCodeLength is the default length of generated short codes
	ShortCodeLength = 7
)

var (
	// aliasPattern defines valid characters for custom aliases
	aliasPattern = regexp.MustCompile("^[0-9A-Za-z-_]+$")
)

// Shortener handles URL shortening operations
type Shortener struct {
	base62Chars []rune
}

// NewShortener creates a new Shortener instance
func NewShortener() *Shortener {
	return &Shortener{
		base62Chars: []rune(Base62Chars),
	}
}

// GenerateShortCode creates a short code from a URL
func (s *Shortener) GenerateShortCode(url string, userID string) string {
	// Create a hash of the URL and userID to ensure uniqueness
	hasher := sha256.New()
	hasher.Write([]byte(url))
	hasher.Write([]byte(userID))
	hash := hasher.Sum(nil)

	// Take first 8 bytes of hash and convert to uint64
	num := binary.BigEndian.Uint64(hash[:8])

	// Convert to base62
	return s.toBase62(num)
}

// ValidateCustomAlias checks if a custom alias is valid
func (s *Shortener) ValidateCustomAlias(alias string) error {
	if len(alias) < 3 || len(alias) > 32 {
		return ErrInvalidAlias
	}

	if !aliasPattern.MatchString(alias) {
		return ErrInvalidAlias
	}

	return nil
}

// toBase62 converts a number to base62 string
func (s *Shortener) toBase62(num uint64) string {
	// Initialize result with ShortCodeLength zeros
	result := make([]rune, ShortCodeLength)
	for i := range result {
		result[i] = s.base62Chars[0]
	}

	// Convert to base62, filling from right to left
	i := ShortCodeLength - 1
	for num > 0 && i >= 0 {
		result[i] = s.base62Chars[num%62]
		num /= 62
		i--
	}

	return string(result)
}

// fromBase62 converts a base62 string back to uint64
func (s *Shortener) fromBase62(str string) uint64 {
	var n uint64
	base := uint64(len(s.base62Chars))
	
	for _, char := range str {
		n *= base
		n += uint64(strings.IndexRune(Base62Chars, char))
	}
	
	return n
}
