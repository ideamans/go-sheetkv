package googlesheets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ideamans/go-sheetkv"
	"google.golang.org/api/option"
)

func TestSheetsAdaptor_Load(t *testing.T) {
	tests := []struct {
		name        string
		sheetData   string
		wantRecords []*sheetkv.Record
		wantSchema  []string
		wantErr     bool
	}{
		{
			name: "load with data",
			sheetData: `{
				"values": [
					["name", "age", "active"],
					["John Doe", "30", "true"],
					["Jane Smith", "25", "false"],
					["Bob Johnson", "35", "true"]
				]
			}`,
			wantRecords: []*sheetkv.Record{
				{
					Key: 2,
					Values: map[string]interface{}{
						"name":   "John Doe",
						"age":    int64(30),
						"active": true,
					},
				},
				{
					Key: 3,
					Values: map[string]interface{}{
						"name":   "Jane Smith",
						"age":    int64(25),
						"active": false,
					},
				},
				{
					Key: 4,
					Values: map[string]interface{}{
						"name":   "Bob Johnson",
						"age":    int64(35),
						"active": true,
					},
				},
			},
			wantSchema: []string{"name", "age", "active"},
			wantErr:    false,
		},
		{
			name: "load empty sheet",
			sheetData: `{
				"values": []
			}`,
			wantRecords: []*sheetkv.Record{},
			wantSchema:  []string{},
			wantErr:     false,
		},
		{
			name: "load schema only",
			sheetData: `{
				"values": [
					["name", "age"]
				]
			}`,
			wantRecords: []*sheetkv.Record{},
			wantSchema:  []string{"name", "age"},
			wantErr:     false,
		},
		{
			name: "skip empty rows",
			sheetData: `{
				"values": [
					["name"],
					["John"],
					[],
					["Jane"]
				]
			}`,
			wantRecords: []*sheetkv.Record{
				{
					Key:    2,
					Values: map[string]interface{}{"name": "John"},
				},
				{
					Key:    4,
					Values: map[string]interface{}{"name": "Jane"},
				},
			},
			wantSchema: []string{"name"},
			wantErr:    false,
		},
		{
			name: "handle all data rows",
			sheetData: `{
				"values": [
					["name"],
					["John"],
					["No Key"],
					["Jane"]
				]
			}`,
			wantRecords: []*sheetkv.Record{
				{
					Key:    2,
					Values: map[string]interface{}{"name": "John"},
				},
				{
					Key:    3,
					Values: map[string]interface{}{"name": "No Key"},
				},
				{
					Key:    4,
					Values: map[string]interface{}{"name": "Jane"},
				},
			},
			wantSchema: []string{"name"},
			wantErr:    false,
		},
		{
			name: "handle mixed types",
			sheetData: `{
				"values": [
					["score", "rating", "count"],
					["99.5", "4.5", "100"],
					[85.0, 4.0, 50.0]
				]
			}`,
			wantRecords: []*sheetkv.Record{
				{
					Key: 2,
					Values: map[string]interface{}{
						"score":  99.5,
						"rating": 4.5,
						"count":  int64(100),
					},
				},
				{
					Key: 3,
					Values: map[string]interface{}{
						"score":  int64(85),
						"rating": int64(4),
						"count":  int64(50),
					},
				},
			},
			wantSchema: []string{"score", "rating", "count"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP transport
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v4/spreadsheets/test-id/values/TestSheet!A:ZZ" {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(tt.sheetData))
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

			// Test Load
			gotRecords, gotSchema, err := adaptor.Load(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare schema
			if !reflect.DeepEqual(gotSchema, tt.wantSchema) {
				t.Errorf("Load() schema = %v, want %v", gotSchema, tt.wantSchema)
			}

			// Compare records
			if len(gotRecords) != len(tt.wantRecords) {
				t.Errorf("Load() returned %d records, want %d", len(gotRecords), len(tt.wantRecords))
				return
			}

			for i, got := range gotRecords {
				want := tt.wantRecords[i]
				if got.Key != want.Key {
					t.Errorf("Record[%d].Key = %v, want %v", i, got.Key, want.Key)
				}
				if !reflect.DeepEqual(got.Values, want.Values) {
					// Show detailed comparison for debugging
					for k, v := range want.Values {
						if gotVal, ok := got.Values[k]; !ok {
							t.Errorf("Record[%d].Values missing key %v", i, k)
						} else if !reflect.DeepEqual(gotVal, v) {
							t.Errorf("Record[%d].Values[%v] = %v (%T), want %v (%T)",
								i, k, gotVal, gotVal, v, v)
						}
					}
					for k, v := range got.Values {
						if _, ok := want.Values[k]; !ok {
							t.Errorf("Record[%d].Values has unexpected key %v = %v", i, k, v)
						}
					}
				}
			}
		})
	}
}

func TestSheetsAdaptor_Save(t *testing.T) {
	tests := []struct {
		name       string
		records    []*sheetkv.Record
		schema     []string
		wantClear  string
		wantUpdate string
		wantErr    bool
	}{
		{
			name: "save records",
			records: []*sheetkv.Record{
				{
					Key: 3,
					Values: map[string]interface{}{
						"name":   "Jane Smith",
						"age":    25,
						"active": false,
					},
				},
				{
					Key: 2,
					Values: map[string]interface{}{
						"name":   "John Doe",
						"age":    30,
						"active": true,
					},
				},
			},
			schema:     []string{"name", "age", "active"},
			wantClear:  "TestSheet!A:ZZ",
			wantUpdate: "TestSheet!A1",
			wantErr:    false,
		},
		{
			name:       "save empty data",
			records:    []*sheetkv.Record{},
			schema:     []string{"name", "age"},
			wantClear:  "TestSheet!A:ZZ",
			wantUpdate: "TestSheet!A1",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var clearedRange string
			var updatedRange string

			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v4/spreadsheets/test-id/values/TestSheet!A:ZZ:clear":
					clearedRange = "TestSheet!A:ZZ"
					w.Write([]byte(`{}`))
				case "/v4/spreadsheets/test-id/values/TestSheet!A1":
					updatedRange = "TestSheet!A1"
					// Parse request to get values
					// In real test, we would decode the request body
					// For simplicity, we'll just acknowledge
					w.Write([]byte(`{"updatedCells": 10}`))
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

			// Test Save (using gap-preserving for consistency)
			err = adaptor.Save(context.Background(), tt.records, tt.schema, sheetkv.SyncStrategyGapPreserving)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			if clearedRange != tt.wantClear {
				t.Errorf("Cleared range = %v, want %v", clearedRange, tt.wantClear)
			}

			if updatedRange != tt.wantUpdate {
				t.Errorf("Updated range = %v, want %v", updatedRange, tt.wantUpdate)
			}
		})
	}
}

