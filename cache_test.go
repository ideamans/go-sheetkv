package sheetkv_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/ideamans/go-sheetkv"
)

func TestCache_Basic(t *testing.T) {
	cache := sheetkv.NewCache()

	t.Run("Get non-existent key", func(t *testing.T) {
		_, err := cache.Get(999)
		if err != sheetkv.ErrKeyNotFound {
			t.Errorf("Get() error = %v, want %v", err, sheetkv.ErrKeyNotFound)
		}
	})

	t.Run("Set and Get", func(t *testing.T) {
		record := &sheetkv.Record{
			Key:    2,
			Values: map[string]interface{}{"name": "John", "age": 30},
		}

		err := cache.Set(2, record)
		if err != nil {
			t.Errorf("Set() error = %v", err)
		}

		got, err := cache.Get(2)
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}

		if got.Key != 2 {
			t.Errorf("Get() Key = %v, want %v", got.Key, 2)
		}
		if got.Values["name"] != "John" {
			t.Errorf("Get() name = %v, want %v", got.Values["name"], "John")
		}
		if got.Values["age"] != 30 {
			t.Errorf("Get() age = %v, want %v", got.Values["age"], 30)
		}
	})

	t.Run("Size", func(t *testing.T) {
		if cache.Size() != 1 {
			t.Errorf("Size() = %v, want %v", cache.Size(), 1)
		}
	})
}

func TestCache_Append(t *testing.T) {
	cache := sheetkv.NewCache()

	record1 := &sheetkv.Record{
		Key:    2,
		Values: map[string]interface{}{"name": "John"},
	}

	t.Run("Append new record", func(t *testing.T) {
		err := cache.Append(record1)
		if err != nil {
			t.Errorf("Append() error = %v", err)
		}

		got, _ := cache.Get(2)
		if got.Values["name"] != "John" {
			t.Errorf("Append() failed to store record")
		}
	})

	t.Run("Append duplicate key", func(t *testing.T) {
		record2 := &sheetkv.Record{
			Key:    2,
			Values: map[string]interface{}{"name": "Jane"},
		}

		err := cache.Append(record2)
		if err != sheetkv.ErrDuplicateKey {
			t.Errorf("Append() error = %v, want %v", err, sheetkv.ErrDuplicateKey)
		}
	})
}

func TestCache_Update(t *testing.T) {
	cache := sheetkv.NewCache()

	record := &sheetkv.Record{
		Key:    2,
		Values: map[string]interface{}{"name": "John", "age": 30, "city": "NYC"},
	}
	cache.Set(2, record)

	t.Run("Update existing fields", func(t *testing.T) {
		updates := map[string]interface{}{
			"age":  31,
			"city": "Boston",
		}

		err := cache.Update(2, updates)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		got, _ := cache.Get(2)
		if got.Values["age"] != 31 {
			t.Errorf("Update() age = %v, want %v", got.Values["age"], 31)
		}
		if got.Values["city"] != "Boston" {
			t.Errorf("Update() city = %v, want %v", got.Values["city"], "Boston")
		}
		if got.Values["name"] != "John" {
			t.Errorf("Update() name = %v, want %v", got.Values["name"], "John")
		}
	})

	t.Run("Update with nil removes field", func(t *testing.T) {
		updates := map[string]interface{}{
			"city": nil,
		}

		err := cache.Update(2, updates)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		got, _ := cache.Get(2)
		if _, exists := got.Values["city"]; exists {
			t.Error("Update() failed to remove field with nil value")
		}
	})

	t.Run("Update non-existent key", func(t *testing.T) {
		err := cache.Update(999, map[string]interface{}{"name": "Test"})
		if err != sheetkv.ErrKeyNotFound {
			t.Errorf("Update() error = %v, want %v", err, sheetkv.ErrKeyNotFound)
		}
	})
}

func TestCache_Delete(t *testing.T) {
	cache := sheetkv.NewCache()

	record := &sheetkv.Record{
		Key:    2,
		Values: map[string]interface{}{"name": "John"},
	}
	cache.Set(2, record)

	t.Run("Delete existing key", func(t *testing.T) {
		err := cache.Delete(2)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		_, err = cache.Get(2)
		if err != sheetkv.ErrKeyNotFound {
			t.Error("Delete() failed to remove record")
		}

		if cache.Size() != 0 {
			t.Errorf("Size() = %v after delete, want 0", cache.Size())
		}
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		err := cache.Delete(999)
		if err != sheetkv.ErrKeyNotFound {
			t.Errorf("Delete() error = %v, want %v", err, sheetkv.ErrKeyNotFound)
		}
	})
}

