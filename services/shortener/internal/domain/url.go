package domain

import (
	"errors"
	"net/url"
	"time"
)

var (
	ErrInvalidURL     = errors.New("invalid URL format")
	ErrURLNotFound    = errors.New("URL not found")
	ErrCodeExists     = errors.New("short code already exists")
	ErrExpiredURL     = errors.New("URL has expired")
	ErrInvalidAlias   = errors.New("invalid custom alias format")
)

// URL represents a shortened URL
type URL struct {
	ID          string    `json:"id"`
	LongURL     string    `json:"long_url"`
	ShortCode   string    `json:"short_code"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	CustomAlias string    `json:"custom_alias,omitempty"`
}

// ValidateURL checks if the provided URL is valid
func ValidateURL(longURL string) error {
	parsedURL, err := url.Parse(longURL)
	if err != nil {
		return ErrInvalidURL
	}
	
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidURL
	}
	
	if parsedURL.Host == "" {
		return ErrInvalidURL
	}
	
	return nil
}
