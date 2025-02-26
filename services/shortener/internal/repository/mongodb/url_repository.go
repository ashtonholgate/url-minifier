package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ashtonholgate/url-minifier/services/shortener/internal/domain"
	"github.com/ashtonholgate/url-minifier/services/shortener/internal/repository/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	urlCollection = "urls"
	cacheDuration = 24 * time.Hour
)

// URLRepository implements repository.Repository interface using MongoDB and Redis
type URLRepository struct {
	mongoClient *Client
	redisClient *redis.Client
	collection  *mongo.Collection
}

// NewURLRepository creates a new URLRepository instance
func NewURLRepository(mongoClient *Client, redisClient *redis.Client) *URLRepository {
	collection := mongoClient.Collection(urlCollection)

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Index on short_code (unique)
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "short_code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create short_code index: %v", err))
	}

	// Index on user_id
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create user_id index: %v", err))
	}

	return &URLRepository{
		mongoClient: mongoClient,
		redisClient: redisClient,
		collection:  collection,
	}
}

// StoreURL saves a URL to MongoDB and updates Redis cache
func (r *URLRepository) StoreURL(ctx context.Context, url *domain.URL) error {
	// Convert to BSON document
	doc := bson.M{
		"_id":          url.ID,
		"long_url":     url.LongURL,
		"short_code":   url.ShortCode,
		"user_id":      url.UserID,
		"created_at":   url.CreatedAt,
		"expires_at":   url.ExpiresAt,
		"custom_alias": url.CustomAlias,
	}

	// Store in MongoDB
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to store URL in MongoDB: %w", err)
	}

	// Cache in Redis
	return r.cacheURL(ctx, url)
}

// GetURLByCode retrieves a URL by its short code
func (r *URLRepository) GetURLByCode(ctx context.Context, code string) (*domain.URL, error) {
	// Try cache first
	url, err := r.getFromCache(ctx, fmt.Sprintf("code:%s", code))
	if err == nil {
		return url, nil
	}

	// Fallback to MongoDB
	var result domain.URL
	err = r.collection.FindOne(ctx, bson.M{"short_code": code}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, domain.ErrURLNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get URL from MongoDB: %w", err)
	}

	// Update cache
	if err := r.cacheURL(ctx, &result); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to cache URL: %v\n", err)
	}

	return &result, nil
}

// GetURLByID retrieves a URL by its ID
func (r *URLRepository) GetURLByID(ctx context.Context, id string) (*domain.URL, error) {
	// Try cache first
	url, err := r.getFromCache(ctx, fmt.Sprintf("id:%s", id))
	if err == nil {
		return url, nil
	}

	// Fallback to MongoDB
	var result domain.URL
	err = r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, domain.ErrURLNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get URL from MongoDB: %w", err)
	}

	// Update cache
	if err := r.cacheURL(ctx, &result); err != nil {
		fmt.Printf("Failed to cache URL: %v\n", err)
	}

	return &result, nil
}

// DeleteURL removes a URL from both MongoDB and Redis
func (r *URLRepository) DeleteURL(ctx context.Context, id string) error {
	// Delete from MongoDB
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete URL from MongoDB: %w", err)
	}
	if result.DeletedCount == 0 {
		return domain.ErrURLNotFound
	}

	// Get the URL first to get the short code
	url, err := r.GetURLByID(ctx, id)
	if err != nil && err != domain.ErrURLNotFound {
		// Log the error but continue with deletion
		fmt.Printf("Failed to get URL for cache cleanup: %v\n", err)
	}

	// Remove from Redis cache
	if err := r.redisClient.Del(ctx, fmt.Sprintf("id:%s", id)); err != nil {
		fmt.Printf("Failed to remove URL from Redis cache (ID): %v\n", err)
	}

	// If we have the URL, also remove the short code cache
	if url != nil {
		if err := r.redisClient.Del(ctx, fmt.Sprintf("code:%s", url.ShortCode)); err != nil {
			fmt.Printf("Failed to remove URL from Redis cache (code): %v\n", err)
		}
	}

	return nil
}

// IsCodeAvailable checks if a short code is available
func (r *URLRepository) IsCodeAvailable(ctx context.Context, code string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"short_code": code})
	if err != nil {
		return false, fmt.Errorf("failed to check code availability: %w", err)
	}
	return count == 0, nil
}

// ListURLsByUser retrieves all URLs for a given user
func (r *URLRepository) ListURLsByUser(ctx context.Context, userID string) ([]*domain.URL, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to list URLs: %w", err)
	}
	defer cursor.Close(ctx)

	var urls []*domain.URL
	if err := cursor.All(ctx, &urls); err != nil {
		return nil, fmt.Errorf("failed to decode URLs: %w", err)
	}

	return urls, nil
}

// Close cleans up repository resources
func (r *URLRepository) Close(ctx context.Context) error {
	if err := r.mongoClient.Close(ctx); err != nil {
		return fmt.Errorf("failed to close MongoDB client: %w", err)
	}
	if err := r.redisClient.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}
	return nil
}

// Helper functions for caching

func (r *URLRepository) cacheURL(ctx context.Context, url *domain.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	// Cache by ID
	if err := r.redisClient.Set(ctx, fmt.Sprintf("id:%s", url.ID), string(data), cacheDuration); err != nil {
		return err
	}

	// Cache by short code
	return r.redisClient.Set(ctx, fmt.Sprintf("code:%s", url.ShortCode), string(data), cacheDuration)
}

func (r *URLRepository) getFromCache(ctx context.Context, key string) (*domain.URL, error) {
	data, err := r.redisClient.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var url domain.URL
	if err := json.Unmarshal([]byte(data), &url); err != nil {
		return nil, err
	}

	return &url, nil
}
