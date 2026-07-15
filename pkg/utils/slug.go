package utils

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// GenerateSlug creates a URL-friendly slug from a string.
// It also appends a random 6-character hex to ensure uniqueness.
func GenerateSlug(title string) string {
	// Lowercase
	slug := strings.ToLower(title)
	// Replace non-alphanumeric characters with hyphens
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "-")
	// Trim trailing/leading hyphens
	slug = strings.Trim(slug, "-")
	
	// Generate random suffix (10 characters)
	b := make([]byte, 5)
	rand.Read(b)
	suffix := hex.EncodeToString(b)
	
	if slug == "" {
		return suffix
	}
	
	return slug + "-" + suffix
}
