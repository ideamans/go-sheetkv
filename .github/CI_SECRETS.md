# CI Secrets Configuration

This document describes the GitHub Secrets required for the CI workflow to run integration and API tests with Google Sheets.

## Required Secrets

The following secrets must be configured in your GitHub repository settings:

### 1. SERVICE_ACCOUNT_JSON
The complete JSON content of your Google Cloud service account credentials file. This should be the entire JSON file content, not a file path.

Example structure:
```json
{
  "type": "service_account",
  "project_id": "your-project",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "your-service-account@your-project.iam.gserviceaccount.com",
  "client_id": "...",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "..."
}
```

### 2. TEST_CLIENT_EMAIL
The email address of the service account (same as `client_email` in the service account JSON).

Example: `your-service-account@your-project.iam.gserviceaccount.com`

### 3. TEST_CLIENT_PRIVATE_KEY
The private key from the service account (same as `private_key` in the service account JSON).

**Important**: Include the full key with header and footer, including the newline characters:
```
-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDWIOW/n1XPWQm0
...
-----END PRIVATE KEY-----
```

### 4. TEST_GOOGLE_SHEET_ID
The ID of the Google Spreadsheet to use for testing. This spreadsheet should be accessible by the service account.

Example: `1GVySkF-30gq6puUpqzmjCME-XDYO8qpUGQsN_pA6PhU`

## Setting up Secrets

1. Go to your GitHub repository
2. Navigate to Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Add each secret with the exact name and value as described above

## Notes

- The integration and API test job will only run on the main repository (not on forks) to protect secrets
- The job will automatically skip if the required secrets are not configured
- The service account needs read/write access to the test spreadsheet
- The test will create and use sheets named "integration" and "api" in the specified spreadsheet