func TestCache_Query(t *testing.T) {
	cache := sheetkv.NewCache()

	// Load test data
	records := []*sheetkv.Record{
		{Key: 2, Values: map[string]interface{}{"age": 25, "status": "active"}},
		{Key: 3, Values: map[string]interface{}{"age": 30, "status": "inactive"}},
		{Key: 4, Values: map[string]interface{}{"age": 35, "status": "active"}},
		{Key: 5, Values: map[string]interface{}{"age": 20, "status": "active"}},
	}

	for _, r := range records {
		cache.Set(r.Key, r)
	}

	t.Run("Query with conditions", func(t *testing.T) {
		query := sheetkv.Query{
			Conditions: []sheetkv.Condition{
				{Column: "status", Operator: "==", Value: "active"},
				{Column: "age", Operator: ">=", Value: 25},
			},
		}

		results, err := cache.Query(query)
		if err != nil {
			t.Errorf("Query() error = %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Query() returned %d records, want 2", len(results))
		}

		expectedKeys := []int{2, 4}
		for i, r := range results {
			found := false
			for _, k := range expectedKeys {
				if r.Key == k {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Query() result[%d] has unexpected key %v", i, r.Key)
			}
		}
	})

	t.Run("Query with invalid operator", func(t *testing.T) {
		query := sheetkv.Query{
			Conditions: []sheetkv.Condition{
				{Column: "status", Operator: "invalid", Value: "active"},
			},
		}

		_, err := cache.Query(query)
		if err == nil {
			t.Error("Query() should fail with invalid operator")
		}
	})
}

func TestCache_DirtyTracking(t *testing.T) {
	cache := sheetkv.NewCache()

	t.Run("New record marked as dirty", func(t *testing.T) {
		record := &sheetkv.Record{
			Key:    2,
			Values: map[string]interface{}{"name": "John"},
		}
		cache.Set(2, record)

		dirtyKeys := cache.GetDirtyKeys()
		if len(dirtyKeys) != 1 || dirtyKeys[0] != 2 {
			t.Errorf("GetDirtyKeys() = %v, want [2]", dirtyKeys)
		}
	})

	t.Run("Updated record marked as dirty", func(t *testing.T) {
		cache.ClearDirty()
		cache.Update(2, map[string]interface{}{"age": 30})

		dirtyKeys := cache.GetDirtyKeys()
		if len(dirtyKeys) != 1 || dirtyKeys[0] != 2 {
			t.Errorf("GetDirtyKeys() after update = %v, want [2]", dirtyKeys)
		}
	})

	t.Run("ClearDirty removes all dirty marks", func(t *testing.T) {
		cache.ClearDirty()
		dirtyKeys := cache.GetDirtyKeys()
		if len(dirtyKeys) != 0 {
			t.Errorf("GetDirtyKeys() after clear = %v, want []", dirtyKeys)
		}
	})

	t.Run("Deleted record not in dirty", func(t *testing.T) {
		cache.Set(3, &sheetkv.Record{Key: 3, Values: map[string]interface{}{}})
		cache.Delete(3)

		dirtyKeys := cache.GetDirtyKeys()
		for _, k := range dirtyKeys {
			if k == 3 {
				t.Error("Deleted key should not be in dirty keys")
			}
		}
	})
}

func TestCache_Schema(t *testing.T) {
	cache := sheetkv.NewCache()

	t.Run("Schema updated on Set", func(t *testing.T) {
		record1 := &sheetkv.Record{
			Key:    2,
			Values: map[string]interface{}{"name": "John", "age": 30},
		}
		cache.Set(2, record1)

		schema := cache.GetSchema()
		if !containsAll(schema, []string{"name", "age"}) {
			t.Errorf("GetSchema() = %v, should contain name and age", schema)
		}

		record2 := &sheetkv.Record{
			Key:    3,
			Values: map[string]interface{}{"email": "john@example.com", "age": 25},
		}
		cache.Set(3, record2)

		schema = cache.GetSchema()
		if !containsAll(schema, []string{"name", "age", "email"}) {
			t.Errorf("GetSchema() = %v, should contain name, age, and email", schema)
		}
	})

	t.Run("SetSchema", func(t *testing.T) {
		newSchema := []string{"id", "name", "email", "created_at"}
		cache.SetSchema(newSchema)

		schema := cache.GetSchema()
		if !reflect.DeepEqual(schema, newSchema) {
			t.Errorf("GetSchema() = %v, want %v", schema, newSchema)
		}
	})
}

func TestCache_GetAllRecords(t *testing.T) {
	cache := sheetkv.NewCache()

	records := []*sheetkv.Record{
		{Key: 4, Values: map[string]interface{}{"name": "Charlie"}},
		{Key: 2, Values: map[string]interface{}{"name": "Alice"}},
		{Key: 3, Values: map[string]interface{}{"name": "Bob"}},
	}

	for _, r := range records {
		cache.Set(r.Key, r)
	}

	t.Run("GetAllRecords returns sorted records", func(t *testing.T) {
		all := cache.GetAllRecords()

		if len(all) != 3 {
			t.Errorf("GetAllRecords() returned %d records, want 3", len(all))
		}

		// Check if sorted by key
		expectedOrder := []int{2, 3, 4}
		for i, r := range all {
			if r.Key != expectedOrder[i] {
				t.Errorf("GetAllRecords()[%d].Key = %v, want %v", i, r.Key, expectedOrder[i])
			}
		}
	})
}

func TestCache_Load(t *testing.T) {
	cache := sheetkv.NewCache()

	// Add some initial data
	cache.Set(10, &sheetkv.Record{Key: 10, Values: map[string]interface{}{"name": "Old"}})

	newRecords := []*sheetkv.Record{
		{Key: 20, Values: map[string]interface{}{"name": "New1"}},
		{Key: 21, Values: map[string]interface{}{"email": "new2@example.com"}},
	}
	newSchema := []string{"name", "email", "phone"}

	t.Run("Load replaces all data", func(t *testing.T) {
		cache.Load(newRecords, newSchema)

		// Old record should be gone
		_, err := cache.Get(10)
		if err != sheetkv.ErrKeyNotFound {
			t.Error("Load() should have removed old records")
		}

		// New records should exist
		got1, err := cache.Get(20)
		if err != nil || got1.Values["name"] != "New1" {
			t.Error("Load() failed to load new record")
		}

		// Schema should be updated
		schema := cache.GetSchema()
		if !reflect.DeepEqual(schema, newSchema) {
			t.Errorf("GetSchema() = %v, want %v", schema, newSchema)
		}

		// Size should match
		if cache.Size() != 2 {
			t.Errorf("Size() = %v, want 2", cache.Size())
		}

		// Dirty should be cleared
		if len(cache.GetDirtyKeys()) != 0 {
			t.Error("Load() should clear dirty tracking")
		}
	})
}

func TestCache_Clear(t *testing.T) {
	cache := sheetkv.NewCache()

	// Add data
	cache.Set(2, &sheetkv.Record{Key: 2, Values: map[string]interface{}{"name": "John"}})
	cache.SetSchema([]string{"name", "email"})

	t.Run("Clear removes everything", func(t *testing.T) {
		cache.Clear()

		if cache.Size() != 0 {
			t.Errorf("Size() = %v after Clear(), want 0", cache.Size())
		}

		schema := cache.GetSchema()
		if len(schema) != 0 {
			t.Errorf("GetSchema() = %v after Clear(), want []", schema)
		}

		dirtyKeys := cache.GetDirtyKeys()
		if len(dirtyKeys) != 0 {
			t.Errorf("GetDirtyKeys() = %v after Clear(), want []", dirtyKeys)
		}
	})
}

func TestCache_Isolation(t *testing.T) {
	cache := sheetkv.NewCache()

	t.Run("Modifications to returned record don't affect cache", func(t *testing.T) {
		original := &sheetkv.Record{
			Key:    2,
			Values: map[string]interface{}{"name": "John", "age": 30},
		}
		cache.Set(2, original)

		// Modify the original
		original.Values["name"] = "Jane"
		original.Values["age"] = 40

		// Get from cache
		got, _ := cache.Get(2)
		if got.Values["name"] != "John" {
			t.Error("Cache should not be affected by external modifications")
		}
		if got.Values["age"] != 30 {
			t.Error("Cache should not be affected by external modifications")
		}

		// Modify the returned record
		got.Values["name"] = "Bob"

		// Get again
		got2, _ := cache.Get(2)
		if got2.Values["name"] != "John" {
			t.Error("Cache should not be affected by modifications to returned records")
		}
	})
}

func TestCache_Concurrency(t *testing.T) {
	cache := sheetkv.NewCache()

	// Test concurrent operations
	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := n + 100 // Use unique integer keys starting from 100
			record := &sheetkv.Record{
				Key:    key,
				Values: map[string]interface{}{"value": n},
			}
			cache.Set(key, record)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := (n % 10) + 100 // Read some existing keys
			cache.Get(key)
		}(i)
	}

	// Concurrent queries
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			query := sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "value", Operator: ">=", Value: 50},
				},
			}
			cache.Query(query)
		}()
	}

	wg.Wait()

	// Verify data integrity
	if cache.Size() != numGoroutines {
		t.Errorf("Size() = %v, want %v", cache.Size(), numGoroutines)
	}
}