func TestSheetsAdaptor_BatchUpdate(t *testing.T) {
	// Initial data for mock
	initialData := `{
		"values": [
			["name", "age"],
			["John", "30"],
			["Jane", "25"]
		]
	}`

	tests := []struct {
		name       string
		operations []sheetkv.Operation
		wantErr    bool
		errMsg     string
	}{
		{
			name: "add new record",
			operations: []sheetkv.Operation{
				{
					Type: sheetkv.OpAdd,
					Record: &sheetkv.Record{
						Key:    4,
						Values: map[string]interface{}{"name": "Bob", "age": 35},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "update existing record",
			operations: []sheetkv.Operation{
				{
					Type: sheetkv.OpUpdate,
					Record: &sheetkv.Record{
						Key:    2,
						Values: map[string]interface{}{"age": 31},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "delete record",
			operations: []sheetkv.Operation{
				{
					Type: sheetkv.OpDelete,
					Record: &sheetkv.Record{
						Key: 3,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "add duplicate key",
			operations: []sheetkv.Operation{
				{
					Type: sheetkv.OpAdd,
					Record: &sheetkv.Record{
						Key:    2,
						Values: map[string]interface{}{"name": "Duplicate"},
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate key",
		},
		{
			name: "update non-existent",
			operations: []sheetkv.Operation{
				{
					Type: sheetkv.OpUpdate,
					Record: &sheetkv.Record{
						Key:    999,
						Values: map[string]interface{}{"name": "Ghost"},
					},
				},
			},
			wantErr: true,
			errMsg:  "non-existent",
		},
		{
			name: "mixed operations",
			operations: []sheetkv.Operation{
				{
					Type: sheetkv.OpAdd,
					Record: &sheetkv.Record{
						Key:    4,
						Values: map[string]interface{}{"name": "Bob", "email": "bob@example.com"},
					},
				},
				{
					Type: sheetkv.OpUpdate,
					Record: &sheetkv.Record{
						Key:    2,
						Values: map[string]interface{}{"email": "john@example.com"},
					},
				},
				{
					Type: sheetkv.OpDelete,
					Record: &sheetkv.Record{
						Key: 3,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v4/spreadsheets/test-id/values/TestSheet!A:ZZ":
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(initialData))
				case "/v4/spreadsheets/test-id/values/TestSheet!A:ZZ:clear":
					w.Write([]byte(`{}`))
				case "/v4/spreadsheets/test-id/values/TestSheet!A1":
					w.Write([]byte(`{"updatedCells": 10}`))
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

			// Test BatchUpdate
			err = adaptor.BatchUpdate(context.Background(), tt.operations)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("BatchUpdate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestConvertCellValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{"string number to int64", "123", int64(123)},
		{"string float to float64", "123.45", 123.45},
		{"string true to bool", "true", true},
		{"string TRUE to bool", "TRUE", true},
		{"string false to bool", "false", false},
		{"string FALSE to bool", "FALSE", false},
		{"plain string", "hello", "hello"},
		{"float64 integer to int64", 100.0, int64(100)},
		{"float64 decimal", 100.5, 100.5},
		{"bool true", true, true},
		{"bool false", false, false},
		{"other type to string", []int{1, 2, 3}, "[1 2 3]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertCellValue(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertCellValue(%v) = %v (%T), want %v (%T)",
					tt.input, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestConvertToSheetValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{"nil to empty string", nil, ""},
		{"string", "hello", "hello"},
		{"int", 123, "123"},
		{"int64", int64(456), "456"},
		{"float64", 123.45, "123.45"},
		{"bool true", true, "TRUE"},
		{"bool false", false, "FALSE"},
		{"other type", []int{1, 2}, "[1 2]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToSheetValue(tt.input)
			if got != tt.want {
				t.Errorf("convertToSheetValue(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || (len(s) > len(substr) && contains(s[1:], substr)))
}
