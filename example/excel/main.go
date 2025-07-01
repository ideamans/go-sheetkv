package main

import (
	"context"
	"fmt"
	"log"
	"time"

	sheetkv "github.com/ideamans/go-sheetkv"
	"github.com/ideamans/go-sheetkv/adapters/excel"
)

func main() {
	// Excel adapter configuration
	adapterConfig := &excel.Config{
		FilePath:  "./example_data.xlsx",
		SheetName: "users",
	}

	// Create Excel adapter (no authentication required)
	adapter, err := excel.New(adapterConfig)
	if err != nil {
		log.Fatalf("Failed to create Excel adapter: %v", err)
	}

	// Create client using recommended defaults for Excel
	clientConfig := excel.DefaultClientConfig()
	// Optionally customize:
	// clientConfig.SyncInterval = 30 * time.Second

	client := sheetkv.New(adapter, clientConfig)

	// Initialize client (loads existing data if file exists)
	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	defer func() {
		// Ensure final sync before closing
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}()

	// 1. Add some records
	fmt.Println("Adding records...")
	users := []sheetkv.Record{
		{
			Values: map[string]interface{}{
				"id":         int64(1),
				"name":       "Alice Johnson",
				"email":      "alice@example.com",
				"age":        int64(30),
				"department": "Engineering",
				"active":     true,
				"joined_at":  time.Now().Format(time.RFC3339),
			},
		},
		{
			Values: map[string]interface{}{
				"id":         int64(2),
				"name":       "Bob Smith",
				"email":      "bob@example.com",
				"age":        int64(25),
				"department": "Marketing",
				"active":     true,
				"joined_at":  time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			},
		},
		{
			Values: map[string]interface{}{
				"id":         int64(3),
				"name":       "Charlie Brown",
				"email":      "charlie@example.com",
				"age":        int64(35),
				"department": "Engineering",
				"active":     false,
				"joined_at":  time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			},
		},
	}

	for _, user := range users {
		if err := client.Append(&user); err != nil {
			log.Printf("Failed to append user: %v", err)
		} else {
			fmt.Printf("Added user: %s (Row %d)\n", user.GetAsString("name", ""), user.Key)
		}
	}

	// 2. Query records
	fmt.Println("\nQuerying active engineers...")
	results, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "department", Operator: "==", Value: "Engineering"},
			{Column: "active", Operator: "==", Value: true},
		},
	})
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for _, r := range results {
			fmt.Printf("- %s (age: %d)\n",
				r.GetAsString("name", ""),
				r.GetAsInt64("age", 0))
		}
	}

	// 3. Update a record
	fmt.Println("\nUpdating Bob's department...")
	bobResults, err := client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "name", Operator: "==", Value: "Bob Smith"},
		},
	})
	if err == nil && len(bobResults) > 0 {
		bobKey := bobResults[0].Key
		err = client.Update(bobKey, map[string]interface{}{
			"department": "Sales",
			"updated_at": time.Now().Format(time.RFC3339),
		})
		if err != nil {
			log.Printf("Update failed: %v", err)
		} else {
			fmt.Println("Updated successfully")
		}
	}

	// 4. Type-safe operations
	fmt.Println("\nUsing type-safe methods...")
	record := &sheetkv.Record{
		Values: make(map[string]interface{}),
	}

	// Set values with type-safe setters
	record.SetString("name", "Diana Prince")
	record.SetString("email", "diana@example.com")
	record.SetInt64("age", 28)
	record.SetBool("active", true)
	record.SetStrings("skills", []string{"Java", "Python", "Go"})
	record.SetTime("created_at", time.Now())

	if err := client.Append(record); err != nil {
		log.Printf("Failed to append record: %v", err)
	} else {
		fmt.Printf("Added Diana (Row %d)\n", record.Key)
	}

	// 5. Complex query with multiple conditions
	fmt.Println("\nFinding users between 25-35 years old...")
	results, err = client.Query(sheetkv.Query{
		Conditions: []sheetkv.Condition{
			{Column: "age", Operator: ">=", Value: int64(25)},
			{Column: "age", Operator: "<=", Value: int64(35)},
		},
		Limit: 10,
	})
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for _, r := range results {
			skills := r.GetAsStrings("skills", []string{})
			fmt.Printf("- %s (age: %d, skills: %v)\n",
				r.GetAsString("name", ""),
				r.GetAsInt64("age", 0),
				skills)
		}
	}

	// 6. Force sync to ensure data is written to Excel file
	fmt.Println("\nSyncing data to Excel file...")
	if err := client.Sync(); err != nil {
		log.Printf("Sync failed: %v", err)
	} else {
		fmt.Println("Data synced successfully")
	}

	// The Excel file will be created/updated at ./example_data.xlsx
	fmt.Println("\nExample completed. Check ./example_data.xlsx for the data.")
}
