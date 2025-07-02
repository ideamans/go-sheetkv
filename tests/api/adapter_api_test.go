package api

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// getAPITestAdapters returns all adapters to test for API tests
func getAPITestAdapters(t *testing.T) []common.AdapterTestCase {
	// Load .env file if it exists
	envPath := filepath.Join("..", "..", ".env")
	if _, err := os.Stat(envPath); err == nil {
		loadEnvFile(envPath)
	}

	var adapters []common.AdapterTestCase

	// Always test Excel adapter
	tempDir := t.TempDir()
	excelFile := filepath.Join(tempDir, "api_test.xlsx")
	excelConfig := &excel.Config{
		FilePath:  excelFile,
		SheetName: "api",
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
				SheetName:     "api",
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
				SheetName:     "api",
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

func TestAPIOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	adapters := getAPITestAdapters(t)
	if len(adapters) == 0 {
		t.Fatal("No adapters available for testing")
	}

	for _, tc := range adapters {
		t.Run(tc.Name, func(t *testing.T) {
			t.Logf("Testing API with %s", tc.Description)

			client := common.CreateTestClient(t, tc.Adapter)
			defer common.CleanupClient(t, client)

			// Run all test scenarios
			t.Run("BasicCRUD", func(t *testing.T) {
				testBasicCRUD(t, client)
			})

			t.Run("DataTypes", func(t *testing.T) {
				testDataTypes(t, client)
			})

			t.Run("QueryOperations", func(t *testing.T) {
				testQueryOperations(t, client)
			})

			t.Run("ConcurrentOperations", func(t *testing.T) {
				testConcurrentOperations(t, client)
			})

			t.Run("LargeDataSet", func(t *testing.T) {
				testLargeDataSet(t, client)
			})
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

// testBasicCRUD tests basic create, read, update, delete operations
func testBasicCRUD(t *testing.T, client *sheetkv.Client) {
	clearAllRecords(t, client)

	// Test Append
	record1 := &sheetkv.Record{
		Values: map[string]any{
			"name":   "Alice Johnson",
			"email":  "alice@example.com",
			"age":    int64(30),
			"status": "active",
		},
	}

	err := client.Append(record1)
	if err != nil {
		t.Fatalf("Failed to append record: %v", err)
	}

	if record1.Key <= 1 {
		t.Errorf("Expected record key > 1, got %d", record1.Key)
	}

	// Test Get
	retrieved, err := client.Get(record1.Key)
	if err != nil {
		t.Fatalf("Failed to get record: %v", err)
	}

	if retrieved.GetAsString("name", "") != "Alice Johnson" {
		t.Errorf("Retrieved name = %s, want Alice Johnson", retrieved.GetAsString("name", ""))
	}

	// Test Set (full replacement)
	newRecord := &sheetkv.Record{
		Key: record1.Key,
		Values: map[string]any{
			"name":  "Alice Smith",
			"email": "alice.smith@example.com",
			"age":   int64(31),
		},
	}

	err = client.Set(record1.Key, newRecord)
	if err != nil {
		t.Fatalf("Failed to set record: %v", err)
	}

	// Test Update (partial update)
	updates := map[string]any{
		"status":     "inactive",
		"updated_at": time.Now().Format(time.RFC3339),
	}

	err = client.Update(record1.Key, updates)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// Verify updates
	final, err := client.Get(record1.Key)
	if err != nil {
		t.Fatalf("Failed to get final record: %v", err)
	}

	if final.GetAsString("status", "") != "inactive" {
		t.Errorf("Status = %s, want inactive", final.GetAsString("status", ""))
	}

	// Test Delete
	err = client.Delete(record1.Key)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	_, err = client.Get(record1.Key)
	if err != sheetkv.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

// testDataTypes tests type conversion and preservation
func testDataTypes(t *testing.T, client *sheetkv.Client) {
	clearAllRecords(t, client)

	// Create record with various types
	now := time.Now()
	record := &sheetkv.Record{
		Values: make(map[string]any),
	}

	// Use type-safe setters
	record.SetString("string_field", "Hello, 世界")
	record.SetInt64("int_field", 42)
	record.SetFloat64("float_field", 3.14159)
	record.SetBool("bool_field", true)
	record.SetStrings("array_field", []string{"apple", "banana", "cherry"})
	record.SetTime("time_field", now)

	// Also test direct value setting
	record.Values["null_field"] = nil
	record.Values["zero_int"] = int64(0)
	record.Values["empty_string"] = ""

	err := client.Append(record)
	if err != nil {
		t.Fatalf("Failed to append record: %v", err)
	}

	// Force sync to ensure data goes through serialization
	err = client.Sync()
	if err != nil {
		t.Fatalf("Failed to sync: %v", err)
	}

	// Retrieve and verify
	retrieved, err := client.Get(record.Key)
	if err != nil {
		t.Fatalf("Failed to retrieve record: %v", err)
	}

	// Test getters
	if s := retrieved.GetAsString("string_field", ""); s != "Hello, 世界" {
		t.Errorf("String field = %s, want Hello, 世界", s)
	}

	if i := retrieved.GetAsInt64("int_field", 0); i != 42 {
		t.Errorf("Int field = %d, want 42", i)
	}

	if f := retrieved.GetAsFloat64("float_field", 0); f != 3.14159 {
		t.Errorf("Float field = %f, want 3.14159", f)
	}

	if b := retrieved.GetAsBool("bool_field", false); b != true {
		t.Errorf("Bool field = %v, want true", b)
	}

	arr := retrieved.GetAsStrings("array_field", nil)
	if len(arr) != 3 || arr[0] != "apple" || arr[1] != "banana" || arr[2] != "cherry" {
		t.Errorf("Array field = %v, want [apple banana cherry]", arr)
	}

	// Time comparison (allow some tolerance for serialization)
	retrievedTime := retrieved.GetAsTime("time_field", time.Time{})
	if retrievedTime.Unix() != now.Unix() {
		t.Errorf("Time field differs: got %v, want %v", retrievedTime, now)
	}

	// Test edge cases
	// Note: null fields might not be stored in some adapters, so they return the default
	nullValue := retrieved.GetAsString("null_field", "default")
	if nullValue != "" && nullValue != "default" && nullValue != "<nil>" {
		t.Errorf("Null field should return empty string, default, or <nil>, got %q", nullValue)
	}

	if retrieved.GetAsInt64("zero_int", -1) != 0 {
		t.Errorf("Zero int should be preserved")
	}

	if retrieved.GetAsString("empty_string", "default") != "" {
		t.Errorf("Empty string should be preserved")
	}
}

// testQueryOperations tests various query features
func testQueryOperations(t *testing.T, client *sheetkv.Client) {
	clearAllRecords(t, client)

	// Create test dataset
	testData := []struct {
		name       string
		age        int64
		salary     float64
		department string
		active     bool
		tags       []string
		joinDate   time.Time
	}{
		{"Alice", 25, 50000, "Engineering", true, []string{"backend", "go"}, time.Now().AddDate(-2, 0, 0)},
		{"Bob", 30, 60000, "Engineering", true, []string{"frontend", "react"}, time.Now().AddDate(-1, -6, 0)},
		{"Charlie", 35, 70000, "Sales", false, []string{"senior", "leader"}, time.Now().AddDate(-5, 0, 0)},
		{"David", 28, 55000, "Marketing", true, []string{"digital", "seo"}, time.Now().AddDate(-1, 0, 0)},
		{"Eve", 32, 65000, "Sales", true, []string{"senior", "closer"}, time.Now().AddDate(-3, 0, 0)},
		{"Frank", 27, 52000, "Engineering", true, []string{"backend", "python"}, time.Now().AddDate(0, -6, 0)},
		{"Grace", 29, 58000, "Marketing", false, []string{"content", "writer"}, time.Now().AddDate(-2, -3, 0)},
		{"Henry", 31, 62000, "Engineering", true, []string{"fullstack", "lead"}, time.Now().AddDate(-4, 0, 0)},
		{"Iris", 26, 51000, "Sales", true, []string{"junior", "eager"}, time.Now().AddDate(0, -3, 0)},
		{"Jack", 33, 68000, "Marketing", true, []string{"senior", "strategy"}, time.Now().AddDate(-4, -6, 0)},
	}

	for _, data := range testData {
		record := &sheetkv.Record{
			Values: map[string]any{
				"name":       data.name,
				"age":        data.age,
				"salary":     data.salary,
				"department": data.department,
				"active":     data.active,
				"join_date":  data.joinDate.Format(time.RFC3339),
			},
		}
		record.SetStrings("tags", data.tags)

		if err := client.Append(record); err != nil {
			t.Fatalf("Failed to append test data: %v", err)
		}
	}

	// Test 1: Simple equality
	results, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "department", Operator: "==", Value: "Engineering"},
		},
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Expected 4 Engineering employees, got %d", len(results))
	}

	// Test 2: Range queries
	results, err = client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "age", Operator: ">=", Value: int64(30)},
			{Column: "age", Operator: "<=", Value: int64(35)},
		},
	})
	if err != nil {
		t.Fatalf("Range query failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 employees aged 30-35, got %d", len(results))
	}

	// Test 3: IN operator
	results, err = client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "department", Operator: "in", Value: []any{"Sales", "Marketing"}},
		},
	})
	if err != nil {
		t.Fatalf("IN query failed: %v", err)
	}
	if len(results) != 6 {
		t.Errorf("Expected 6 Sales/Marketing employees, got %d", len(results))
	}

	// Test 4: BETWEEN operator
	results, err = client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "salary", Operator: "between", Value: [2]any{55000.0, 65000.0}},
		},
	})
	if err != nil {
		t.Fatalf("BETWEEN query failed: %v", err)
	}
	count := 0
	for _, r := range results {
		salary := r.GetAsFloat64("salary", 0)
		if salary >= 55000 && salary <= 65000 {
			count++
		}
	}
	if count != len(results) {
		t.Errorf("BETWEEN query returned incorrect results")
	}

	// Test 5: Complex query with multiple conditions
	results, err = client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "department", Operator: "==", Value: "Engineering"},
			{Column: "active", Operator: "==", Value: true},
			{Column: "salary", Operator: ">", Value: 55000.0},
		},
	})
	if err != nil {
		t.Fatalf("Complex query failed: %v", err)
	}
	// Should get Bob and Henry
	if len(results) != 2 {
		t.Errorf("Expected 2 results for complex query, got %d", len(results))
	}

	// Test 6: Pagination
	_, err = client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "active", Operator: "==", Value: true},
		},
	})
	if err != nil {
		t.Fatalf("Query for pagination test failed: %v", err)
	}

	page1, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "active", Operator: "==", Value: true},
		},
		Limit:  3,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Page 1 query failed: %v", err)
	}
	if len(page1) != 3 {
		t.Errorf("Expected 3 results in page 1, got %d", len(page1))
	}

	page2, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "active", Operator: "==", Value: true},
		},
		Limit:  3,
		Offset: 3,
	})
	if err != nil {
		t.Fatalf("Page 2 query failed: %v", err)
	}

	// Count total active records for verification
	allActive, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "active", Operator: "==", Value: true},
		},
	})
	if err != nil {
		t.Fatalf("Failed to query all active records: %v", err)
	}

	// Verify pagination results
	// Note: The order of records might vary between adapters,
	// so we just check that we got some results in each page
	if len(page1) == 0 {
		t.Errorf("Page 1 should have results")
	}
	if len(page2) == 0 && len(allActive) > 3 {
		t.Errorf("Page 2 should have results when total > page size")
	}

	// Verify that pagination doesn't lose records
	totalFromPages := len(page1) + len(page2)
	if totalFromPages > len(allActive) {
		t.Errorf("Pagination returned more records (%d) than total (%d)", totalFromPages, len(allActive))
	}
}

