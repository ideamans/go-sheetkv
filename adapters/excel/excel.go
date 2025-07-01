package excel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/ideamans/go-sheetkv"
	"github.com/xuri/excelize/v2"
)

// Adapter implements the sheetkv.Adapter interface for Excel files
type Adapter struct {
	config *Config
	mu     sync.RWMutex
}

// New creates a new Excel adapter with the given configuration
func New(config *Config) (*Adapter, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create a copy of config to avoid external modifications
	configCopy := *config

	return &Adapter{
		config: &configCopy,
	}, nil
}

// Load retrieves all records and schema from the Excel file
func (a *Adapter) Load(ctx context.Context) ([]*sheetkv.Record, []string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}

	// Open the Excel file
	f, err := excelize.OpenFile(a.config.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty data
			return []*sheetkv.Record{}, []string{}, nil
		}
		return nil, nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Check if sheet exists
	sheetIndex, err := f.GetSheetIndex(a.config.SheetName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sheet index: %w", err)
	}
	if sheetIndex == -1 {
		// Sheet doesn't exist, return empty data
		return []*sheetkv.Record{}, []string{}, nil
	}

	// Get all rows from the sheet
	rows, err := f.GetRows(a.config.SheetName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) == 0 {
		return []*sheetkv.Record{}, []string{}, nil
	}

	// First row is the schema
	schema := rows[0]

	// Convert rows to records
	records := make([]*sheetkv.Record, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue // Skip empty rows
		}

		record := &sheetkv.Record{
			Key:    i + 1, // Row number (1-based, but data starts from row 2)
			Values: make(map[string]interface{}),
		}

		// Map values to schema columns
		for j, value := range row {
			if j < len(schema) && schema[j] != "" {
				// Try to parse as number first
				if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					// Check if it's an integer
					if intVal := int64(floatVal); float64(intVal) == floatVal {
						record.Values[schema[j]] = intVal
					} else {
						record.Values[schema[j]] = floatVal
					}
				} else if value == "true" || value == "false" || value == "TRUE" || value == "FALSE" {
					record.Values[schema[j]] = (value == "true" || value == "TRUE")
				} else {
					record.Values[schema[j]] = value
				}
			}
		}

		records = append(records, record)
	}

	return records, schema, nil
}

// Save replaces all data in the Excel file with the provided records
func (a *Adapter) Save(ctx context.Context, records []*sheetkv.Record, schema []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(a.config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a new Excel file or open existing one
	var f *excelize.File
	if _, err := os.Stat(a.config.FilePath); err == nil {
		// File exists, open it
		f, err = excelize.OpenFile(a.config.FilePath)
		if err != nil {
			return fmt.Errorf("failed to open Excel file: %w", err)
		}
	} else {
		// File doesn't exist, create new
		f = excelize.NewFile()
	}
	defer f.Close()

	// Check if sheet exists, create if not
	sheetIndex, err := f.GetSheetIndex(a.config.SheetName)
	if err != nil {
		return fmt.Errorf("failed to get sheet index: %w", err)
	}

	if sheetIndex == -1 {
		// Create new sheet
		index, err := f.NewSheet(a.config.SheetName)
		if err != nil {
			return fmt.Errorf("failed to create sheet: %w", err)
		}
		f.SetActiveSheet(index)

		// Delete default sheet if it exists and is not our sheet
		if defaultSheet := f.GetSheetName(0); defaultSheet != a.config.SheetName {
			_ = f.DeleteSheet(defaultSheet) // Ignore error - not critical
		}
	} else {
		// Clear existing sheet
		// Get the dimensions of the sheet
		rows, err := f.GetRows(a.config.SheetName)
		if err == nil && len(rows) > 0 {
			// Clear all cells
			maxCol := 0
			for _, row := range rows {
				if len(row) > maxCol {
					maxCol = len(row)
				}
			}

			// Clear the range
			if maxCol > 0 && len(rows) > 0 {
				// Note: excelize doesn't have a direct "clear range" method,
				// so we'll overwrite with our new data
				_ = f.SetSheetRow(a.config.SheetName, "A1", &[]interface{}{}) // Best effort clear
			}
		}
	}

	// Write schema (header row)
	headerValues := make([]interface{}, len(schema))
	for i, col := range schema {
		headerValues[i] = col
	}

	cell := "A1"
	if err := f.SetSheetRow(a.config.SheetName, cell, &headerValues); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write records
	for _, record := range records {
		rowNum := record.Key
		if rowNum < 2 {
			rowNum = 2 // Ensure we don't overwrite header
		}

		rowValues := make([]interface{}, len(schema))
		for i, col := range schema {
			if val, ok := record.Values[col]; ok {
				rowValues[i] = val
			} else {
				rowValues[i] = ""
			}
		}

		cell := fmt.Sprintf("A%d", rowNum)
		if err := f.SetSheetRow(a.config.SheetName, cell, &rowValues); err != nil {
			return fmt.Errorf("failed to write row %d: %w", rowNum, err)
		}
	}

	// Save the file
	if err := f.SaveAs(a.config.FilePath); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}

	return nil
}

// BatchUpdate performs multiple operations in a single request
func (a *Adapter) BatchUpdate(ctx context.Context, operations []sheetkv.Operation) error {
	// For Excel, we need to load all data, apply operations, and save back
	records, schema, err := a.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load data for batch update: %w", err)
	}

	// Convert to map for easier manipulation
	recordMap := make(map[int]*sheetkv.Record)
	for _, record := range records {
		recordMap[record.Key] = record
	}

	// Apply operations
	for _, op := range operations {
		switch op.Type {
		case sheetkv.OpAdd:
			if op.Record != nil {
				// Find next available key if not specified
				if op.Record.Key == 0 {
					maxKey := 1
					for key := range recordMap {
						if key > maxKey {
							maxKey = key
						}
					}
					op.Record.Key = maxKey + 1
				}
				recordMap[op.Record.Key] = op.Record

				// Update schema if new columns exist
				for col := range op.Record.Values {
					found := false
					for _, existingCol := range schema {
						if existingCol == col {
							found = true
							break
						}
					}
					if !found {
						schema = append(schema, col)
					}
				}
			}

		case sheetkv.OpUpdate:
			if op.Record != nil && op.Record.Key > 0 {
				if existing, ok := recordMap[op.Record.Key]; ok {
					// Update existing record
					for k, v := range op.Record.Values {
						existing.Values[k] = v
					}
				} else {
					// Add as new record if doesn't exist
					recordMap[op.Record.Key] = op.Record
				}

				// Update schema if new columns exist
				for col := range op.Record.Values {
					found := false
					for _, existingCol := range schema {
						if existingCol == col {
							found = true
							break
						}
					}
					if !found {
						schema = append(schema, col)
					}
				}
			}

		case sheetkv.OpDelete:
			if op.Record != nil && op.Record.Key > 0 {
				delete(recordMap, op.Record.Key)
			}
		}
	}

	// Convert back to slice
	newRecords := make([]*sheetkv.Record, 0, len(recordMap))
	for _, record := range recordMap {
		newRecords = append(newRecords, record)
	}

	// Save the updated data
	return a.Save(ctx, newRecords, schema)
}

// columnName converts a column number to Excel column name (1 -> A, 26 -> Z, 27 -> AA)
func columnName(col int) string {
	result := ""
	for col > 0 {
		col--
		result = string(rune('A'+col%26)) + result
		col /= 26
	}
	return result
}
