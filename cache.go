package sheetkv

import (
	"fmt"
	"sort"
	"sync"
)

// Cache manages in-memory storage of records
type Cache struct {
	mu     sync.RWMutex
	data   map[int]*Record // Key -> Record (row number)
	dirty  map[int]bool    // 変更追跡
	schema []string        // カラム名のリスト
}

// NewCache creates a new Cache instance
func NewCache() *Cache {
	return &Cache{
		data:   make(map[int]*Record),
		dirty:  make(map[int]bool),
		schema: []string{},
	}
}

// Get retrieves a record by key (row number)
func (c *Cache) Get(key int) (*Record, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	record, exists := c.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	// Return a copy to prevent external modification
	return c.copyRecord(record), nil
}

// Set stores or updates a record
func (c *Cache) Set(key int, record *Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ensure the record has the correct key
	record.Key = key

	// Store a copy
	c.data[key] = c.copyRecord(record)
	c.dirty[key] = true

	// Update schema
	c.updateSchema(record)

	return nil
}

// Append adds a new record (fails if key already exists)
func (c *Cache) Append(record *Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[record.Key]; exists {
		return ErrDuplicateKey
	}

	// Store a copy
	c.data[record.Key] = c.copyRecord(record)
	c.dirty[record.Key] = true

	// Update schema
	c.updateSchema(record)

	return nil
}

// Update partially updates a record
func (c *Cache) Update(key int, updates map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	record, exists := c.data[key]
	if !exists {
		return ErrKeyNotFound
	}

	// Apply updates to a copy
	updatedRecord := c.copyRecord(record)
	for k, v := range updates {
		if v == nil {
			delete(updatedRecord.Values, k)
		} else {
			updatedRecord.Values[k] = v
		}
	}

	c.data[key] = updatedRecord
	c.dirty[key] = true

	// Update schema
	c.updateSchema(updatedRecord)

	return nil
}

// Delete removes a record
func (c *Cache) Delete(key int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[key]; !exists {
		return ErrKeyNotFound
	}

	delete(c.data, key)
	delete(c.dirty, key)

	return nil
}

// Query searches for records matching the given conditions
func (c *Cache) Query(query Query) ([]*Record, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Validate query
	if err := ValidateQuery(query); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	// Collect all records
	records := make([]*Record, 0, len(c.data))
	for _, record := range c.data {
		records = append(records, c.copyRecord(record))
	}

	// Apply query
	results := ApplyQuery(records, query)

	return results, nil
}

// GetAllRecords returns all records sorted by key
func (c *Cache) GetAllRecords() []*Record {
	c.mu.RLock()
	defer c.mu.RUnlock()

	records := make([]*Record, 0, len(c.data))
	for _, record := range c.data {
		records = append(records, c.copyRecord(record))
	}

	// Sort by key
	sort.Slice(records, func(i, j int) bool {
		return records[i].Key < records[j].Key
	})

	return records
}

// GetDirtyKeys returns keys of modified records
func (c *Cache) GetDirtyKeys() []int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]int, 0, len(c.dirty))
	for key, isDirty := range c.dirty {
		if isDirty {
			keys = append(keys, key)
		}
	}

	sort.Ints(keys)
	return keys
}

// ClearDirty marks all records as clean
func (c *Cache) ClearDirty() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dirty = make(map[int]bool)
}

// GetSchema returns the current schema
func (c *Cache) GetSchema() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	schema := make([]string, len(c.schema))
	copy(schema, c.schema)
	return schema
}

// SetSchema sets the schema
func (c *Cache) SetSchema(schema []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.schema = make([]string, len(schema))
	copy(c.schema, schema)
}

// Load replaces all data with the provided records
func (c *Cache) Load(records []*Record, schema []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear existing data
	c.data = make(map[int]*Record)
	c.dirty = make(map[int]bool)

	// Load new data
	for _, record := range records {
		c.data[record.Key] = c.copyRecord(record)
	}

	// Set schema
	c.schema = make([]string, len(schema))
	copy(c.schema, schema)
}

// Size returns the number of records
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.data)
}

// Clear removes all data
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[int]*Record)
	c.dirty = make(map[int]bool)
	c.schema = []string{}
}

// copyRecord creates a deep copy of a record
func (c *Cache) copyRecord(record *Record) *Record {
	copy := &Record{
		Key:    record.Key,
		Values: make(map[string]interface{}),
	}

	for k, v := range record.Values {
		copy.Values[k] = v
	}

	return copy
}

// updateSchema updates the schema based on record columns
func (c *Cache) updateSchema(record *Record) {
	// Create a map of existing columns for fast lookup
	existing := make(map[string]bool)
	for _, col := range c.schema {
		existing[col] = true
	}

	// Add new columns from the record
	for col := range record.Values {
		if !existing[col] {
			c.schema = append(c.schema, col)
		}
	}
}

// MergeSchemas merges current schema with sheet schema preserving order
func MergeSchemas(current, sheet []string) []string {
	result := make([]string, 0)
	seen := make(map[string]bool)

	// First, keep existing sheet columns in their order
	for _, col := range sheet {
		// Check if the column exists in current schema
		found := false
		for _, currCol := range current {
			if currCol == col {
				found = true
				break
			}
		}
		if found {
			result = append(result, col)
			seen[col] = true
		}
	}

	// Then, append new columns from current schema
	for _, col := range current {
		if !seen[col] {
			result = append(result, col)
		}
	}

	return result
}