// testConcurrentOperations tests thread safety
func testConcurrentOperations(t *testing.T, client *sheetkv.Client) {
	clearAllRecords(t, client)

	// Number of goroutines and operations per goroutine
	numGoroutines := 10
	opsPerGoroutine := 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*opsPerGoroutine)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				record := &sheetkv.Record{
					Values: map[string]any{
						"routine_id": routineID,
						"op_id":      j,
						"value":      fmt.Sprintf("routine_%d_op_%d", routineID, j),
						"timestamp":  time.Now().UnixNano(),
					},
				}

				if err := client.Append(record); err != nil {
					errors <- fmt.Errorf("routine %d op %d: append failed: %w", routineID, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify all records were created
	allRecords, err := client.Query(sheetkv.Query{})
	if err != nil {
		t.Fatalf("Failed to query all records: %v", err)
	}

	expectedCount := numGoroutines * opsPerGoroutine
	if len(allRecords) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(allRecords))
	}

	// Verify data integrity
	recordMap := make(map[string]bool)
	for _, record := range allRecords {
		key := record.GetAsString("value", "")
		if recordMap[key] {
			t.Errorf("Duplicate record found: %s", key)
		}
		recordMap[key] = true
	}
}

// testLargeDataSet tests performance with larger datasets
func testLargeDataSet(t *testing.T, client *sheetkv.Client) {
	clearAllRecords(t, client)

	// Create a large dataset
	recordCount := 200
	departments := []string{"Engineering", "Sales", "Marketing", "HR", "Finance"}
	statuses := []string{"active", "inactive", "pending"}

	start := time.Now()

	for i := 1; i <= recordCount; i++ {
		record := &sheetkv.Record{
			Values: map[string]any{
				"id":         int64(i),
				"name":       fmt.Sprintf("Employee_%d", i),
				"email":      fmt.Sprintf("emp%d@company.com", i),
				"age":        int64(25 + (i % 40)),
				"salary":     50000.0 + float64(i*100),
				"department": departments[i%len(departments)],
				"status":     statuses[i%len(statuses)],
				"manager_id": int64((i / 10) + 1),
				"hire_date":  time.Now().AddDate(-(i % 10), 0, 0).Format(time.RFC3339),
			},
		}

		if err := client.Append(record); err != nil {
			t.Fatalf("Failed to append record %d: %v", i, err)
		}
	}

	insertTime := time.Since(start)
	t.Logf("Inserted %d records in %v", recordCount, insertTime)

	// Force sync to ensure all data is persisted
	syncStart := time.Now()
	if err := client.Sync(); err != nil {
		t.Fatalf("Failed to sync: %v", err)
	}
	t.Logf("Sync completed in %v", time.Since(syncStart))

	// Test 1: Department aggregation
	for _, dept := range departments {
		results, err := client.Query(sheetkv.Query{
			Conditions: []sheetkv.Condition{
				{Column: "department", Operator: "==", Value: dept},
			},
		})
		if err != nil {
			t.Fatalf("Query for %s failed: %v", dept, err)
		}
		expectedCount := recordCount / len(departments)
		if len(results) != expectedCount {
			t.Errorf("Expected %d records for %s, got %d", expectedCount, dept, len(results))
		}
	}

	// Test 2: Complex query performance
	queryStart := time.Now()
	results, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "age", Operator: ">=", Value: int64(30)},
			{Column: "age", Operator: "<=", Value: int64(40)},
			{Column: "department", Operator: "in", Value: []any{"Engineering", "Sales"}},
			{Column: "status", Operator: "==", Value: "active"},
		},
	})
	if err != nil {
		t.Fatalf("Complex query failed: %v", err)
	}
	t.Logf("Complex query returned %d results in %v", len(results), time.Since(queryStart))

	// Test 3: Pagination performance
	pageSize := 20
	totalPages := (recordCount + pageSize - 1) / pageSize

	for page := 0; page < 5 && page < totalPages; page++ { // Test first 5 pages
		pageStart := time.Now()
		pageResults, err := client.Query(sheetkv.Query{
			Limit:  pageSize,
			Offset: page * pageSize,
		})
		if err != nil {
			t.Fatalf("Page %d query failed: %v", page, err)
		}

		expectedSize := pageSize
		if page == totalPages-1 {
			expectedSize = recordCount % pageSize
			if expectedSize == 0 {
				expectedSize = pageSize
			}
		}

		if len(pageResults) != expectedSize {
			t.Errorf("Page %d: expected %d results, got %d", page, expectedSize, len(pageResults))
		}

		t.Logf("Page %d retrieved in %v", page, time.Since(pageStart))
	}

	// Test 4: Random updates
	updateCount := 20
	updateStart := time.Now()

	for i := 0; i < updateCount; i++ {
		// Random record between 2 and recordCount+1 (keys start at 2)
		key := 2 + rand.Intn(recordCount)

		updates := map[string]any{
			"last_review":       time.Now().Format(time.RFC3339),
			"performance_score": rand.Float64() * 100,
		}

		if err := client.Update(key, updates); err != nil {
			// Some keys might have been deleted in previous tests
			if err != sheetkv.ErrKeyNotFound {
				t.Errorf("Failed to update record %d: %v", key, err)
			}
		}
	}

	t.Logf("Updated %d records in %v", updateCount, time.Since(updateStart))
}

