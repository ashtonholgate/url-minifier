package repository

import (
	"context"
	"time"

	"github.com/ashtonholgate/url-minifier/services/shortener/internal/domain"
)

// Repository defines the interface for URL storage operations
type Repository interface {
	// StoreURL saves a URL to the repository
	StoreURL(ctx context.Context, url *domain.URL) error

	// GetURLByCode retrieves a URL by its short code
	GetURLByCode(ctx context.Context, code string) (*domain.URL, error)

	// GetURLByID retrieves a URL by its ID
	GetURLByID(ctx context.Context, id string) (*domain.URL, error)

	// DeleteURL removes a URL from the repository
	DeleteURL(ctx context.Context, id string) error

	// IsCodeAvailable checks if a short code is available for use
	IsCodeAvailable(ctx context.Context, code string) (bool, error)

	// ListURLsByUser retrieves all URLs for a given user
	ListURLsByUser(ctx context.Context, userID string) ([]*domain.URL, error)

	// Close cleans up any resources used by the repository
	Close(ctx context.Context) error
}
