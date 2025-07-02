package googlesheets

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/ideamans/go-sheetkv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsAdaptor implements the Adapter interface for Google Sheets
type SheetsAdaptor struct {
	service       *sheets.Service
	spreadsheetID string
	sheetName     string
}

// NewSheetsAdaptor creates a new Google Sheets adaptor with provided options
func NewSheetsAdaptor(ctx context.Context, config Config, opts ...option.ClientOption) (*SheetsAdaptor, error) {
	service, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &SheetsAdaptor{
		service:       service,
		spreadsheetID: config.SpreadsheetID,
		sheetName:     config.SheetName,
	}, nil
}

// Load retrieves all records and schema from the spreadsheet
func (a *SheetsAdaptor) Load(ctx context.Context) ([]*sheetkv.Record, []string, error) {

	// Get all data from the sheet
	readRange := fmt.Sprintf("%s!A:ZZ", a.sheetName)
	resp, err := a.service.Spreadsheets.Values.Get(a.spreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sheet data: %w", err)
	}

	if len(resp.Values) == 0 {
		return []*sheetkv.Record{}, []string{}, nil
	}

	// First row is schema
	schema := make([]string, 0)
	if len(resp.Values) > 0 && len(resp.Values[0]) > 0 {
		for i := 0; i < len(resp.Values[0]); i++ {
			if col, ok := resp.Values[0][i].(string); ok && col != "" {
				schema = append(schema, col)
			}
		}
	}

	// Parse records from remaining rows
	records := make([]*sheetkv.Record, 0)
	for i := 1; i < len(resp.Values); i++ {
		row := resp.Values[i]
		if len(row) == 0 {
			continue
		}

		// Build record with row number as key (row 1 is header, so data starts at row 2)
		record := &sheetkv.Record{
			Key:    i + 1, // Row number (1-based, but data starts at row 2)
			Values: make(map[string]interface{}),
		}

		for j := 0; j < len(row) && j < len(schema); j++ {
			colName := schema[j]
			if colName != "" && row[j] != nil {
				record.Values[colName] = convertCellValue(row[j])
			}
		}

		records = append(records, record)
	}

	return records, schema, nil
}

// Save replaces all data in the spreadsheet with the provided records
func (a *SheetsAdaptor) Save(ctx context.Context, records []*sheetkv.Record, schema []string, strategy sheetkv.SyncStrategy) error {

	// Sort records by key (row number)
	sortedRecords := make([]*sheetkv.Record, len(records))
	copy(sortedRecords, records)
	sort.Slice(sortedRecords, func(i, j int) bool {
		return sortedRecords[i].Key < sortedRecords[j].Key
	})

	// Build values array
	values := make([][]interface{}, 0)

	// Header row (schema columns only)
	header := make([]interface{}, len(schema))
	for i, col := range schema {
		header[i] = col
	}
	values = append(values, header)

	// Data rows based on sync strategy
	if strategy == sheetkv.SyncStrategyGapPreserving {
		// Gap-preserving sync: maintain row numbers, use empty rows for deleted records
		currentRow := 2 // Start from row 2 (after header)

		for _, record := range sortedRecords {
			// Fill gaps with empty rows
			for currentRow < record.Key {
				emptyRow := make([]interface{}, len(schema))
				for i := range emptyRow {
					emptyRow[i] = ""
				}
				values = append(values, emptyRow)
				currentRow++
			}

			// Add the actual record
			row := make([]interface{}, len(schema))
			for i, col := range schema {
				if val, ok := record.Values[col]; ok {
					row[i] = convertToSheetValue(val)
				} else {
					row[i] = ""
				}
			}
			values = append(values, row)
			currentRow++
		}
	} else {
		// Compacting sync: remove gaps, compact all records
		for _, record := range sortedRecords {
			row := make([]interface{}, len(schema))
			for i, col := range schema {
				if val, ok := record.Values[col]; ok {
					row[i] = convertToSheetValue(val)
				} else {
					row[i] = ""
				}
			}
			values = append(values, row)
		}
	}

	// Clear the entire sheet first
	clearRange := fmt.Sprintf("%s!A:ZZ", a.sheetName)
	_, err := a.service.Spreadsheets.Values.Clear(a.spreadsheetID, clearRange, &sheets.ClearValuesRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to clear sheet: %w", err)
	}

	// Write all data
	writeRange := fmt.Sprintf("%s!A1", a.sheetName)
	vr := &sheets.ValueRange{
		Values: values,
	}
	_, err = a.service.Spreadsheets.Values.Update(a.spreadsheetID, writeRange, vr).
		ValueInputOption("RAW").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to update sheet: %w", err)
	}

	return nil
}

// BatchUpdate performs multiple operations in a single request
func (a *SheetsAdaptor) BatchUpdate(ctx context.Context, operations []sheetkv.Operation) error {
	// For simplicity, we'll load all data, apply operations, and save
	// In a production implementation, this could be optimized with batch API calls

	records, schema, err := a.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load data for batch update: %w", err)
	}

	// Convert to map for easier manipulation
	recordMap := make(map[int]*sheetkv.Record)
	for _, r := range records {
		recordMap[r.Key] = r
	}

	// Apply operations
	for _, op := range operations {
		switch op.Type {
		case sheetkv.OpAdd:
			if _, exists := recordMap[op.Record.Key]; exists {
				return fmt.Errorf("cannot add record with duplicate key: %d", op.Record.Key)
			}
			recordMap[op.Record.Key] = op.Record
			// Update schema if needed
			for col := range op.Record.Values {
				found := false
				for _, s := range schema {
					if s == col {
						found = true
						break
					}
				}
				if !found {
					schema = append(schema, col)
				}
			}

		case sheetkv.OpUpdate:
			if existing, exists := recordMap[op.Record.Key]; exists {
				// Merge values
				for k, v := range op.Record.Values {
					existing.Values[k] = v
				}
				// Update schema if needed
				for col := range op.Record.Values {
					found := false
					for _, s := range schema {
						if s == col {
							found = true
							break
						}
					}
					if !found {
						schema = append(schema, col)
					}
				}
			} else {
				return fmt.Errorf("cannot update non-existent record: %d", op.Record.Key)
			}

		case sheetkv.OpDelete:
			delete(recordMap, op.Record.Key)
		}
	}

	// Convert back to slice
	newRecords := make([]*sheetkv.Record, 0, len(recordMap))
	for _, r := range recordMap {
		newRecords = append(newRecords, r)
	}

	// Save all data (use gap-preserving strategy for batch updates)
	return a.Save(ctx, newRecords, schema, sheetkv.SyncStrategyGapPreserving)
}

// convertCellValue converts a Google Sheets cell value to Go type
func convertCellValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		// Try to parse as number
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		// Try to parse as bool
		if val == "true" || val == "TRUE" {
			return true
		}
		if val == "false" || val == "FALSE" {
			return false
		}
		return val
	case float64:
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return int64(val)
		}
		return val
	case bool:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// convertToSheetValue converts a Go value to Google Sheets cell value
func convertToSheetValue(v interface{}) interface{} {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprintf("%v", val)
	}
}
