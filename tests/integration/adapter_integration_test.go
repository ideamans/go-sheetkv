package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	sheetkv "github.com/ideamans/go-sheetkv"
	"github.com/ideamans/go-sheetkv/adapters/excel"
	"github.com/ideamans/go-sheetkv/adapters/googlesheets"
	"github.com/ideamans/go-sheetkv/tests/common"
)

// getSyncTestAdapters returns fresh adapters specifically for sync strategy tests
func getSyncTestAdapters(t *testing.T) []common.AdapterTestCase {
	// Load .env file if it exists
	envPath := filepath.Join("..", "..", ".env")
	if _, err := os.Stat(envPath); err == nil {
		loadEnvFile(envPath)
	}

	var adapters []common.AdapterTestCase

	// Always test Excel adapter
	tempDir := t.TempDir()
	excelFile := filepath.Join(tempDir, "sync_test.xlsx")
	excelConfig := &excel.Config{
		FilePath:  excelFile,
		SheetName: "sync",
	}
	excelAdapter, err := excel.New(excelConfig)
	if err != nil {
		t.Fatalf("Failed to create Excel adapter: %v", err)
	}
	adapters = append(adapters, common.AdapterTestCase{
		Name:        "Excel",
		Adapter:     excelAdapter,
		Description: fmt.Sprintf("Excel file: %s", excelFile),
	})

	// Test Google Sheets if configured
	spreadsheetID := os.Getenv("TEST_GOOGLE_SHEET_ID")
	if spreadsheetID != "" {
		ctx := context.Background()

		// Test with JSON file auth if available
		jsonPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if jsonPath != "" {
			// If path is relative, make it absolute
			if !filepath.IsAbs(jsonPath) {
				jsonPath = filepath.Join("..", "..", jsonPath)
			}

			gsConfig := googlesheets.Config{
				SpreadsheetID: spreadsheetID,
				SheetName:     "sync",
			}
			adapter, err := googlesheets.NewWithJSONKeyFile(ctx, gsConfig, jsonPath)
			if err == nil {
				adapters = append(adapters, common.AdapterTestCase{
					Name:        "GoogleSheets-JSON",
					Adapter:     adapter,
					Description: "Google Sheets with JSON file auth",
				})
			}
		}

		// Test with email/key auth if available
		email := os.Getenv("TEST_CLIENT_EMAIL")
		privateKey := os.Getenv("TEST_CLIENT_PRIVATE_KEY")
		if email != "" && privateKey != "" {
			// In CI, the private key might have literal \n instead of actual newlines
			// Apply the same transformation that loadEnvFile does
			if !strings.Contains(privateKey, "\n") && strings.Contains(privateKey, "\\n") {
				privateKey = strings.ReplaceAll(privateKey, "\\n", "\n")
			}

			gsConfig := googlesheets.Config{
				SpreadsheetID: spreadsheetID,
				SheetName:     "sync",
			}
			adapter, err := googlesheets.NewWithServiceAccountKey(ctx, gsConfig, email, privateKey)
			if err == nil {
				adapters = append(adapters, common.AdapterTestCase{
					Name:        "GoogleSheets-EmailKey",
					Adapter:     adapter,
					Description: "Google Sheets with email/key auth",
				})
			}
		}
	}

	return adapters
}

