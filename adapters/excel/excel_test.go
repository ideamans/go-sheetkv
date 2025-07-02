package excel

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ideamans/go-sheetkv"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				FilePath:  "test.xlsx",
				SheetName: "Sheet1",
			},
			wantErr: false,
		},
		{
			name: "missing file path",
			config: &Config{
				SheetName: "Sheet1",
			},
			wantErr: true,
		},
		{
			name: "missing sheet name",
			config: &Config{
				FilePath: "test.xlsx",
			},
			wantErr: true,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdapter_LoadSave(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "excel-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.xlsx")

	config := &Config{
		FilePath:  testFile,
		SheetName: "TestSheet",
	}

	adapter, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()

	t.Run("Load non-existent file", func(t *testing.T) {
		records, schema, err := adapter.Load(ctx)
		if err != nil {
			t.Errorf("Load() error = %v, want nil", err)
		}
		if len(records) != 0 {
			t.Errorf("Load() got %d records, want 0", len(records))
		}
		if len(schema) != 0 {
			t.Errorf("Load() got %d schema columns, want 0", len(schema))
		}
	})

	t.Run("Save and Load", func(t *testing.T) {
		// Prepare test data
		schema := []string{"id", "name", "age", "active"}
		records := []*sheetkv.Record{
			{
				Key: 2,
				Values: map[string]interface{}{
					"id":     int64(1),
					"name":   "Alice",
					"age":    int64(30),
					"active": true,
				},
			},
			{
				Key: 3,
				Values: map[string]interface{}{
					"id":     int64(2),
					"name":   "Bob",
					"age":    int64(25),
					"active": false,
				},
			},
		}

		// Save data (using gap-preserving for test)
		err := adapter.Save(ctx, records, schema, sheetkv.SyncStrategyGapPreserving)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Errorf("Excel file was not created")
		}

		// Load data back
		loadedRecords, loadedSchema, err := adapter.Load(ctx)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Verify schema
		if len(loadedSchema) != len(schema) {
			t.Errorf("Load() got %d schema columns, want %d", len(loadedSchema), len(schema))
		}
		for i, col := range schema {
			if i < len(loadedSchema) && loadedSchema[i] != col {
				t.Errorf("Schema column %d = %s, want %s", i, loadedSchema[i], col)
			}
		}

		// Verify records
		if len(loadedRecords) != len(records) {
			t.Errorf("Load() got %d records, want %d", len(loadedRecords), len(records))
		}

		// Check first record
		if len(loadedRecords) > 0 {
			record := loadedRecords[0]
			if record.Key != 2 {
				t.Errorf("First record key = %d, want 2", record.Key)
			}
			if name, ok := record.Values["name"].(string); !ok || name != "Alice" {
				t.Errorf("First record name = %v, want Alice", record.Values["name"])
			}
			if age, ok := record.Values["age"].(int64); !ok || age != 30 {
				t.Errorf("First record age = %v, want 30", record.Values["age"])
			}
			if active, ok := record.Values["active"].(bool); !ok || active != true {
				t.Errorf("First record active = %v, want true", record.Values["active"])
			}
		}
	})

	t.Run("Save empty data", func(t *testing.T) {
		err := adapter.Save(ctx, []*sheetkv.Record{}, []string{}, sheetkv.SyncStrategyGapPreserving)
		if err != nil {
			t.Errorf("Save() with empty data error = %v", err)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, _, err := adapter.Load(cancelCtx)
		if err == nil {
			t.Errorf("Load() with cancelled context should return error")
		}

		err = adapter.Save(cancelCtx, []*sheetkv.Record{}, []string{}, sheetkv.SyncStrategyGapPreserving)
		if err == nil {
			t.Errorf("Save() with cancelled context should return error")
		}
	})
}

