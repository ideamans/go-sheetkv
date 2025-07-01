# API Tests

API tests validate the full sheetkv client API with Google Sheets as the backend.

## Setup

1. Create a Google Spreadsheet for testing
2. Create a sheet named "api" in the spreadsheet (or use a different name and set TEST_API_SHEET)
3. Configure your `.env` file with:

```env
# Required: Google Sheets ID
TEST_GOOGLE_SHEET_ID=your-spreadsheet-id

# Required: Sheet name for API tests (default: api-test)
TEST_API_SHEET=api

# Authentication (one of the following):
# Option 1: Service account JSON file
GOOGLE_APPLICATION_CREDENTIALS=./service-account.json

# Option 2: Service account email and private key
TEST_CLIENT_EMAIL=your-service-account@project.iam.gserviceaccount.com
TEST_CLIENT_PRIVATE_KEY=-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----
```

## Running Tests

```bash
# Run all API tests
make test-api

# Run specific test
go test -v ./tests/api -run TestAPIOperations/BasicCRUD
```

## Test Coverage

The API tests cover:

- **Basic CRUD**: Append, Get, Set, Update, Delete operations
- **Data Types**: Strings, integers, floats, booleans, arrays, times, special characters
- **Query Operations**: All query operators (==, !=, >, >=, <, <=, in, between)
- **Concurrent Operations**: Thread-safe concurrent access
- **Large Data Sets**: Performance with multiple records and pagination

## Notes

- Tests focus on happy path scenarios (edge cases are covered in unit/integration tests)
- The test sheet is cleared before each test run
- Retry logic handles temporary Google API failures (503 errors)
- All tests use the full client API, not direct adaptor access