package googlesheets

import (
	"time"

	sheetkv "github.com/ideamans/go-sheetkv"
)

// Config represents configuration specific to Google Sheets adapter
type Config struct {
	SpreadsheetID string
	SheetName     string
}

// DefaultClientConfig returns the recommended default configuration for Google Sheets
func DefaultClientConfig() *sheetkv.Config {
	return &sheetkv.Config{
		SyncInterval:  10 * time.Second,
		MaxRetries:    3,
		RetryInterval: 20 * time.Second,
	}
}