func TestAdapter_BatchUpdate(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "excel-batch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "batch_test.xlsx")

	config := &Config{
		FilePath:  testFile,
		SheetName: "BatchTest",
	}

	adapter, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()

	// Initial data
	schema := []string{"id", "name"}
	records := []*sheetkv.Record{
		{
			Key: 2,
			Values: map[string]interface{}{
				"id":   int64(1),
				"name": "Initial",
			},
		},
	}

	err = adapter.Save(ctx, records, schema, sheetkv.SyncStrategyGapPreserving)
	if err != nil {
		t.Fatalf("Failed to save initial data: %v", err)
	}

	t.Run("Add, Update, Delete operations", func(t *testing.T) {
		operations := []sheetkv.Operation{
			// Add new record
			{
				Type: sheetkv.OpAdd,
				Record: &sheetkv.Record{
					Key: 3,
					Values: map[string]interface{}{
						"id":         int64(2),
						"name":       "Added",
						"new_column": "value",
					},
				},
			},
			// Update existing record
			{
				Type: sheetkv.OpUpdate,
				Record: &sheetkv.Record{
					Key: 2,
					Values: map[string]interface{}{
						"name": "Updated",
					},
				},
			},
			// Delete a record (we'll add one first)
			{
				Type: sheetkv.OpAdd,
				Record: &sheetkv.Record{
					Key: 4,
					Values: map[string]interface{}{
						"id":   int64(3),
						"name": "ToDelete",
					},
				},
			},
			{
				Type: sheetkv.OpDelete,
				Record: &sheetkv.Record{
					Key: 4,
				},
			},
		}

		err := adapter.BatchUpdate(ctx, operations)
		if err != nil {
			t.Fatalf("BatchUpdate() error = %v", err)
		}

		// Verify results
		loadedRecords, loadedSchema, err := adapter.Load(ctx)
		if err != nil {
			t.Fatalf("Load() after batch update error = %v", err)
		}

		// Should have 2 records (initial updated + added)
		if len(loadedRecords) != 2 {
			t.Errorf("Got %d records, want 2", len(loadedRecords))
		}

		// Check schema includes new column
		hasNewColumn := false
		for _, col := range loadedSchema {
			if col == "new_column" {
				hasNewColumn = true
				break
			}
		}
		if !hasNewColumn {
			t.Errorf("Schema doesn't include new_column")
		}

		// Verify updated record
		for _, record := range loadedRecords {
			if record.Key == 2 {
				if name, ok := record.Values["name"].(string); !ok || name != "Updated" {
					t.Errorf("Updated record name = %v, want Updated", record.Values["name"])
				}
			}
			if record.Key == 3 {
				if name, ok := record.Values["name"].(string); !ok || name != "Added" {
					t.Errorf("Added record name = %v, want Added", record.Values["name"])
				}
			}
			if record.Key == 4 {
				t.Errorf("Deleted record still exists")
			}
		}
	})
}