// getTestAdapters returns all adapters to test
func getTestAdapters(t *testing.T) []common.AdapterTestCase {
	// Load .env file if it exists
	envPath := filepath.Join("..", "..", ".env")
	if _, err := os.Stat(envPath); err == nil {
		loadEnvFile(envPath)
	}

	var adapters []common.AdapterTestCase

	// Always test Excel adapter
	tempDir := t.TempDir()
	excelFile := filepath.Join(tempDir, "integration_test.xlsx")
	excelConfig := &excel.Config{
		FilePath:  excelFile,
		SheetName: "integration",
	}
	excelAdapter, err := excel.New(excelConfig)
	if err != nil {
		t.Fatalf("Failed to create Excel adapter: %v", err)
	}
	adapters = append(adapters, common.AdapterTestCase{
		Name:        "Excel",
		Adapter:     excelAdapter,
		Description: fmt.Sprintf("Excel file: %s", excelFile),
	})

	// Test Google Sheets if configured
	spreadsheetID := os.Getenv("TEST_GOOGLE_SHEET_ID")
	if spreadsheetID == "" {
		t.Log("⚠️  Skipping Google Sheets tests: TEST_GOOGLE_SHEET_ID not set")
	} else {
		ctx := context.Background()

		// Test with JSON file auth if available
		jsonPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if jsonPath != "" {
			// If path is relative, make it absolute
			if !filepath.IsAbs(jsonPath) {
				jsonPath = filepath.Join("..", "..", jsonPath)
			}

			gsConfig := googlesheets.Config{
				SpreadsheetID: spreadsheetID,
				SheetName:     "integration",
			}
			adapter, err := googlesheets.NewWithJSONKeyFile(ctx, gsConfig, jsonPath)
			if err != nil {
				t.Logf("⚠️  Failed to create Google Sheets adapter with JSON auth: %v", err)
			} else {
				adapters = append(adapters, common.AdapterTestCase{
					Name:        "GoogleSheets-JSON",
					Adapter:     adapter,
					Description: "Google Sheets with JSON file auth",
				})
			}
		} else {
			t.Log("⚠️  Skipping Google Sheets JSON auth test: GOOGLE_APPLICATION_CREDENTIALS not set")
		}

		// Test with email/key auth if available
		email := os.Getenv("TEST_CLIENT_EMAIL")
		privateKey := os.Getenv("TEST_CLIENT_PRIVATE_KEY")
		if email != "" && privateKey != "" {

			// In CI, the private key might have literal \n instead of actual newlines
			// Apply the same transformation that loadEnvFile does
			if !strings.Contains(privateKey, "\n") && strings.Contains(privateKey, "\\n") {
				privateKey = strings.ReplaceAll(privateKey, "\\n", "\n")
			}

			gsConfig := googlesheets.Config{
				SpreadsheetID: spreadsheetID,
				SheetName:     "integration",
			}
			adapter, err := googlesheets.NewWithServiceAccountKey(ctx, gsConfig, email, privateKey)
			if err != nil {
				t.Logf("⚠️  Failed to create Google Sheets adapter with email/key auth: %v", err)
			} else {
				adapters = append(adapters, common.AdapterTestCase{
					Name:        "GoogleSheets-EmailKey",
					Adapter:     adapter,
					Description: "Google Sheets with email/key auth",
				})
			}
		} else {
			t.Log("⚠️  Skipping Google Sheets email/key auth test: TEST_CLIENT_EMAIL or TEST_CLIENT_PRIVATE_KEY not set")
		}
	}

	return adapters
}

func TestAdapterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	adapters := getTestAdapters(t)
	if len(adapters) == 0 {
		t.Fatal("No adapters available for testing")
	}

	for _, tc := range adapters {
		t.Run(tc.Name, func(t *testing.T) {
			t.Logf("Testing with %s", tc.Description)

			client := common.CreateTestClient(t, tc.Adapter)
			defer common.CleanupClient(t, client)

			// Run all test suites
			t.Run("BasicCRUD", func(t *testing.T) {
				testBasicCRUD(t, client)
			})

			t.Run("DataTypes", func(t *testing.T) {
				testDataTypes(t, client)
			})

			t.Run("QueryOperations", func(t *testing.T) {
				testQueryOperations(t, client)
			})

			t.Run("SyncFunctionality", func(t *testing.T) {
				testSyncFunctionality(t, client, tc.Adapter)
			})

			t.Run("LargeDataSet", func(t *testing.T) {
				testLargeDataSet(t, client)
			})

			// Note: SyncStrategies test is run separately with fresh adapters
		})
	}

	// Run sync strategy tests separately with fresh adapters to avoid interference
	t.Run("SyncStrategies", func(t *testing.T) {
		syncAdapters := getSyncTestAdapters(t)
		for _, tc := range syncAdapters {
			t.Run(tc.Name, func(t *testing.T) {
				testSyncStrategies(t, tc.Adapter)
			})
		}
	})
}