func TestMergeSchemas(t *testing.T) {
	tests := []struct {
		name    string
		current []string
		sheet   []string
		want    []string
	}{
		{
			name:    "preserve sheet order for common columns",
			current: []string{"name", "email", "age"},
			sheet:   []string{"age", "name", "phone"},
			want:    []string{"age", "name", "email"},
		},
		{
			name:    "append new columns",
			current: []string{"name", "email", "city"},
			sheet:   []string{"name", "age"},
			want:    []string{"name", "email", "city"},
		},
		{
			name:    "empty sheet schema",
			current: []string{"name", "email"},
			sheet:   []string{},
			want:    []string{"name", "email"},
		},
		{
			name:    "empty current schema",
			current: []string{},
			sheet:   []string{"name", "email"},
			want:    []string{},
		},
		{
			name:    "no common columns",
			current: []string{"a", "b", "c"},
			sheet:   []string{"x", "y", "z"},
			want:    []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sheetkv.MergeSchemas(tt.current, tt.sheet)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeSchemas() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to check if slice contains all elements
func containsAll(slice []string, elements []string) bool {
	elementMap := make(map[string]bool)
	for _, s := range slice {
		elementMap[s] = true
	}
	for _, e := range elements {
		if !elementMap[e] {
			return false
		}
	}
	return true
}