func TestAdapter_SyncStrategies(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "excel-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ctx := context.Background()

	t.Run("GapPreserving Strategy", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "gap_test.xlsx")
		config := &Config{
			FilePath:  testFile,
			SheetName: "GapTest",
		}
		adapter, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create adapter: %v", err)
		}

		// Create records with gaps (deleted records)
		schema := []string{"id", "name"}
		records := []*sheetkv.Record{
			{
				Key: 2,
				Values: map[string]interface{}{
					"id":   int64(1),
					"name": "First",
				},
			},
			// Gap at row 3 (deleted)
			{
				Key: 4,
				Values: map[string]interface{}{
					"id":   int64(3),
					"name": "Third",
				},
			},
			// Gap at row 5 (deleted)
			{
				Key: 6,
				Values: map[string]interface{}{
					"id":   int64(5),
					"name": "Fifth",
				},
			},
		}

		// Save with gap-preserving strategy
		err = adapter.Save(ctx, records, schema, sheetkv.SyncStrategyGapPreserving)
		if err != nil {
			t.Fatalf("Save with gap-preserving error = %v", err)
		}

		// Load and verify gaps are preserved
		loaded, _, err := adapter.Load(ctx)
		if err != nil {
			t.Fatalf("Load error = %v", err)
		}

		// Should have 5 records (including empty rows)
		if len(loaded) != 5 {
			t.Errorf("Got %d records, want 5 (including gaps)", len(loaded))
		}

		// Verify key positions
		for _, r := range loaded {
			switch r.Key {
			case 2:
				if name := r.GetAsString("name", ""); name != "First" {
					t.Errorf("Row 2 name = %s, want First", name)
				}
			case 3:
				// Should be empty
				if name := r.GetAsString("name", ""); name != "" {
					t.Errorf("Row 3 should be empty, got name = %s", name)
				}
			case 4:
				if name := r.GetAsString("name", ""); name != "Third" {
					t.Errorf("Row 4 name = %s, want Third", name)
				}
			case 5:
				// Should be empty
				if name := r.GetAsString("name", ""); name != "" {
					t.Errorf("Row 5 should be empty, got name = %s", name)
				}
			case 6:
				if name := r.GetAsString("name", ""); name != "Fifth" {
					t.Errorf("Row 6 name = %s, want Fifth", name)
				}
			}
		}
	})

	t.Run("Compacting Strategy", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "compact_test.xlsx")
		config := &Config{
			FilePath:  testFile,
			SheetName: "CompactTest",
		}
		adapter, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create adapter: %v", err)
		}

		// First, save some data with gaps using gap-preserving
		schema := []string{"id", "name"}
		recordsWithGaps := []*sheetkv.Record{
			{
				Key: 2,
				Values: map[string]interface{}{
					"id":   int64(1),
					"name": "First",
				},
			},
			// Gap at row 3
			{
				Key: 4,
				Values: map[string]interface{}{
					"id":   int64(3),
					"name": "Third",
				},
			},
			// Gap at row 5
			{
				Key: 6,
				Values: map[string]interface{}{
					"id":   int64(5),
					"name": "Fifth",
				},
			},
			// Extra rows to verify trailing cleanup
			{
				Key: 10,
				Values: map[string]interface{}{
					"id":   int64(10),
					"name": "Tenth",
				},
			},
		}

		// First save with gaps
		err = adapter.Save(ctx, recordsWithGaps, schema, sheetkv.SyncStrategyGapPreserving)
		if err != nil {
			t.Fatalf("Initial save error = %v", err)
		}

		// Now save compacted data (without the 10th record)
		compactRecords := recordsWithGaps[:3] // Only first 3 records

		// Save with compacting strategy
		err = adapter.Save(ctx, compactRecords, schema, sheetkv.SyncStrategyCompacting)
		if err != nil {
			t.Fatalf("Save with compacting error = %v", err)
		}

		// Load and verify data is compacted
		loaded, _, err := adapter.Load(ctx)
		if err != nil {
			t.Fatalf("Load error = %v", err)
		}

		// Should have exactly 3 records (no gaps)
		if len(loaded) != 3 {
			t.Errorf("Got %d records, want 3 (compacted)", len(loaded))
		}

		// Verify records are sequential starting from row 2
		expectedKeys := []int{2, 3, 4}
		expectedNames := []string{"First", "Third", "Fifth"}

		for i, r := range loaded {
			if r.Key != expectedKeys[i] {
				t.Errorf("Record %d key = %d, want %d", i, r.Key, expectedKeys[i])
			}
			if name := r.GetAsString("name", ""); name != expectedNames[i] {
				t.Errorf("Record %d name = %s, want %s", i, name, expectedNames[i])
			}
		}

		// Verify no trailing data
		for _, r := range loaded {
			if r.Key > 4 {
				t.Errorf("Found unexpected record at row %d after compacting", r.Key)
			}
		}
	})
}

func TestColumnName(t *testing.T) {
	tests := []struct {
		col  int
		want string
	}{
		{1, "A"},
		{26, "Z"},
		{27, "AA"},
		{52, "AZ"},
		{53, "BA"},
		{702, "ZZ"},
		{703, "AAA"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := columnName(tt.col)
			if got != tt.want {
				t.Errorf("columnName(%d) = %s, want %s", tt.col, got, tt.want)
			}
		})
	}
}