// testBasicCRUD tests basic create, read, update, delete operations
func testBasicCRUD(t *testing.T, client *sheetkv.Client) {
	// Clear existing data
	clearAllRecords(t, client)

	// Test Append
	record1 := &sheetkv.Record{
		Values: map[string]any{
			"id":     int64(1),
			"name":   "Test User 1",
			"email":  "test1@example.com",
			"age":    int64(25),
			"active": true,
		},
	}

	err := client.Append(record1)
	if err != nil {
		t.Fatalf("Failed to append record: %v", err)
	}

	// Test Get
	retrieved, err := client.Get(2) // Key should be 2 (row 2)
	if err != nil {
		t.Fatalf("Failed to get record: %v", err)
	}

	if retrieved.GetAsString("name", "") != "Test User 1" {
		t.Errorf("Retrieved name = %s, want Test User 1", retrieved.GetAsString("name", ""))
	}

	// Test Update
	err = client.Update(2, map[string]any{
		"email": "updated@example.com",
		"age":   int64(26),
	})
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// Verify update
	updated, err := client.Get(2)
	if err != nil {
		t.Fatalf("Failed to get updated record: %v", err)
	}

	if updated.GetAsString("email", "") != "updated@example.com" {
		t.Errorf("Updated email = %s, want updated@example.com", updated.GetAsString("email", ""))
	}

	// Test Delete
	err = client.Delete(2)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// Verify deletion
	_, err = client.Get(2)
	if err != sheetkv.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after deletion, got %v", err)
	}
}

// testDataTypes tests various data type conversions
func testDataTypes(t *testing.T, client *sheetkv.Client) {
	clearAllRecords(t, client)

	// Test various data types
	record := &sheetkv.Record{
		Values: map[string]any{
			"string_val": "hello",
			"int_val":    int64(42),
			"float_val":  3.14,
			"bool_val":   true,
			"time_val":   time.Now().Format(time.RFC3339),
		},
	}

	record.SetStrings("tags", []string{"tag1", "tag2", "tag3"})

	err := client.Append(record)
	if err != nil {
		t.Fatalf("Failed to append record: %v", err)
	}

	// Force sync and reload
	err = client.Sync()
	if err != nil {
		t.Fatalf("Failed to sync: %v", err)
	}

	// Query the record
	results, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "string_val", Operator: "==", Value: "hello"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	retrieved := results[0]

	// Verify types
	if retrieved.GetAsString("string_val", "") != "hello" {
		t.Errorf("String value mismatch")
	}
	if retrieved.GetAsInt64("int_val", 0) != 42 {
		t.Errorf("Int value mismatch")
	}
	if retrieved.GetAsFloat64("float_val", 0) != 3.14 {
		t.Errorf("Float value mismatch")
	}
	if retrieved.GetAsBool("bool_val", false) != true {
		t.Errorf("Bool value mismatch")
	}

	tags := retrieved.GetAsStrings("tags", []string{})
	if len(tags) != 3 || tags[0] != "tag1" || tags[1] != "tag2" || tags[2] != "tag3" {
		t.Errorf("Tags mismatch: %v", tags)
	}
}

// testQueryOperations tests various query conditions
func testQueryOperations(t *testing.T, client *sheetkv.Client) {
	clearAllRecords(t, client)

	// Add test data
	testData := []sheetkv.Record{
		{Values: map[string]any{"name": "Alice", "age": int64(25), "department": "Engineering", "active": true}},
		{Values: map[string]any{"name": "Bob", "age": int64(30), "department": "Sales", "active": true}},
		{Values: map[string]any{"name": "Charlie", "age": int64(35), "department": "Engineering", "active": false}},
		{Values: map[string]any{"name": "David", "age": int64(28), "department": "Marketing", "active": true}},
		{Values: map[string]any{"name": "Eve", "age": int64(32), "department": "Sales", "active": false}},
	}

	for _, record := range testData {
		if err := client.Append(&record); err != nil {
			t.Fatalf("Failed to append test data: %v", err)
		}
	}

	// Test equality
	results, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "department", Operator: "==", Value: "Engineering"},
		},
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 Engineering results, got %d", len(results))
	}

	// Test range
	results, err = client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "age", Operator: ">=", Value: int64(30)},
			{Column: "age", Operator: "<=", Value: int64(35)},
		},
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results in age range, got %d", len(results))
	}

	// Test IN operator
	results, err = client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "department", Operator: "in", Value: []any{"Sales", "Marketing"}},
		},
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results for Sales/Marketing, got %d", len(results))
	}
}

