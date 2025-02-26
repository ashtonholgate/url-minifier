package service

import (
	"context"
	"testing"
	"time"

	"github.com/ashtonholgate/url-minifier/services/shortener/internal/domain"
)

// mockRepository is a mock implementation of repository.Repository for testing
type mockRepository struct {
	urls map[string]*domain.URL // Map of ID to URL
	codes map[string]*domain.URL // Map of short code to URL
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		urls:  make(map[string]*domain.URL),
		codes: make(map[string]*domain.URL),
	}
}

func (m *mockRepository) StoreURL(ctx context.Context, url *domain.URL) error {
	m.urls[url.ID] = url
	m.codes[url.ShortCode] = url
	return nil
}

func (m *mockRepository) GetURLByCode(ctx context.Context, code string) (*domain.URL, error) {
	if url, ok := m.codes[code]; ok {
		return url, nil
	}
	return nil, domain.ErrURLNotFound
}

func (m *mockRepository) GetURLByID(ctx context.Context, id string) (*domain.URL, error) {
	if url, ok := m.urls[id]; ok {
		return url, nil
	}
	return nil, domain.ErrURLNotFound
}

func (m *mockRepository) DeleteURL(ctx context.Context, id string) error {
	url, ok := m.urls[id]
	if !ok {
		return domain.ErrURLNotFound
	}
	delete(m.urls, id)
	delete(m.codes, url.ShortCode)
	return nil
}

func (m *mockRepository) IsCodeAvailable(ctx context.Context, code string) (bool, error) {
	_, exists := m.codes[code]
	return !exists, nil
}

func (m *mockRepository) ListURLsByUser(ctx context.Context, userID string) ([]*domain.URL, error) {
	var urls []*domain.URL
	for _, url := range m.urls {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}
	return urls, nil
}

func (m *mockRepository) Close(ctx context.Context) error {
	return nil
}

func TestURLService_CreateURL(t *testing.T) {
	repo := newMockRepository()
	service := NewURLService(repo)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     CreateURLRequest
		wantErr bool
	}{
		{
			name: "Valid URL",
			req: CreateURLRequest{
				LongURL: "https://example.com",
				UserID:  "user1",
			},
			wantErr: false,
		},
		{
			name: "Invalid URL",
			req: CreateURLRequest{
				LongURL: "not-a-url",
				UserID:  "user1",
			},
			wantErr: true,
		},
		{
			name: "Valid Custom Alias",
			req: CreateURLRequest{
				LongURL:     "https://example.com",
				UserID:      "user1",
				CustomAlias: "my-link",
			},
			wantErr: false,
		},
		{
			name: "Invalid Custom Alias",
			req: CreateURLRequest{
				LongURL:     "https://example.com",
				UserID:      "user1",
				CustomAlias: "a", // Too short
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := service.CreateURL(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if url == nil {
					t.Error("CreateURL() returned nil URL")
				}
				if url.LongURL != tt.req.LongURL {
					t.Errorf("CreateURL() LongURL = %v, want %v", url.LongURL, tt.req.LongURL)
				}
				if tt.req.CustomAlias != "" && url.ShortCode != tt.req.CustomAlias {
					t.Errorf("CreateURL() ShortCode = %v, want %v", url.ShortCode, tt.req.CustomAlias)
				}
			}
		})
	}
}

func TestURLService_GetURL(t *testing.T) {
	repo := newMockRepository()
	service := NewURLService(repo)
	ctx := context.Background()

	// Create a test URL
	url := &domain.URL{
		ID:        "test1",
		LongURL:   "https://example.com",
		ShortCode: "abc123",
		UserID:    "user1",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	repo.StoreURL(ctx, url)

	// Create an expired URL
	expiredURL := &domain.URL{
		ID:        "test2",
		LongURL:   "https://example.com/expired",
		ShortCode: "expired",
		UserID:    "user1",
		CreatedAt: time.Now().Add(-48 * time.Hour),
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}
	repo.StoreURL(ctx, expiredURL)

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "Existing URL",
			code:    "abc123",
			wantErr: false,
		},
		{
			name:    "Non-existent URL",
			code:    "notfound",
			wantErr: true,
		},
		{
			name:    "Expired URL",
			code:    "expired",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.GetURL(ctx, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("GetURL() returned nil URL")
			}
		})
	}
}

func TestURLService_DeleteURL(t *testing.T) {
	repo := newMockRepository()
	service := NewURLService(repo)
	ctx := context.Background()

	// Create a test URL
	url := &domain.URL{
		ID:        "test1",
		LongURL:   "https://example.com",
		ShortCode: "abc123",
		UserID:    "user1",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	repo.StoreURL(ctx, url)

	tests := []struct {
		name    string
		id      string
		userID  string
		wantErr bool
	}{
		{
			name:    "Owner deleting URL",
			id:      "test1",
			userID:  "user1",
			wantErr: false,
		},
		{
			name:    "Non-owner deleting URL",
			id:      "test1",
			userID:  "user2",
			wantErr: true,
		},
		{
			name:    "Non-existent URL",
			id:      "notfound",
			userID:  "user1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteURL(ctx, tt.id, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
