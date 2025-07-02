package googlesheets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ideamans/go-sheetkv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func TestSheetsAdaptor_SyncStrategies(t *testing.T) {
	ctx := context.Background()

	t.Run("GapPreserving Strategy", func(t *testing.T) {
		var savedValues [][]interface{}

		// Mock server to capture the save request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/v4/spreadsheets/test-sheet-id/values:batchGet":
				// Initial empty sheet
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"valueRanges": []map[string]interface{}{
						{"values": [][]interface{}{}},
					},
				})
			case "/v4/spreadsheets/test-sheet-id/values/TestSheet!A:ZZ:clear":
				// Clear request
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{})
			case "/v4/spreadsheets/test-sheet-id/values/TestSheet!A1":
				// Capture the values being saved
				var req sheets.ValueRange
				json.NewDecoder(r.Body).Decode(&req)
				savedValues = req.Values

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"updatedCells": len(savedValues) * len(savedValues[0]),
				})
			default:
				t.Errorf("Unexpected request to %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		// Create adapter with mock server
		adapter := &SheetsAdaptor{
			spreadsheetID: "test-sheet-id",
			sheetName:     "TestSheet",
		}

		service, err := sheets.NewService(ctx, option.WithHTTPClient(server.Client()), option.WithEndpoint(server.URL))
		if err != nil {
			t.Fatalf("Failed to create sheets service: %v", err)
		}
		adapter.service = service

		// Test data with gaps
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
			t.Fatalf("Save error = %v", err)
		}

		// Verify the saved data has gaps
		expectedRows := 6 // Header + 5 data rows (including gaps)
		if len(savedValues) != expectedRows {
			t.Errorf("Saved %d rows, want %d", len(savedValues), expectedRows)
		}

		// Check header
		if len(savedValues) > 0 && !reflect.DeepEqual(savedValues[0], []interface{}{"id", "name"}) {
			t.Errorf("Header = %v, want [id name]", savedValues[0])
		}

		// Check data rows with gaps
		if len(savedValues) > 1 {
			// Row 2 (index 1) should have data
			if !reflect.DeepEqual(savedValues[1], []interface{}{"1", "First"}) {
				t.Errorf("Row 2 = %v, want [1 First]", savedValues[1])
			}
		}
		if len(savedValues) > 2 {
			// Row 3 (index 2) should be empty
			if !reflect.DeepEqual(savedValues[2], []interface{}{"", ""}) {
				t.Errorf("Row 3 = %v, want empty row", savedValues[2])
			}
		}
		if len(savedValues) > 3 {
			// Row 4 (index 3) should have data
			if !reflect.DeepEqual(savedValues[3], []interface{}{"3", "Third"}) {
				t.Errorf("Row 4 = %v, want [3 Third]", savedValues[3])
			}
		}
		if len(savedValues) > 4 {
			// Row 5 (index 4) should be empty
			if !reflect.DeepEqual(savedValues[4], []interface{}{"", ""}) {
				t.Errorf("Row 5 = %v, want empty row", savedValues[4])
			}
		}
		if len(savedValues) > 5 {
			// Row 6 (index 5) should have data
			if !reflect.DeepEqual(savedValues[5], []interface{}{"5", "Fifth"}) {
				t.Errorf("Row 6 = %v, want [5 Fifth]", savedValues[5])
			}
		}
	})

	t.Run("Compacting Strategy", func(t *testing.T) {
		var savedValues [][]interface{}

		// Mock server to capture the save request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/v4/spreadsheets/test-sheet-id/values:batchGet":
				// Initial empty sheet
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"valueRanges": []map[string]interface{}{
						{"values": [][]interface{}{}},
					},
				})
			case "/v4/spreadsheets/test-sheet-id/values/TestSheet!A:ZZ:clear":
				// Clear request
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{})
			case "/v4/spreadsheets/test-sheet-id/values/TestSheet!A1":
				// Capture the values being saved
				var req sheets.ValueRange
				json.NewDecoder(r.Body).Decode(&req)
				savedValues = req.Values

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"updatedCells": len(savedValues) * len(savedValues[0]),
				})
			default:
				t.Errorf("Unexpected request to %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		// Create adapter with mock server
		adapter := &SheetsAdaptor{
			spreadsheetID: "test-sheet-id",
			sheetName:     "TestSheet",
		}

		service, err := sheets.NewService(ctx, option.WithHTTPClient(server.Client()), option.WithEndpoint(server.URL))
		if err != nil {
			t.Fatalf("Failed to create sheets service: %v", err)
		}
		adapter.service = service

		// Test data with gaps
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

		// Save with compacting strategy
		err = adapter.Save(ctx, records, schema, sheetkv.SyncStrategyCompacting)
		if err != nil {
			t.Fatalf("Save error = %v", err)
		}

		// Verify the saved data is compacted
		expectedRows := 4 // Header + 3 data rows (no gaps)
		if len(savedValues) != expectedRows {
			t.Errorf("Saved %d rows, want %d", len(savedValues), expectedRows)
		}

		// Check header
		if len(savedValues) > 0 && !reflect.DeepEqual(savedValues[0], []interface{}{"id", "name"}) {
			t.Errorf("Header = %v, want [id name]", savedValues[0])
		}

		// Check data rows are compacted (no gaps)
		if len(savedValues) > 1 {
			// Row 2 (index 1) should have first record
			if !reflect.DeepEqual(savedValues[1], []interface{}{"1", "First"}) {
				t.Errorf("Row 2 = %v, want [1 First]", savedValues[1])
			}
		}
		if len(savedValues) > 2 {
			// Row 3 (index 2) should have second record (no gap)
			if !reflect.DeepEqual(savedValues[2], []interface{}{"3", "Third"}) {
				t.Errorf("Row 3 = %v, want [3 Third]", savedValues[2])
			}
		}
		if len(savedValues) > 3 {
			// Row 4 (index 3) should have third record (no gap)
			if !reflect.DeepEqual(savedValues[3], []interface{}{"5", "Fifth"}) {
				t.Errorf("Row 4 = %v, want [5 Fifth]", savedValues[3])
			}
		}
	})
}