// testSyncFunctionality tests sync between client and backend
func testSyncFunctionality(t *testing.T, client *sheetkv.Client, adapter sheetkv.Adapter) {
	clearAllRecords(t, client)

	// Add a record
	record := &sheetkv.Record{
		Values: map[string]any{
			"id":   int64(100),
			"name": "Sync Test",
		},
	}
	err := client.Append(record)
	if err != nil {
		t.Fatalf("Failed to append record: %v", err)
	}

	// Force sync
	err = client.Sync()
	if err != nil {
		t.Fatalf("Failed to sync: %v", err)
	}

	// Create a new client to verify data was persisted
	client2 := common.CreateTestClient(t, adapter)
	defer common.CleanupClient(t, client2)

	// Query for the record
	results, err := client2.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "name", Operator: "==", Value: "Sync Test"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to query from second client: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// testLargeDataSet tests performance with larger datasets
func testLargeDataSet(t *testing.T, client *sheetkv.Client) {
	clearAllRecords(t, client)

	// Insert 100 records
	recordCount := 100
	for i := 1; i <= recordCount; i++ {
		record := &sheetkv.Record{
			Values: map[string]any{
				"id":         int64(i),
				"name":       "User " + strconv.Itoa(i),
				"email":      "user" + strconv.Itoa(i) + "@example.com",
				"age":        int64(20 + i%50),
				"score":      float64(i) * 1.5,
				"active":     i%2 == 0,
				"department": "Dept" + strconv.Itoa(i%5),
			},
		}

		err := client.Append(record)
		if err != nil {
			t.Fatalf("Failed to append record %d: %v", i, err)
		}
	}

	// Test query performance
	results, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "department", Operator: "==", Value: "Dept1"},
		},
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 20 { // 100 records / 5 departments
		t.Errorf("Expected 20 results, got %d", len(results))
	}

	// Test with limit and offset
	results, err = client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "active", Operator: "==", Value: true},
		},
		Limit:  10,
		Offset: 5,
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("Expected 10 results with limit, got %d", len(results))
	}
}

// clearAllRecords removes all records from the sheet
func clearAllRecords(t *testing.T, client *sheetkv.Client) {
	records, err := client.Query(sheetkv.Query{})
	if err != nil {
		t.Fatalf("Failed to query records for clearing: %v", err)
	}

	for _, record := range records {
		if err := client.Delete(record.Key); err != nil {
			t.Errorf("Failed to delete record %d: %v", record.Key, err)
		}
	}

	if err := client.Sync(); err != nil {
		t.Fatalf("Failed to sync after clearing: %v", err)
	}
}

