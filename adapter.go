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

// SyncStrategy represents the synchronization strategy
type SyncStrategy int

const (
	// SyncStrategyGapPreserving maintains deleted rows as empty rows to preserve row numbers
	SyncStrategyGapPreserving SyncStrategy = iota
	// SyncStrategyCompacting removes deleted rows and compacts the data
	SyncStrategyCompacting
)

// Adapter interface defines methods for interacting with different spreadsheet backends
type Adapter interface {
	// Load retrieves all records and schema from the spreadsheet
	Load(ctx context.Context) ([]*Record, []string, error)

	// Save replaces all data in the spreadsheet with the provided records
	// The strategy parameter determines how deleted records are handled
	Save(ctx context.Context, records []*Record, schema []string, strategy SyncStrategy) error

	// BatchUpdate performs multiple operations in a single request
	BatchUpdate(ctx context.Context, operations []Operation) error
}
