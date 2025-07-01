package sheetkv

import "errors"

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrDuplicateKey  = errors.New("duplicate key")
	ErrSyncFailed    = errors.New("sync failed")
	ErrQuotaExceeded = errors.New("quota exceeded")
)
