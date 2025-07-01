package googlesheets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/ideamans/go-sheetkv"
	"google.golang.org/api/option"
)

func TestSheetsAdaptor_LoadWithRetry(t *testing.T) {
	tests := []struct {
		name         string
		failCount    int32
		wantErr      bool
		responseData string
	}{
		{
			name:      "success on first try",
			failCount: 0,
			wantErr:   false,
			responseData: `{
				"values": [
					["name", "age"],
					["John", "30"],
					["Jane", "25"]
				]
			}`,
		},
		{
			name:      "success after one retry",
			failCount: 1,
			wantErr:   false,
			responseData: `{
				"values": [
					["name", "age"],
					["John", "30"]
				]
			}`,
		},
		{
			name:      "success after two retries",
			failCount: 2,
			wantErr:   false,
			responseData: `{
				"values": [
					["status"],
					["active"]
				]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount int32

			// Create mock server that fails initially
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				currentCall := atomic.AddInt32(&callCount, 1)

				if currentCall <= tt.failCount {
					// Return error for initial calls
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte(`{"error": {"code": 503, "message": "Service Unavailable"}}`))
					return
				}

				// Success response
				if r.URL.Path == "/v4/spreadsheets/test-id/values/TestSheet!A:ZZ" {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(tt.responseData))
				} else {
					w.WriteHeader(404)
				}
			}))
			defer server.Close()

			// Create adaptor with mock
			ctx := context.Background()
			adaptor, err := NewSheetsAdaptor(ctx, Config{
				SpreadsheetID: "test-id",
				SheetName:     "TestSheet",
			}, option.WithEndpoint(server.URL), option.WithoutAuthentication())

			if err != nil {
				t.Fatalf("Failed to create adaptor: %v", err)
			}

			// Create client with retry configuration
			client := sheetkv.New(adaptor, &sheetkv.Config{
				MaxRetries: 3,
			})

			// Initialize client
			err = client.Initialize(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && client != nil {
				// Verify data was loaded
				records, err := client.Query(sheetkv.Query{})
				if err != nil {
					t.Errorf("Query() error = %v", err)
				}

				// Count expected records based on response data
				expectedRecords := 0
				if tt.name == "success on first try" {
					expectedRecords = 2 // John and Jane
				} else if tt.name == "success after one retry" {
					expectedRecords = 1 // Just John
				} else if tt.name == "success after two retries" {
					expectedRecords = 1 // Just active status
				}

				if len(records) != expectedRecords {
					t.Errorf("Expected %d records, got %d", expectedRecords, len(records))
				}

				// Verify retry was attempted
				finalCallCount := atomic.LoadInt32(&callCount)
				expectedCalls := tt.failCount + 1
				if finalCallCount != expectedCalls {
					t.Errorf("Expected %d API calls, got %d", expectedCalls, finalCallCount)
				}

				client.Close()
			}
		})
	}
}

func TestSheetsAdaptor_SaveWithRetry(t *testing.T) {
	var callCount int32
	failCount := int32(2)

	// Create mock server that fails initially
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentCall := atomic.AddInt32(&callCount, 1)

		switch r.URL.Path {
		case "/v4/spreadsheets/test-id/values/TestSheet!A:ZZ":
			// Initial load
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"values": []}`))

		case "/v4/spreadsheets/test-id/values/TestSheet!A:ZZ:clear":
			if currentCall <= failCount+1 { // +1 because initial load counts as a call
				// Return error for initial save attempts
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error": {"code": 503, "message": "Service Unavailable"}}`))
				return
			}
			// Success
			w.Write([]byte(`{}`))

		case "/v4/spreadsheets/test-id/values/TestSheet!A1":
			// Update after clear
			w.Write([]byte(`{"updatedCells": 4}`))

		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	// Create adaptor
	ctx := context.Background()
	adaptor, err := NewSheetsAdaptor(ctx, Config{
		SpreadsheetID: "test-id",
		SheetName:     "TestSheet",
	}, option.WithEndpoint(server.URL), option.WithoutAuthentication())

	if err != nil {
		t.Fatalf("Failed to create adaptor: %v", err)
	}

	// Create client with retry
	client := sheetkv.New(adaptor, &sheetkv.Config{
		MaxRetries: 3,
	})

	// Initialize client
	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}
	defer client.Close()

	// Add a record
	err = client.Append(&sheetkv.Record{
		Values: map[string]interface{}{
			"name": "Test User",
			"age":  25,
		},
	})

	if err != nil {
		t.Errorf("Append() error = %v", err)
	}

	// Force sync (which should retry)
	err = client.Sync()
	if err != nil {
		t.Errorf("Sync() error = %v", err)
	}

	// Verify retries occurred
	finalCallCount := atomic.LoadInt32(&callCount)
	// Expected: 1 initial load + 2 failed saves + 1 successful save (clear) + 1 update
	if finalCallCount < 4 {
		t.Errorf("Expected at least 4 API calls for retries, got %d", finalCallCount)
	}
}
