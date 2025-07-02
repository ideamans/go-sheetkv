package main

import (
	"context"
	"fmt"
	"log"
	"time"

	sheetkv "github.com/ideamans/go-sheetkv"
	"github.com/ideamans/go-sheetkv/adapters/googlesheets"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// Create adapter configuration
	adapterConfig := googlesheets.Config{
		SpreadsheetID: "your-spreadsheet-id",
		SheetName:     "example",
	}

	// Initialize Google Sheets adapter with JSON key file
	adapter, err := googlesheets.NewWithJSONKeyFile(ctx, adapterConfig, "./service-account.json")
	if err != nil {
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	// Create client using recommended defaults for Google Sheets
	clientConfig := googlesheets.DefaultClientConfig()
	// Optionally customize:
	// clientConfig.SyncInterval = 30 * time.Second

	// Initialize KVS client
	client := sheetkv.New(adapter, clientConfig)
	if err = client.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}
	defer client.Close()

	// Create a new record
	user := &sheetkv.Record{
		Values: map[string]interface{}{
			"name":       "John Doe",
			"email":      "john@example.com",
			"age":        30,
			"created_at": time.Now(),
		},
	}

	// Append the record
	err = client.Append(user)
	if err != nil {
		return fmt.Errorf("failed to append record: %w", err)
	}
	fmt.Printf("Added user with key (row number): %d\n", user.Key)

	// Query records
	results, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "age", Operator: ">=", Value: 25},
			{Column: "age", Operator: "<=", Value: 35},
		},
		Limit: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}

	fmt.Printf("Found %d users aged 25-35:\n", len(results))
	for _, record := range results {
		name := record.GetAsString("name", "Unknown")
		age := record.GetAsInt64("age", 0)
		fmt.Printf("  Row %d: %s (age: %d)\n", record.Key, name, age)
	}

	// Update a record
	if len(results) > 0 {
		firstRecord := results[0]
		err = client.Update(firstRecord.Key, map[string]interface{}{
			"last_login":  time.Now(),
			"login_count": firstRecord.GetAsInt64("login_count", 0) + 1,
		})
		if err != nil {
			log.Printf("Failed to update record: %v", err)
		} else {
			fmt.Printf("Updated record at row %d\n", firstRecord.Key)
		}
	}

	// Force sync
	err = client.Sync()
	if err != nil {
		log.Printf("Failed to sync: %v", err)
	}

	return nil
}