// testSyncStrategies tests gap-preserving and compacting sync strategies
func testSyncStrategies(t *testing.T, adapter sheetkv.Adapter) {
	ctx := context.Background()
	
	// Create two clients to test different sync behaviors
	config := &sheetkv.Config{
		SyncInterval:  30 * time.Second, // Long interval to control sync timing
		MaxRetries:    3,
		RetryInterval: 100 * time.Millisecond,
	}
	
	t.Run("API-Level Sync Strategy Testing", func(t *testing.T) {
		// Note: For sync strategy tests, we need to ensure clean state
		// Clear the adapter's data directly first
		if err := adapter.Save(ctx, []*sheetkv.Record{}, []string{}, sheetkv.SyncStrategyCompacting); err != nil {
			// Skip if Google Sheets sheet doesn't exist
			if strings.Contains(err.Error(), "Unable to parse range") || strings.Contains(err.Error(), "badRequest") {
				t.Skipf("Skipping test - sheet may not exist: %v", err)
			}
			t.Fatalf("Failed to clear adapter data: %v", err)
		}
		
		// Initialize client
		client := sheetkv.New(adapter, config)
		if err := client.Initialize(ctx); err != nil {
			t.Fatalf("Failed to initialize: %v", err)
		}
		
		// Create initial dataset
		initialData := []struct {
			name   string
			role   string
			salary float64
		}{
			{"Alice", "Engineer", 100000},
			{"Bob", "Manager", 120000},
			{"Charlie", "Designer", 90000},
			{"David", "Engineer", 95000},
			{"Eve", "Manager", 115000},
			{"Frank", "Designer", 85000},
			{"Grace", "Engineer", 105000},
			{"Henry", "Manager", 125000},
		}
		
		// Append all records and track keys
		keyMap := make(map[string]int)
		for _, data := range initialData {
			record := &sheetkv.Record{
				Values: map[string]any{
					"name":   data.name,
					"role":   data.role,
					"salary": data.salary,
				},
			}
			if err := client.Append(record); err != nil {
				t.Fatalf("Failed to append %s: %v", data.name, err)
			}
			keyMap[data.name] = record.Key
		}
		
		// Delete some records to create gaps
		toDelete := []string{"Bob", "David", "Frank"}
		for _, name := range toDelete {
			if key, ok := keyMap[name]; ok {
				if err := client.Delete(key); err != nil {
					t.Fatalf("Failed to delete %s: %v", name, err)
				}
			}
		}
		
		// Test 1: Gap-Preserving Sync (manual sync)
		t.Run("Manual Sync Preserves Gaps", func(t *testing.T) {
			// Perform manual sync (should use gap-preserving)
			if err := client.Sync(); err != nil {
				t.Fatalf("Manual sync failed: %v", err)
			}
			
			// Query all records to verify gaps
			allRecords, err := client.Query(sheetkv.Query{})
			if err != nil {
				t.Fatalf("Query failed: %v", err)
			}
			
			// Should still have records at original positions
			for _, r := range allRecords {
				switch r.Key {
				case keyMap["Alice"], keyMap["Charlie"], keyMap["Eve"], keyMap["Grace"], keyMap["Henry"]:
					// These should have data
					if name := r.GetAsString("name", ""); name == "" {
						t.Errorf("Expected data at key %d, but found empty", r.Key)
					}
				case keyMap["Bob"], keyMap["David"], keyMap["Frank"]:
					// These should be deleted (not in results)
					t.Errorf("Found deleted record at key %d", r.Key)
				}
			}
			
			// Verify highest key is still at original position
			if key := keyMap["Henry"]; key != 9 { // Henry should be at row 9 (8 records + header)
				t.Errorf("Expected Henry at row 9, but key is %d", key)
			}
		})
		
		// Test 2: Append after deletion maintains continuity
		t.Run("Append After Deletion", func(t *testing.T) {
			// Add new records after deletions
			newData := []struct {
				name   string
				role   string
				salary float64
			}{
				{"Iris", "Engineer", 98000},
				{"Jack", "Designer", 92000},
			}
			
			var newKeys []int
			for _, data := range newData {
				record := &sheetkv.Record{
					Values: map[string]any{
						"name":   data.name,
						"role":   data.role,
						"salary": data.salary,
					},
				}
				if err := client.Append(record); err != nil {
					t.Fatalf("Failed to append %s: %v", data.name, err)
				}
				newKeys = append(newKeys, record.Key)
			}
			
			// New records should continue from the highest key
			expectedFirstNewKey := 10 // After Henry at 9
			if newKeys[0] != expectedFirstNewKey {
				t.Errorf("Expected first new record at key %d, got %d", expectedFirstNewKey, newKeys[0])
			}
			if newKeys[1] != expectedFirstNewKey+1 {
				t.Errorf("Expected second new record at key %d, got %d", expectedFirstNewKey+1, newKeys[1])
			}
		})
		
		// Test 3: Close with compacting sync
		t.Run("Close Compacts Data", func(t *testing.T) {
			// Close the client (triggers compacting sync)
			if err := client.Close(); err != nil {
				t.Fatalf("Close failed: %v", err)
			}
			
			// Load directly from adapter to verify compacting
			records, _, err := adapter.Load(ctx)
			if err != nil {
				t.Fatalf("Failed to load after close: %v", err)
			}
			
			// Should have exactly 7 records (5 original - 3 deleted + 2 new)
			expectedCount := 7
			if len(records) != expectedCount {
				t.Errorf("Expected %d compacted records, got %d", expectedCount, len(records))
			}
			
			// Verify records are sequential from row 2
			expectedNames := []string{"Alice", "Charlie", "Eve", "Grace", "Henry", "Iris", "Jack"}
			for i, r := range records {
				if r.Key != i+2 { // Should be sequential starting from 2
					t.Errorf("Record %d: expected key %d, got %d", i, i+2, r.Key)
				}
				if i < len(expectedNames) {
					if name := r.GetAsString("name", ""); name != expectedNames[i] {
						t.Errorf("Record %d: expected name %s, got %s", i, expectedNames[i], name)
					}
				}
			}
			
			// Verify salary data is preserved
			for _, r := range records {
				salary := r.GetAsFloat64("salary", 0)
				if salary < 80000 || salary > 130000 {
					t.Errorf("Unexpected salary value: %f", salary)
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
