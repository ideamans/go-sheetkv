# Google Sheets KVS Integration Tests

This directory contains comprehensive integration tests for the Google Sheets KVS synchronization functionality.

## Test Coverage

The integration tests cover the following scenarios:

1. **Clear Sheet** - Tests clearing the entire spreadsheet
2. **Basic CRUD Operations** - Create, Read, Update, Delete operations
3. **Data Types** - Tests various data types (string, int, float, bool, arrays, time)
4. **Schema Changes** - Adding/removing columns, column reordering
5. **Batch Operations** - Multiple operations in a single request
6. **Edge Cases** - Special characters, Unicode, long values, many columns
7. **Large Data Sets** - Performance testing with 100+ records
8. **Authentication Methods** - Tests both JSON file and email/private key authentication

## Configuration

Before running the tests, you need to configure authentication and specify a test spreadsheet.

### 1. Create a Test Spreadsheet

1. Go to [Google Sheets](https://sheets.google.com)
2. Create a new spreadsheet for testing
3. Copy the spreadsheet ID from the URL (it's the long string between `/d/` and `/edit`)

### 2. Set Up Authentication

You need a Google Cloud service account with access to the Google Sheets API.

#### Option A: Using a Service Account JSON File

1. Go to the [Google Cloud Console](https://console.cloud.google.com)
2. Create or select a project
3. Enable the Google Sheets API
4. Create a service account and download the JSON key file
5. Place the JSON file in the project root (e.g., `service-account.json`)

#### Option B: Using Email and Private Key

If you already have the service account email and private key, you can use them directly.

### 3. Configure Environment Variables

Copy `.env.example` to `.env` and fill in the values:

```bash
cd ../..  # Go to project root
cp .env.example .env
```

Edit `.env`:

```bash
# For JSON file authentication
GOOGLE_APPLICATION_CREDENTIALS=./service-account.json

# For email/private key authentication (alternative to JSON file)
TEST_CLIENT_EMAIL=your-service-account@project.iam.gserviceaccount.com
TEST_CLIENT_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"

# Test spreadsheet configuration
TEST_GOOGLE_SHEET_ID=your-spreadsheet-id-here
TEST_INTEGRATION_SHEET=integration-test  # Optional, defaults to "integration-test"
```

### 4. Grant Access to the Service Account

Share your test spreadsheet with the service account email:
1. Open your test spreadsheet
2. Click "Share" button
3. Add the service account email (e.g., `service-account@project.iam.gserviceaccount.com`)
4. Give it "Editor" access

## Running the Tests

### Using the Test Runner Script

The easiest way to run the tests is using the provided script:

```bash
cd tests/integration
./run_tests.sh
```

This script will:
- Load the `.env` file
- Verify required environment variables
- Run all integration tests
- Run authentication tests

### Running Tests Manually

You can also run the tests directly with Go:

```bash
# From project root
go test -v ./tests/integration -timeout 10m

# Run specific test
go test -v ./tests/integration -run TestGoogleSheetsIntegration/BasicCRUD -timeout 5m

# Run with custom environment
TEST_GOOGLE_SHEET_ID=your-id go test -v ./tests/integration
```

### Running Individual Test Suites

```bash
# Test only authentication methods
go test -v ./tests/integration -run TestAuthenticationMethods

# Test only basic CRUD operations
go test -v ./tests/integration -run TestGoogleSheetsIntegration/BasicCRUD

# Test only edge cases
go test -v ./tests/integration -run TestGoogleSheetsIntegration/EdgeCases
```

## Test Output

The tests provide detailed output including:
- Authentication method used
- Number of records and columns processed
- Performance metrics for large data sets
- Detailed error messages for failures

Example output:
```
=== RUN   TestGoogleSheetsIntegration
=== RUN   TestGoogleSheetsIntegration/ClearSheet
=== RUN   TestGoogleSheetsIntegration/BasicCRUD
=== RUN   TestGoogleSheetsIntegration/DataTypes
=== RUN   TestGoogleSheetsIntegration/SchemaChanges
=== RUN   TestGoogleSheetsIntegration/BatchOperations
=== RUN   TestGoogleSheetsIntegration/EdgeCases
=== RUN   TestGoogleSheetsIntegration/LargeDataSet
    googlesheets_test.go:865: Saved 100 records with 10 columns in 1.234s
    googlesheets_test.go:875: Loaded 100 records with 10 columns in 0.567s
--- PASS: TestGoogleSheetsIntegration (15.23s)
```

## Troubleshooting

### Tests are skipped

If you see "Skipping integration tests", check that:
1. The `.env` file exists and is properly formatted
2. Required environment variables are set
3. The service account has access to the spreadsheet

### Authentication failures

1. Verify the service account JSON file path is correct
2. Check that the private key is properly formatted (including `\n` for newlines)
3. Ensure the Google Sheets API is enabled in your Google Cloud project

### Permission errors

Make sure the service account has "Editor" access to the test spreadsheet.

### Quota errors

Google Sheets API has usage quotas. If you hit quota limits:
1. Wait a few minutes before retrying
2. Reduce the size of the large data set test
3. Use a different Google Cloud project

## Clean Up

The tests automatically clean up data after each test run. However, you may want to:
1. Periodically check the test spreadsheet for leftover data
2. Delete old test sheets if using different sheet names

## Adding New Tests

To add new integration tests:

1. Add a new test function in `googlesheets_test.go`
2. Follow the naming convention `testScenarioName(t *testing.T, adaptor *googlesheets.SheetsAdaptor)`
3. Add the test to the main `TestGoogleSheetsIntegration` function
4. Clear the sheet at the beginning of your test
5. Verify both the operation and the data integrity