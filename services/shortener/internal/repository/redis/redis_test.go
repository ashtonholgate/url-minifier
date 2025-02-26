package redis

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test")
	}

	t.Log("Creating Redis client...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "redis://localhost:6379", "", 0)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer func() {
		t.Log("Closing Redis connection...")
		client.Close()
	}()

	// Test set and get
	key := "test_key"
	value := "test_value"
	expiration := 1 * time.Minute

	t.Log("Testing Redis SET operation...")
	// Set value
	err = client.Set(ctx, key, value, expiration)
	if err != nil {
		t.Errorf("Failed to set value in Redis: %v", err)
	}
	t.Log("SET operation successful")

	t.Log("Testing Redis GET operation...")
	// Get value
	got, err := client.Get(ctx, key)
	if err != nil {
		t.Errorf("Failed to get value from Redis: %v", err)
	}
	if got != value {
		t.Errorf("Got %v, want %v", got, value)
	}
	t.Log("GET operation successful")

	t.Log("Testing Redis DEL operation...")
	// Delete value
	err = client.Del(ctx, key)
	if err != nil {
		t.Errorf("Failed to delete value from Redis: %v", err)
	}
	t.Log("DEL operation successful")
}
