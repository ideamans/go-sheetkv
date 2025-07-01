package sheetkv

import "time"

// Config represents configuration for the KVS client
type Config struct {
	SyncInterval  time.Duration // Interval for periodic sync (default: 30s)
	MaxRetries    int           // Maximum number of retries for API calls (default: 3)
	RetryInterval time.Duration // Base interval between retries for exponential backoff (default: 1s)
}
