package mongodb

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

	t.Log("Creating MongoDB client...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "mongodb://localhost:27017", "test_db")
	if err != nil {
		t.Fatalf("Failed to create MongoDB client: %v", err)
	}
	defer func() {
		t.Log("Closing MongoDB connection...")
		client.Close(ctx)
	}()

	t.Log("Testing MongoDB ping...")
	if err := client.Database().Client().Ping(ctx, nil); err != nil {
		t.Errorf("Failed to ping MongoDB: %v", err)
	}
	t.Log("MongoDB ping successful")
}