// testSyncStrategies tests both gap-preserving and compacting sync strategies
func testSyncStrategies(t *testing.T, adapter sheetkv.Adapter) {
	ctx := context.Background()

	t.Run("Gap-Preserving and Compacting Strategies", func(t *testing.T) {
		// Note: For sync strategy tests, we need to ensure clean state
		// Clear the adapter's data directly first
		if err := adapter.Save(ctx, []*sheetkv.Record{}, []string{}, sheetkv.SyncStrategyCompacting); err != nil {
			// Skip if Google Sheets sheet doesn't exist
			if strings.Contains(err.Error(), "Unable to parse range") || strings.Contains(err.Error(), "badRequest") {
				t.Skipf("Skipping test - sheet may not exist: %v", err)
			}
			t.Fatalf("Failed to clear adapter data: %v", err)
		}

		// Create a client to manage cache
		config := &sheetkv.Config{
			SyncInterval:  30 * time.Second, // Long interval to prevent auto-sync
			MaxRetries:    3,
			RetryInterval: 100 * time.Millisecond,
		}
		client := sheetkv.New(adapter, config)
		if err := client.Initialize(ctx); err != nil {
			t.Fatalf("Failed to initialize client: %v", err)
		}
		// Note: We don't use defer client.Close() here because we test Close() explicitly

		// Step 1: Add records with gaps
		records := []*sheetkv.Record{
			{Values: map[string]any{"id": int64(1), "name": "Alice", "status": "active"}},
			{Values: map[string]any{"id": int64(2), "name": "Bob", "status": "active"}},
			{Values: map[string]any{"id": int64(3), "name": "Charlie", "status": "inactive"}},
			{Values: map[string]any{"id": int64(4), "name": "David", "status": "active"}},
			{Values: map[string]any{"id": int64(5), "name": "Eve", "status": "active"}},
		}

		// Append all records
		for _, r := range records {
			if err := client.Append(r); err != nil {
				t.Fatalf("Failed to append record: %v", err)
			}
		}

		// Delete some records to create gaps (Bob and David)
		if err := client.Delete(3); err != nil { // Bob is at row 3
			t.Fatalf("Failed to delete Bob: %v", err)
		}
		if err := client.Delete(5); err != nil { // David is at row 5
			t.Fatalf("Failed to delete David: %v", err)
		}

		// Step 2: Test Gap-Preserving Sync
		t.Run("Gap-Preserving Sync", func(t *testing.T) {
			// Force sync with gap-preserving (default for Sync())
			if err := client.Sync(); err != nil {
				t.Fatalf("Gap-preserving sync failed: %v", err)
			}

			// Load data directly from adapter to verify
			loadedRecords, _, err := adapter.Load(ctx)
			if err != nil {
				t.Fatalf("Failed to load after gap-preserving sync: %v", err)
			}

			// Should have exactly 5 records (including empty rows for deleted ones)
			if len(loadedRecords) != 5 {
				t.Errorf("Expected 5 records with gaps, got %d", len(loadedRecords))
			}

			// Verify specific positions
			for _, r := range loadedRecords {
				switch r.Key {
				case 2: // Alice
					if name := r.GetAsString("name", ""); name != "Alice" {
						t.Errorf("Row 2 should be Alice, got %s", name)
					}
				case 3: // Deleted (Bob)
					if name := r.GetAsString("name", ""); name != "" {
						t.Errorf("Row 3 should be empty (deleted Bob), got %s", name)
					}
				case 4: // Charlie
					if name := r.GetAsString("name", ""); name != "Charlie" {
						t.Errorf("Row 4 should be Charlie, got %s", name)
					}
				case 5: // Deleted (David)
					if name := r.GetAsString("name", ""); name != "" {
						t.Errorf("Row 5 should be empty (deleted David), got %s", name)
					}
				case 6: // Eve
					if name := r.GetAsString("name", ""); name != "Eve" {
						t.Errorf("Row 6 should be Eve, got %s", name)
					}
				}
			}
		})

		// Step 3: Add more data to test compacting with trailing cleanup
		moreRecords := []*sheetkv.Record{
			{Values: map[string]any{"id": int64(10), "name": "Frank", "status": "active"}},
			{Values: map[string]any{"id": int64(11), "name": "Grace", "status": "active"}},
		}

		for _, r := range moreRecords {
			if err := client.Append(r); err != nil {
				t.Fatalf("Failed to append additional record: %v", err)
			}
		}

		// Delete Grace to test trailing cleanup
		if err := client.Delete(8); err != nil { // Grace should be at row 8
			t.Fatalf("Failed to delete Grace: %v", err)
		}

		// Step 4: Test Compacting Sync (via Close)
		t.Run("Compacting Sync", func(t *testing.T) {
			// Close will trigger compacting sync
			if err := client.Close(); err != nil {
				t.Fatalf("Close (compacting sync) failed: %v", err)
			}

			// Load data directly from adapter to verify
			loadedRecords, _, err := adapter.Load(ctx)
			if err != nil {
				t.Fatalf("Failed to load after compacting sync: %v", err)
			}

			// Should have exactly 4 records (Alice, Charlie, Eve, Frank - no gaps)
			if len(loadedRecords) != 4 {
				t.Errorf("Expected 4 compacted records, got %d", len(loadedRecords))
			}

			// Verify records are compacted (sequential from row 2)
			expectedNames := []string{"Alice", "Charlie", "Eve", "Frank"}
			for i, r := range loadedRecords {
				expectedKey := i + 2 // Start from row 2
				if r.Key != expectedKey {
					t.Errorf("Record %d: expected key %d, got %d", i, expectedKey, r.Key)
				}
				if i < len(expectedNames) {
					if name := r.GetAsString("name", ""); name != expectedNames[i] {
						t.Errorf("Record %d: expected name %s, got %s", i, expectedNames[i], name)
					}
				}
			}

			// Verify no trailing data beyond row 5
			for _, r := range loadedRecords {
				if r.Key > 5 {
					t.Errorf("Found unexpected record at row %d after compacting", r.Key)
				}
			}
		})
	})
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		// Convert \n escape sequences to actual newlines for private keys
		if key == "TEST_CLIENT_PRIVATE_KEY" {
			value = strings.ReplaceAll(value, "\\n", "\n")
		}

		os.Setenv(key, value)
	}

	return nil
}
