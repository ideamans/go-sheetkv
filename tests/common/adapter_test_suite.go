package common

import (
	"context"
	"testing"

	sheetkv "github.com/ideamans/go-sheetkv"
)

// AdapterTestCase represents a test case for an adapter
type AdapterTestCase struct {
	Name        string
	Adapter     sheetkv.Adapter
	Description string
}

// CreateTestClient creates a test client with the given adapter
func CreateTestClient(t *testing.T, adapter sheetkv.Adapter) *sheetkv.Client {
	clientConfig := &sheetkv.Config{
		SyncInterval: 0, // No auto-sync for tests
		MaxRetries:   3,
	}

	client := sheetkv.New(adapter, clientConfig)

	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	return client
}

// CleanupClient properly closes the client
func CleanupClient(t *testing.T, client *sheetkv.Client) {
	if err := client.Sync(); err != nil {
		t.Errorf("Failed to sync before close: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Errorf("Failed to close client: %v", err)
	}
}
