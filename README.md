# go-sheetkv

A Go library that provides a Key-Value Store (KVS) backed by spreadsheets. Supports both Google Sheets and Excel files.

## Features

- Use Google Sheets and Excel as a KVS backend
- Fast access with memory caching
- Automatic synchronization
- Type-safe API
- Built-in retry mechanism
- Multiple authentication methods (Google Sheets)
- No authentication required for local Excel files

## Important Note

⚠️ **This package is designed for simple batch processing and does not support concurrent access from multiple processes.** All data is cached in memory within each process, and there is no inter-process synchronization mechanism. Using this package from multiple processes simultaneously may result in data inconsistencies.

## Installation

```bash
go get github.com/ideamans/go-sheetkv
```

## Usage

```go
import (
    sheetkv "github.com/ideamans/go-sheetkv"
    "github.com/ideamans/go-sheetkv/adapters/googlesheets"
    "github.com/ideamans/go-sheetkv/adapters/excel"
)
```

### Basic Example

```go
package main

import (
    "context"
    "log"
    "time"
    
    sheetkv "github.com/ideamans/go-sheetkv"
    "github.com/ideamans/go-sheetkv/adapters/googlesheets"
)

func main() {
    ctx := context.Background()
    
    // Configure and create adapter
    adapterConfig := googlesheets.Config{
        SpreadsheetID: "your-spreadsheet-id",
        SheetName:     "users",
    }
    adapter, err := googlesheets.NewWithJSONKeyFile(ctx, adapterConfig, "./credentials.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create client with recommended defaults for Google Sheets
    clientConfig := googlesheets.DefaultClientConfig()
    // Optionally customize:
    // clientConfig.SyncInterval = 30 * time.Second
    
    client := sheetkv.New(adapter, clientConfig)
    
    // Initialize and load existing data
    if err := client.Initialize(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Append a record
    record := &sheetkv.Record{
        Values: map[string]interface{}{
            "name": "John Doe",
            "age":  25,
        },
    }
    err = client.Append(record)
    if err != nil {
        log.Fatal(err)
    }

    // Query records
    results, err := client.Query(sheetkv.Query{
        Conditions: []sheetkv.Condition{
            {Column: "age", Operator: ">=", Value: 20},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Display results
    for _, r := range results {
        name := r.GetAsString("name", "")
        age := r.GetAsInt64("age", 0)
        log.Printf("Row %d: %s (age: %d)", r.Key, name, age)
    }
}
```

### Using Excel

```go
package main

import (
    "context"
    "log"
    "time"
    
    sheetkv "github.com/ideamans/go-sheetkv"
    "github.com/ideamans/go-sheetkv/adapters/excel"
)

func main() {
    // Configure Excel adapter (no authentication required)
    adapterConfig := &excel.Config{
        FilePath:  "./data.xlsx",
        SheetName: "users",
    }
    adapter, err := excel.New(adapterConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Create client with recommended defaults for Excel
    client := sheetkv.New(adapter, excel.DefaultClientConfig())
    
    ctx := context.Background()
    if err := client.Initialize(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Operations are the same as Google Sheets
}
```

## Authentication

### Google Sheets Authentication

#### 1. Service Account JSON File

```go
adapter, err := googlesheets.NewWithJSONKeyFile(ctx, adapterConfig, "./service-account.json")
```

#### 2. Environment Variable

```go
// Uses GOOGLE_APPLICATION_CREDENTIALS environment variable
adapter, err := googlesheets.NewWithJSONKeyFile(ctx, adapterConfig, "")
```

#### 3. Service Account Key Direct

```go
adapter, err := googlesheets.NewWithServiceAccountKey(
    ctx, 
    adapterConfig,
    "service-account@project.iam.gserviceaccount.com",
    "-----BEGIN PRIVATE KEY-----\n...",
)
```

## Data Types

Record Values are `map[string]interface{}`, but type-safe helper methods are provided:

```go
// Getter methods
name := record.GetAsString("name", "default")
age := record.GetAsInt64("age", 0)
price := record.GetAsFloat64("price", 0.0)
active := record.GetAsBool("active", false)
tags := record.GetAsStrings("tags", []string{})
created := record.GetAsTime("created_at", time.Now())

// Setter methods
record.SetString("name", "New Name")
record.SetInt64("age", 30)
record.SetFloat64("price", 1980.0)
record.SetBool("active", true)
record.SetStrings("tags", []string{"tag1", "tag2"})
record.SetTime("updated_at", time.Now())
```

## Queries

Combine multiple conditions for complex queries:

```go
results, err := client.Query(sheetkv.Query{
    Conditions: []sheetkv.Condition{
        {Column: "status", Operator: "==", Value: "active"},
        {Column: "age", Operator: ">=", Value: 18},
        {Column: "age", Operator: "<=", Value: 65},
        {Column: "role", Operator: "in", Value: []interface{}{"admin", "user"}},
    },
    Limit:  10,
    Offset: 0,
})
```

### Supported Operators

- `==` : Equal
- `!=` : Not equal
- `>` : Greater than
- `>=` : Greater than or equal
- `<` : Less than
- `<=` : Less than or equal
- `in` : In array (value must be an array)
- `between` : Between range (value must be [2]interface{})

## Spreadsheet Structure

- Row 1: Column names (schema definition)
- Row 2+: Data records
- Keys are row numbers (starting from 2)

## Synchronization Strategies

This library implements two synchronization strategies:

### Gap-Preserving Sync (Default for Scheduled Sync)
- Deleted records are synchronized as empty rows
- Maintains consistency between memory row numbers (keys) and spreadsheet row numbers
- When appending new records, keys continue incrementing from the highest existing key
- Used automatically during periodic synchronization

### Compacting Sync (Used on Close)
- Deleted records are removed and remaining data is compacted
- Provides optimal spreadsheet size by removing empty rows
- Row numbers in the spreadsheet may not match record keys after sync
- Automatically removes trailing empty rows to maintain clean data
- Used automatically when calling `Close()` to finalize the session

## Default Configurations

### Google Sheets
- SyncInterval: 10 seconds
- MaxRetries: 3
- RetryInterval: 20 seconds

### Excel
- SyncInterval: 1 second
- MaxRetries: 3
- RetryInterval: 5 seconds

## Development

### Running Tests

```bash
# Unit tests only
make test-unit

# Integration tests (requires .env configuration)
make test-integration

# API tests (requires .env configuration)
make test-api

# All tests
make test
```

### Environment Variables

Tests require a `.env` file:

```env
# For Google Sheets testing
GOOGLE_APPLICATION_CREDENTIALS=./service-account.json
TEST_GOOGLE_SHEET_ID=your-test-spreadsheet-id

# Additional authentication methods (optional)
TEST_CLIENT_EMAIL=service-account@project.iam.gserviceaccount.com
TEST_CLIENT_PRIVATE_KEY=-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----

# Note:
# - Sheet names are automatically set (integration/api)
# - If Google Sheets credentials are not configured, tests automatically fall back to Excel
# - Excel adapter is always tested
```

### CI/CD

To run tests with Google Sheets in GitHub Actions, configure these repository secrets:

- `SERVICE_ACCOUNT_JSON`: Service account JSON file content
- `TEST_CLIENT_EMAIL`: Service account email
- `TEST_CLIENT_PRIVATE_KEY`: Service account private key
- `TEST_GOOGLE_SHEET_ID`: Test spreadsheet ID

See [.github/CI_SECRETS.md](.github/CI_SECRETS.md) for details.

## License

MIT License