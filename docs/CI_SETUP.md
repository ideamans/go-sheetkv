# CI Setup Guide

This guide explains how to properly set up GitHub Secrets for the CI/CD pipeline.

## Required Secrets

### 1. SERVICE_ACCOUNT_JSON
The complete service account JSON file content. This should be the entire JSON file, not base64 encoded.

### 2. TEST_GOOGLE_SHEET_ID
The ID of the Google Spreadsheet to use for testing. This can be found in the spreadsheet URL:
`https://docs.google.com/spreadsheets/d/SPREADSHEET_ID/edit`

### 3. TEST_CLIENT_EMAIL
The email address from the service account JSON file. This is typically in the format:
`service-account-name@project-id.iam.gserviceaccount.com`

### 4. TEST_CLIENT_PRIVATE_KEY
The private key from the service account JSON file. This is the most critical part:

#### Important: Private Key Format
The private key must be stored in GitHub Secrets with **escaped newlines**. This means:
- Replace all actual newline characters with `\n`
- The key should be a single line in the GitHub Secret

#### Example:
```
-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...\n...\n-----END PRIVATE KEY-----
```

#### How to prepare the private key:
1. Copy the private key from your service account JSON
2. Replace all newlines with `\n`:
   - In a text editor: Find and replace all line breaks with `\n`
   - Or use this command: `cat private_key.txt | tr '\n' '\\n'`
3. Make sure the result is a single line
4. Paste this single line into the GitHub Secret

## Setting GitHub Secrets

1. Go to your repository on GitHub
2. Navigate to Settings > Secrets and variables > Actions
3. Click "New repository secret"
4. Add each secret with the exact names above
5. For TEST_CLIENT_PRIVATE_KEY, ensure you use the escaped format

## Verifying the Setup

The test logs will now show diagnostic information with masked sensitive data:
- Email address (first 5 characters visible, rest masked)
- Private key length
- Whether the key contains newlines
- The first 30 characters of the key (rest masked with asterisks)

This helps identify if the key format is correct in the CI environment while keeping sensitive information protected.

## Troubleshooting

### Error: "Failed to create Google Sheets adapter with email/key auth"
1. Check that all secrets are set correctly
2. Verify the private key format (should have `\n` not actual newlines)
3. Ensure the service account has access to the test spreadsheet
4. Check the diagnostic logs for clues about the private key format

### The private key looks correct locally but fails in CI
This usually means the newline handling is different. The code now automatically handles both formats:
- Local `.env` file: Uses escaped newlines (`\n`) which are converted by `loadEnvFile`
- GitHub Secrets: Should also use escaped newlines (`\n`) which are now converted by the test code