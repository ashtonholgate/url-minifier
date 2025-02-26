package mongodb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ashtonholgate/url-minifier/services/shortener/internal/domain"
	"github.com/ashtonholgate/url-minifier/services/shortener/internal/repository/redis"
)

func setupTestRepository(t *testing.T) (*URLRepository, context.Context, func()) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	t.Log("Creating MongoDB client...")
	mongoClient, err := NewClient(ctx, "mongodb://localhost:27017", "test_db")
	if err != nil {
		t.Fatalf("Failed to create MongoDB client: %v", err)
	}

	t.Log("Creating Redis client...")
	redisClient, err := redis.NewClient(ctx, "redis://localhost:6379", "", 0)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}

	repo := NewURLRepository(mongoClient, redisClient)

	// Create cleanup function
	cleanup := func() {
		t.Log("Cleaning up test data...")
		// Drop the collection
		if err := repo.collection.Drop(ctx); err != nil {
			t.Errorf("Failed to drop collection: %v", err)
		}
		// Close connections
		if err := repo.Close(ctx); err != nil {
			t.Errorf("Failed to close repository: %v", err)
		}
	}

	return repo, ctx, cleanup
}

func TestURLRepository_StoreAndRetrieve(t *testing.T) {
	repo, ctx, cleanup := setupTestRepository(t)
	defer cleanup()

	// Create test URL
	url := &domain.URL{
		ID:        "test-id",
		LongURL:   "https://example.com",
		ShortCode: "abc123",
		UserID:    "user1",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Test storing URL
	t.Log("Testing URL storage...")
	err := repo.StoreURL(ctx, url)
	if err != nil {
		t.Fatalf("Failed to store URL: %v", err)
	}

	// Test retrieving by code
	t.Log("Testing URL retrieval by code...")
	retrieved, err := repo.GetURLByCode(ctx, url.ShortCode)
	if err != nil {
		t.Fatalf("Failed to retrieve URL by code: %v", err)
	}
	if retrieved.ID != url.ID || retrieved.LongURL != url.LongURL {
		t.Errorf("Retrieved URL doesn't match stored URL")
	}

	// Test retrieving by ID
	t.Log("Testing URL retrieval by ID...")
	retrieved, err = repo.GetURLByID(ctx, url.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve URL by ID: %v", err)
	}
	if retrieved.ID != url.ID || retrieved.LongURL != url.LongURL {
		t.Errorf("Retrieved URL doesn't match stored URL")
	}
}

func TestURLRepository_ListURLsByUser(t *testing.T) {
	repo, ctx, cleanup := setupTestRepository(t)
	defer cleanup()

	// Create test URLs for two users
	urls := []*domain.URL{
		{
			ID:        "test-id-1",
			LongURL:   "https://example1.com",
			ShortCode: "abc123",
			UserID:    "user1",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:        "test-id-2",
			LongURL:   "https://example2.com",
			ShortCode: "def456",
			UserID:    "user1",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:        "test-id-3",
			LongURL:   "https://example3.com",
			ShortCode: "ghi789",
			UserID:    "user2",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	// Store URLs
	t.Log("Storing test URLs...")
	for _, url := range urls {
		if err := repo.StoreURL(ctx, url); err != nil {
			t.Fatalf("Failed to store URL: %v", err)
		}
	}

	// Test listing URLs for user1
	t.Log("Testing URL listing for user1...")
	user1URLs, err := repo.ListURLsByUser(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to list URLs for user1: %v", err)
	}
	if len(user1URLs) != 2 {
		t.Errorf("Expected 2 URLs for user1, got %d", len(user1URLs))
	}

	// Test listing URLs for user2
	t.Log("Testing URL listing for user2...")
	user2URLs, err := repo.ListURLsByUser(ctx, "user2")
	if err != nil {
		t.Fatalf("Failed to list URLs for user2: %v", err)
	}
	if len(user2URLs) != 1 {
		t.Errorf("Expected 1 URL for user2, got %d", len(user2URLs))
	}
}

func TestURLRepository_DeleteURL(t *testing.T) {
	repo, ctx, cleanup := setupTestRepository(t)
	defer cleanup()

	// Create test URL
	url := &domain.URL{
		ID:        "test-id",
		LongURL:   "https://example.com",
		ShortCode: "abc123",
		UserID:    "user1",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Store URL
	t.Log("Storing test URL...")
	err := repo.StoreURL(ctx, url)
	if err != nil {
		t.Fatalf("Failed to store URL: %v", err)
	}

	// Delete URL
	t.Log("Testing URL deletion...")
	err = repo.DeleteURL(ctx, url.ID)
	if err != nil {
		t.Fatalf("Failed to delete URL: %v", err)
	}

	// Verify deletion
	t.Log("Verifying URL deletion...")
	_, err = repo.GetURLByID(ctx, url.ID)
	if err != domain.ErrURLNotFound {
		t.Errorf("Expected URL not found error, got: %v", err)
	}
}
