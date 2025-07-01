//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ideamans/go-sheetkv"
	"github.com/ideamans/go-sheetkv/adapters/googlesheets"
)

// This example demonstrates how to use the Google Sheets adaptor
// for basic KVS operations with proper authentication and error handling.
func ExampleGoogleSheetsAdaptor() {
	ctx := context.Background()

	// Configuration
	config := googlesheets.Config{
		SpreadsheetID: os.Getenv("TEST_GOOGLE_SHEET_ID"),
		SheetName:     "example-data",
	}

	// Create adaptor using JSON file authentication
	adaptor, err := googlesheets.NewWithJSONKeyFile(ctx, config, "")
	if err != nil {
		log.Fatalf("Failed to create adaptor: %v", err)
	}

	// Clear the sheet first
	if err := adaptor.Save(ctx, []*sheetkv.Record{}, []string{}); err != nil {
		log.Fatalf("Failed to clear sheet: %v", err)
	}

	// Define schema
	schema := []string{"name", "email", "age", "active"}

	// Create sample records
	records := []*sheetkv.Record{
		{
			Key: 2,
			Values: map[string]interface{}{
				"name":   "Alice Johnson",
				"email":  "alice@example.com",
				"age":    int64(28),
				"active": true,
			},
		},
		{
			Key: 3,
			Values: map[string]interface{}{
				"name":   "Bob Smith",
				"email":  "bob@example.com",
				"age":    int64(35),
				"active": false,
			},
		},
	}

	// Save records to sheet
	if err := adaptor.Save(ctx, records, schema); err != nil {
		log.Fatalf("Failed to save records: %v", err)
	}
	fmt.Println("Records saved successfully!")

	// Load records back
	loadedRecords, loadedSchema, err := adaptor.Load(ctx)
	if err != nil {
		log.Fatalf("Failed to load records: %v", err)
	}

	fmt.Printf("Loaded %d records with schema: %v\n", len(loadedRecords), loadedSchema)

	// Display loaded records
	for _, record := range loadedRecords {
		fmt.Printf("\nRecord %d:\n", record.Key)
		fmt.Printf("  Name: %s\n", record.GetAsString("name", ""))
		fmt.Printf("  Email: %s\n", record.GetAsString("email", ""))
		fmt.Printf("  Age: %d\n", record.GetAsInt64("age", 0))
		fmt.Printf("  Active: %v\n", record.GetAsBool("active", false))
	}

	// Perform batch operations
	operations := []sheetkv.Operation{
		// Update existing record
		{
			Type: sheetkv.OpUpdate,
			Record: &sheetkv.Record{
				Key: 2,
				Values: map[string]interface{}{
					"age": int64(29), // Birthday!
				},
			},
		},
		// Add new record
		{
			Type: sheetkv.OpAdd,
			Record: &sheetkv.Record{
				Key: 4,
				Values: map[string]interface{}{
					"name":   "Charlie Brown",
					"email":  "charlie@example.com",
					"age":    int64(42),
					"active": true,
				},
			},
		},
		// Delete a record
		{
			Type: sheetkv.OpDelete,
			Record: &sheetkv.Record{
				Key: 3,
			},
		},
	}

	if err := adaptor.BatchUpdate(ctx, operations); err != nil {
		log.Fatalf("Failed to perform batch update: %v", err)
	}
	fmt.Println("\nBatch operations completed successfully!")

	// Load and display final state
	finalRecords, _, err := adaptor.Load(ctx)
	if err != nil {
		log.Fatalf("Failed to load final records: %v", err)
	}

	fmt.Printf("\nFinal state: %d records\n", len(finalRecords))
	for _, record := range finalRecords {
		fmt.Printf("- %d: %s (age: %d)\n",
			record.Key,
			record.GetAsString("name", ""),
			record.GetAsInt64("age", 0))
	}
}

func main() {
	ExampleGoogleSheetsAdaptor()
}
