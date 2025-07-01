package sheetkv

import "context"

// OperationType represents the type of operation
type OperationType int

const (
	OpAdd OperationType = iota
	OpUpdate
	OpDelete
)

// Operation represents a single data operation
type Operation struct {
	Type   OperationType
	Record *Record
}

// Adapter interface defines methods for interacting with different spreadsheet backends
type Adapter interface {
	// Load retrieves all records and schema from the spreadsheet
	Load(ctx context.Context) ([]*Record, []string, error)

	// Save replaces all data in the spreadsheet with the provided records
	Save(ctx context.Context, records []*Record, schema []string) error

	// BatchUpdate performs multiple operations in a single request
	BatchUpdate(ctx context.Context, operations []Operation) error
}
