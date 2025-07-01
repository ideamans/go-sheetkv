package excel

import (
	"time"

	sheetkv "github.com/ideamans/go-sheetkv"
)

// Config holds configuration for Excel adapter
type Config struct {
	FilePath  string // Path to the Excel file
	SheetName string // Name of the sheet to use
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.FilePath == "" {
		return ErrMissingFilePath
	}
	if c.SheetName == "" {
		return ErrMissingSheetName
	}
	return nil
}

// DefaultClientConfig returns the recommended default configuration for Excel
func DefaultClientConfig() *sheetkv.Config {
	return &sheetkv.Config{
		SyncInterval:  1 * time.Second,
		MaxRetries:    3,
		RetryInterval: 5 * time.Second,
	}
}
