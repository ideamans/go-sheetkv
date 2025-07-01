package sheetkv

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Client is the main KVS client
type Client struct {
	config      Config
	cache       *Cache
	adaptor     Adapter
	syncManager *SyncManager
	mu          sync.Mutex
	closed      bool
}

// New creates a new KVS client with the given adapter and configuration
func New(adapter Adapter, config *Config) *Client {
	// Use default config if not provided
	if config == nil {
		config = &Config{
			SyncInterval:  30 * time.Second,
			MaxRetries:    3,
			RetryInterval: 1 * time.Second,
		}
	}

	// Set defaults for zero values
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.RetryInterval <= 0 {
		config.RetryInterval = 1 * time.Second
	}

	cache := NewCache()

	client := &Client{
		config:  *config,
		cache:   cache,
		adaptor: adapter,
	}

	// Note: Initial data loading is done lazily or can be done explicitly
	// to avoid error in constructor. This matches the new API design.

	// Start sync manager if interval is specified
	if config.SyncInterval > 0 {
		client.syncManager = NewSyncManager(client, config.SyncInterval)
		client.syncManager.Start()
	}

	return client
}

// Initialize loads initial data from the adapter
func (c *Client) Initialize(ctx context.Context) error {
	return c.loadFromAdapter(ctx)
}

// loadFromAdapter loads data from the adaptor with retry logic
func (c *Client) loadFromAdapter(ctx context.Context) error {
	var records []*Record
	var schema []string
	var err error

	for i := 0; i <= c.config.MaxRetries; i++ {
		records, schema, err = c.adaptor.Load(ctx)
		if err == nil {
			break
		}

		if i < c.config.MaxRetries {
			// Exponential backoff with reasonable limits
			backoff := time.Duration(1<<uint(i)) * 100 * time.Millisecond
			if backoff > 2*time.Second {
				backoff = 2 * time.Second
			}
			time.Sleep(backoff)
		}
	}

	if err != nil {
		return fmt.Errorf("failed after %d retries: %w", c.config.MaxRetries, err)
	}

	c.cache.Load(records, schema)
	return nil
}

// saveToAdapter saves data to the adaptor with retry logic
func (c *Client) saveToAdapter(ctx context.Context) error {
	// Check if there's any dirty data to save
	dirtyKeys := c.cache.GetDirtyKeys()
	if len(dirtyKeys) == 0 {
		return nil // Nothing to save
	}

	records := c.cache.GetAllRecords()
	schema := c.cache.GetSchema()

	var err error
	for i := 0; i <= c.config.MaxRetries; i++ {
		err = c.adaptor.Save(ctx, records, schema)
		if err == nil {
			c.cache.ClearDirty()
			return nil
		}

		if i < c.config.MaxRetries {
			// Exponential backoff with reasonable limits
			backoff := time.Duration(1<<uint(i)) * 100 * time.Millisecond
			if backoff > 2*time.Second {
				backoff = 2 * time.Second
			}
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("failed after %d retries: %w", c.config.MaxRetries, err)
}

// Get retrieves a record by key
func (c *Client) Get(key int) (*Record, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, fmt.Errorf("client is closed")
	}

	return c.cache.Get(key)
}

// Set stores or updates a record
func (c *Client) Set(key int, record *Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	return c.cache.Set(key, record)
}

// Append adds a new record
func (c *Client) Append(record *Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	// Find the next available key (row number)
	maxKey := 1 // Start from row 2 (row 1 is header)
	for _, r := range c.cache.GetAllRecords() {
		if r.Key > maxKey {
			maxKey = r.Key
		}
	}

	record.Key = maxKey + 1
	return c.cache.Append(record)
}

// Update partially updates a record
func (c *Client) Update(key int, updates map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	return c.cache.Update(key, updates)
}

// Delete removes a record
func (c *Client) Delete(key int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	return c.cache.Delete(key)
}

// Query searches for records matching the given conditions
func (c *Client) Query(query Query) ([]*Record, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, fmt.Errorf("client is closed")
	}

	return c.cache.Query(query)
}

// Sync forces synchronization with the backend
func (c *Client) Sync() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	return c.saveToAdapter(context.Background())
}

// Close closes the client and ensures final sync
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}

	// Mark as closed to prevent new operations
	c.closed = true
	syncManager := c.syncManager
	c.syncManager = nil
	c.mu.Unlock()

	// Stop the sync manager if running (without holding the mutex)
	if syncManager != nil {
		syncManager.Stop()
	}

	// Perform final sync (without holding the mutex)
	if err := c.saveToAdapter(context.Background()); err != nil {
		return fmt.Errorf("failed to sync on close: %w", err)
	}

	return nil
}

// SyncManager manages periodic synchronization
type SyncManager struct {
	client    *Client
	interval  time.Duration
	ticker    *time.Ticker
	done      chan bool
	syncMutex sync.Mutex
	syncing   bool
	wg        sync.WaitGroup
}

// NewSyncManager creates a new sync manager
func NewSyncManager(client *Client, interval time.Duration) *SyncManager {
	return &SyncManager{
		client:   client,
		interval: interval,
		done:     make(chan bool),
	}
}

// Start begins the periodic sync process
func (sm *SyncManager) Start() {
	sm.ticker = time.NewTicker(sm.interval)
	sm.wg.Add(1)

	go func() {
		defer sm.wg.Done()

		for {
			select {
			case <-sm.ticker.C:
				sm.performSync()
			case <-sm.done:
				return
			}
		}
	}()
}

// performSync executes synchronization with exclusive control
func (sm *SyncManager) performSync() {
	// Try to acquire sync lock, skip if already syncing
	if !sm.syncMutex.TryLock() {
		// Previous sync still running, skip this cycle
		return
	}
	defer sm.syncMutex.Unlock()

	sm.syncing = true
	defer func() { sm.syncing = false }()

	// Check if there are dirty records
	dirtyKeys := sm.client.cache.GetDirtyKeys()
	if len(dirtyKeys) == 0 {
		return
	}

	// Perform sync
	_ = sm.client.saveToAdapter(context.Background())
}

// Stop stops the sync manager and waits for ongoing sync
func (sm *SyncManager) Stop() {
	if sm.ticker != nil {
		sm.ticker.Stop()
	}

	close(sm.done)

	// Wait for the goroutine to finish
	sm.wg.Wait()

	// Wait for any ongoing sync to complete
	sm.syncMutex.Lock()
	sm.syncMutex.Unlock()
}
