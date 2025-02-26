package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ashtonholgate/url-minifier/services/shortener/internal/domain"
	"github.com/ashtonholgate/url-minifier/services/shortener/internal/repository"
)

// CreateURLRequest represents the request to create a new short URL
type CreateURLRequest struct {
	LongURL     string
	UserID      string
	CustomAlias string
	ExpiresIn   *time.Duration
}

// URLService handles the business logic for URL operations
type URLService interface {
	// CreateURL creates a new short URL
	CreateURL(ctx context.Context, req CreateURLRequest) (*domain.URL, error)

	// GetURL retrieves a URL by its short code
	GetURL(ctx context.Context, code string) (*domain.URL, error)

	// DeleteURL removes a URL by its ID
	DeleteURL(ctx context.Context, id string, userID string) error

	// ListUserURLs retrieves all URLs for a given user
	ListUserURLs(ctx context.Context, userID string) ([]*domain.URL, error)

	// Close cleans up any resources
	Close(ctx context.Context) error
}

// urlService implements URLService
type urlService struct {
	repo     repository.Repository
	shortner *domain.Shortener
}

// NewURLService creates a new URLService instance
func NewURLService(repo repository.Repository) URLService {
	return &urlService{
		repo:     repo,
		shortner: domain.NewShortener(),
	}
}

// CreateURL implements URLService.CreateURL
func (s *urlService) CreateURL(ctx context.Context, req CreateURLRequest) (*domain.URL, error) {
	// Validate URL
	if err := domain.ValidateURL(req.LongURL); err != nil {
		return nil, err
	}

	// Set default expiration if not provided
	expiresAt := time.Now().Add(24 * time.Hour)
	if req.ExpiresIn != nil {
		expiresAt = time.Now().Add(*req.ExpiresIn)
	}

	var shortCode string
	var err error

	// Handle custom alias if provided
	if req.CustomAlias != "" {
		if err = s.shortner.ValidateCustomAlias(req.CustomAlias); err != nil {
			return nil, err
		}

		// Check if alias is available
		available, err := s.repo.IsCodeAvailable(ctx, req.CustomAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to check alias availability: %w", err)
		}
		if !available {
			return nil, domain.ErrCodeExists
		}
		shortCode = req.CustomAlias
	} else {
		// Generate short code and ensure it's unique
		for i := 0; i < 3; i++ { // Try up to 3 times
			shortCode = s.shortner.GenerateShortCode(req.LongURL, req.UserID)
			available, err := s.repo.IsCodeAvailable(ctx, shortCode)
			if err != nil {
				return nil, fmt.Errorf("failed to check code availability: %w", err)
			}
			if available {
				break
			}
			if i == 2 {
				return nil, fmt.Errorf("failed to generate unique short code after %d attempts", i+1)
			}
		}
	}

	// Create URL entity
	url := &domain.URL{
		ID:          fmt.Sprintf("url_%d", time.Now().UnixNano()),
		LongURL:     req.LongURL,
		ShortCode:   shortCode,
		UserID:      req.UserID,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		CustomAlias: req.CustomAlias,
	}

	// Store URL
	if err := s.repo.StoreURL(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to store URL: %w", err)
	}

	return url, nil
}

// GetURL implements URLService.GetURL
func (s *urlService) GetURL(ctx context.Context, code string) (*domain.URL, error) {
	url, err := s.repo.GetURLByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Check if URL has expired
	if time.Now().After(url.ExpiresAt) {
		// Delete expired URL
		if err := s.repo.DeleteURL(ctx, url.ID); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to delete expired URL: %v\n", err)
		}
		return nil, domain.ErrExpiredURL
	}

	return url, nil
}

// DeleteURL implements URLService.DeleteURL
func (s *urlService) DeleteURL(ctx context.Context, id string, userID string) error {
	// Get URL first to check ownership
	url, err := s.repo.GetURLByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify ownership
	if url.UserID != userID {
		return fmt.Errorf("unauthorized: URL belongs to different user")
	}

	return s.repo.DeleteURL(ctx, id)
}

// ListUserURLs implements URLService.ListUserURLs
func (s *urlService) ListUserURLs(ctx context.Context, userID string) ([]*domain.URL, error) {
	urls, err := s.repo.ListURLsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list URLs: %w", err)
	}

	// Filter out expired URLs
	active := make([]*domain.URL, 0, len(urls))
	for _, url := range urls {
		if time.Now().Before(url.ExpiresAt) {
			active = append(active, url)
		}
	}

	return active, nil
}

// Close implements URLService.Close
func (s *urlService) Close(ctx context.Context) error {
	return s.repo.Close(ctx)
}
