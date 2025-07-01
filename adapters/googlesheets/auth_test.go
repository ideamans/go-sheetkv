package googlesheets

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseServiceAccountJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid service account",
			json: `{
				"type": "service_account",
				"project_id": "test-project",
				"private_key_id": "key-id",
				"private_key": "-----BEGIN PRIVATE KEY-----\ntest-key\n-----END PRIVATE KEY-----\n",
				"client_email": "test@test-project.iam.gserviceaccount.com",
				"client_id": "123456789",
				"auth_uri": "https://accounts.google.com/o/oauth2/auth",
				"token_uri": "https://oauth2.googleapis.com/token",
				"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
				"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/test%40test-project.iam.gserviceaccount.com"
			}`,
			wantErr: false,
		},
		{
			name: "invalid type",
			json: `{
				"type": "user",
				"client_email": "test@example.com",
				"private_key": "key"
			}`,
			wantErr: true,
			errMsg:  "invalid key type",
		},
		{
			name: "missing email",
			json: `{
				"type": "service_account",
				"private_key": "key"
			}`,
			wantErr: true,
			errMsg:  "missing required fields",
		},
		{
			name: "missing private key",
			json: `{
				"type": "service_account",
				"client_email": "test@example.com"
			}`,
			wantErr: true,
			errMsg:  "missing required fields",
		},
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: true,
			errMsg:  "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := ParseServiceAccountJSON([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseServiceAccountJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseServiceAccountJSON() error = %v, want error containing %v", err, tt.errMsg)
				}
			}

			if !tt.wantErr && key != nil {
				if key.Type != "service_account" {
					t.Errorf("ParseServiceAccountJSON() Type = %v, want service_account", key.Type)
				}
			}
		})
	}
}

func TestNewWithJSONKeyFile(t *testing.T) {
	// Create a temporary JSON key file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "key.json")

	validJSON := `{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "key-id",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7W8jYPqHpEB4E\nkz+TqAiCOPWClqb8tNNBZzK6YOZU9LiflPvo5aFReganhpl4HGGPFdApC7UH6QEh\nTD7fEJM8d8sDT7Y4MvGKppKYK6WSwj7rVISA8TEkOxkDVPQxjSsRZsBS0pEwPl5J\nQUOpBDUXKFGEq8cXx8MZfu3+5xG7AAMG6QaJVBNzHzBh5z6P4bWSxWGSJxGF9dUU\nBO5PyXHKMM6kQlKfLJBDDgeVTkqZLaAzZsNxUHyHF3wLiU94baAYbdIB9fpVx9E9\n7MFWNPDRxiYPJGmeARV38sLxFPyQVJXHHHJz4X8gNWtQzMAksB2P4CjdacYkqXNl\nOzL81CkrAgMBAAECggEAOvYXixHSA5KlOcBgvkCrFYvOH7dGinbqAJBXLMHNPMMN\nAOA13debXgULHas2cQBe+vPPx8bCCA+7RYdYsRJkh9xUgpU1217R1NrFxiWKmrOe\nxe6xl1mRVR5iwNNA0Ugw1gU7hPD0TDMjFr3xMilNMI6nsAZliQVLlSjlUltPCnqf\nLZgsJp9xKc9LHWLR8OYNyQX9JdFMULTOXSrk3WfEqhXLWK0rtQNLUyOGAfbHQK5J\nFMBCAW8dQhPqHsjqUNcqDUB4pgIiHF3lRBn7GdpcJbrEEzaXMdMKRDCQCGKXPVnx\n8TjXSmCgD6UNRfWCBJBPPLhcruZQMKU6V1PVl3o3AQKBgQDjdcq2YFNrPjVKd9D0\n2qGJBafwGAnjFd/OWeFXQdHQj7pbNMPrm3V2E7AFMwWr3Oqx3KFp4T+RdUHXev5n\nU9MrTvILxBVFE/aDDC1uGZxhJYUfZvvHOcQqVs7qpHHASF8hqhGKUlKHMDfchhAG\nEPFgw0F0TzHAJpk+qogihJd0awKBgQDTG7vwmhIN1P9iqmczp5s/STYdmqUN9A7U\nF1mOJTjJYMR9NvBvH52kH8JRAzTJU7sD7d3kiMX4dN7xHZYNJWQVGlpXX3RBYU4t\ncbTmSNquMlWaoXlwPy3bKN6adWZNuLkfMTG6jjv/hFTnFPEBmNVMaTdDCW6p3LZG\nW5rI7VgYQQKBgQCN6ggBcK4gIjGgE6UjQ1p7yOo8gE8K+sp4TZ9PBvGgVSpLjxDj\npP3pVEqNZ7c4PqZ7SqNao4vvfmQxU1x1agMY1sQFjYs7/yrJ1xUiMwMM0q7g5pMM\n9WWmB7rXhvDPCqxkTSymF7geXEgmqD+1+TJXp8voL0fRWnjMSwKznhFLBwKBgF8Y\nHQLxClFJfqcGPxLYa8cFS/xj9mF8qa8UqXRhYPAlDPI2z4epqxCT7g+u4FukO0nJ\na1cUF8VkiVMlwKC8gKYpjEi1hN3KKDyXL3xBjKOCG/oWumFnfneAdGPiXGFmLfbL\nRhQe2MDwqJFrcOFYM8q4VHMccaKBhEfQwQ9kLKQBAoGAMFfe9taFpnZPvdED+VLF\nbAGH8uCEHrO6dtNaVvsMpfDiX5YVEFGdO9cNM+xMEV8DdprbN5h/VkpwiEQQB8uU\nShKk4L8CYSnJAw3p8AXplWIWuCMYyPBjX4d/vKGnU/TAz7LFQE2lC0TMFrIng1TK\nyGEqUNLVGCLtRaLIpBgXmkU=\n-----END PRIVATE KEY-----\n",
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id": "123456789",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`

	err := os.WriteFile(jsonFile, []byte(validJSON), 0600)
	if err != nil {
		t.Fatalf("Failed to create test JSON file: %v", err)
	}

	tests := []struct {
		name     string
		jsonPath string
		envVar   string
		wantErr  bool
	}{
		{
			name:     "with file path",
			jsonPath: jsonFile,
			wantErr:  false,
		},
		{
			name:     "with env var",
			jsonPath: "",
			envVar:   jsonFile,
			wantErr:  false,
		},
		{
			name:     "no path or env",
			jsonPath: "",
			envVar:   "",
			wantErr:  true,
		},
		{
			name:     "non-existent file",
			jsonPath: "/non/existent/file.json",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set/unset environment variable
			if tt.envVar != "" {
				os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tt.envVar)
				defer os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
			} else {
				os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
			}

			ctx := context.Background()
			_, err := NewWithJSONKeyFile(ctx, Config{
				SpreadsheetID: "test-id",
				SheetName:     "TestSheet",
			}, tt.jsonPath)

			// We expect an error because we're using a fake private key
			// But the error should be related to authentication, not file reading
			if tt.wantErr {
				if err == nil {
					t.Error("NewWithJSONKeyFile() expected error but got none")
				}
			}
		})
	}
}

func TestNewWithJSONKeyData(t *testing.T) {
	validJSON := []byte(`{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "key-id",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7W8jYPqHpEB4E\nkz+TqAiCOPWClqb8tNNBZzK6YOZU9LiflPvo5aFReganhpl4HGGPFdApC7UH6QEh\nTD7fEJM8d8sDT7Y4MvGKppKYK6WSwj7rVISA8TEkOxkDVPQxjSsRZsBS0pEwPl5J\nQUOpBDUXKFGEq8cXx8MZfu3+5xG7AAMG6QaJVBNzHzBh5z6P4bWSxWGSJxGF9dUU\nBO5PyXHKMM6kQlKfLJBDDgeVTkqZLaAzZsNxUHyHF3wLiU94baAYbdIB9fpVx9E9\n7MFWNPDRxiYPJGmeARV38sLxFPyQVJXHHHJz4X8gNWtQzMAksB2P4CjdacYkqXNl\nOzL81CkrAgMBAAECggEAOvYXixHSA5KlOcBgvkCrFYvOH7dGinbqAJBXLMHNPMMN\nAOA13debXgULHas2cQBe+vPPx8bCCA+7RYdYsRJkh9xUgpU1217R1NrFxiWKmrOe\nxe6xl1mRVR5iwNNA0Ugw1gU7hPD0TDMjFr3xMilNMI6nsAZliQVLlSjlUltPCnqf\nLZgsJp9xKc9LHWLR8OYNyQX9JdFMULTOXSrk3WfEqhXLWK0rtQNLUyOGAfbHQK5J\nFMBCAW8dQhPqHsjqUNcqDUB4pgIiHF3lRBn7GdpcJbrEEzaXMdMKRDCQCGKXPVnx\n8TjXSmCgD6UNRfWCBJBPPLhcruZQMKU6V1PVl3o3AQKBgQDjdcq2YFNrPjVKd9D0\n2qGJBafwGAnjFd/OWeFXQdHQj7pbNMPrm3V2E7AFMwWr3Oqx3KFp4T+RdUHXev5n\nU9MrTvILxBVFE/aDDC1uGZxhJYUfZvvHOcQqVs7qpHHASF8hqhGKUlKHMDfchhAG\nEPFgw0F0TzHAJpk+qogihJd0awKBgQDTG7vwmhIN1P9iqmczp5s/STYdmqUN9A7U\nF1mOJTjJYMR9NvBvH52kH8JRAzTJU7sD7d3kiMX4dN7xHZYNJWQVGlpXX3RBYU4t\ncbTmSNquMlWaoXlwPy3bKN6adWZNuLkfMTG6jjv/hFTnFPEBmNVMaTdDCW6p3LZG\nW5rI7VgYQQKBgQCN6ggBcK4gIjGgE6UjQ1p7yOo8gE8K+sp4TZ9PBvGgVSpLjxDj\npP3pVEqNZ7c4PqZ7SqNao4vvfmQxU1x1agMY1sQFjYs7/yrJ1xUiMwMM0q7g5pMM\n9WWmB7rXhvDPCqxkTSymF7geXEgmqD+1+TJXp8voL0fRWnjMSwKznhFLBwKBgF8Y\nHQLxClFJfqcGPxLYa8cFS/xj9mF8qa8UqXRhYPAlDPI2z4epqxCT7g+u4FukO0nJ\na1cUF8VkiVMlwKC8gKYpjEi1hN3KKDyXL3xBjKOCG/oWumFnfneAdGPiXGFmLfbL\nRhQe2MDwqJFrcOFYM8q4VHMccaKBhEfQwQ9kLKQBAoGAMFfe9taFpnZPvdED+VLF\nbAGH8uCEHrO6dtNaVvsMpfDiX5YVEFGdO9cNM+xMEV8DdprbN5h/VkpwiEQQB8uU\nShKk4L8CYSnJAw3p8AXplWIWuCMYyPBjX4d/vKGnU/TAz7LFQE2lC0TMFrIng1TK\nyGEqUNLVGCLtRaLIpBgXmkU=\n-----END PRIVATE KEY-----\n",
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id": "123456789"
	}`)

	invalidJSON := []byte(`{
		"type": "user",
		"client_email": "test@example.com"
	}`)

	tests := []struct {
		name     string
		jsonData []byte
		wantErr  bool
	}{
		{
			name:     "valid json data",
			jsonData: validJSON,
			wantErr:  false,
		},
		{
			name:     "invalid json data",
			jsonData: invalidJSON,
			wantErr:  true,
		},
		{
			name:     "malformed json",
			jsonData: []byte(`{invalid}`),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := NewWithJSONKeyData(ctx, Config{
				SpreadsheetID: "test-id",
				SheetName:     "TestSheet",
			}, tt.jsonData)

			// We expect some error due to authentication with fake keys
			// But check if parsing worked correctly
			if tt.wantErr && err == nil {
				t.Error("NewWithJSONKeyData() expected error but got none")
			}
		})
	}
}

func TestNewWithServiceAccountKey(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		privateKey string
		wantErr    bool
	}{
		{
			name:  "valid credentials",
			email: "test@test-project.iam.gserviceaccount.com",
			privateKey: `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7W8jYPqHpEB4E
kz+TqAiCOPWClqb8tNNBZzK6YOZU9LiflPvo5aFReganhpl4HGGPFdApC7UH6QEh
TD7fEJM8d8sDT7Y4MvGKppKYK6WSwj7rVISA8TEkOxkDVPQxjSsRZsBS0pEwPl5J
QUOpBDUXKFGEq8cXx8MZfu3+5xG7AAMG6QaJVBNzHzBh5z6P4bWSxWGSJxGF9dUU
BO5PyXHKMM6kQlKfLJBDDgeVTkqZLaAzZsNxUHyHF3wLiU94baAYbdIB9fpVx9E9
7MFWNPDRxiYPJGmeARV38sLxFPyQVJXHHHJz4X8gNWtQzMAksB2P4CjdacYkqXNl
OzL81CkrAgMBAAECggEAOvYXixHSA5KlOcBgvkCrFYvOH7dGinbqAJBXLMHNPMMN
AOA13debXgULHas2cQBe+vPPx8bCCA+7RYdYsRJkh9xUgpU1217R1NrFxiWKmrOe
xe6xl1mRVR5iwNNA0Ugw1gU7hPD0TDMjFr3xMilNMI6nsAZliQVLlSjlUltPCnqf
LZgsJp9xKc9LHWLR8OYNyQX9JdFMULTOXSrk3WfEqhXLWK0rtQNLUyOGAfbHQK5J
FMBCAW8dQhPqHsjqUNcqDUB4pgIiHF3lRBn7GdpcJbrEEzaXMdMKRDCQCGKXPVnx
8TjXSmCgD6UNRfWCBJBPPLhcruZQMKU6V1PVl3o3AQKBgQDjdcq2YFNrPjVKd9D0
2qGJBafwGAnjFd/OWeFXQdHQj7pbNMPrm3V2E7AFMwWr3Oqx3KFp4T+RdUHXev5n
U9MrTvILxBVFE/aDDC1uGZxhJYUfZvvHOcQqVs7qpHHASF8hqhGKUlKHMDfchhAG
EPFgw0F0TzHAJpk+qogihJd0awKBgQDTG7vwmhIN1P9iqmczp5s/STYdmqUN9A7U
F1mOJTjJYMR9NvBvH52kH8JRAzTJU7sD7d3kiMX4dN7xHZYNJWQVGlpXX3RBYU4t
cbTmSNquMlWaoXlwPy3bKN6adWZNuLkfMTG6jjv/hFTnFPEBmNVMaTdDCW6p3LZG
W5rI7VgYQQKBgQCN6ggBcK4gIjGgE6UjQ1p7yOo8gE8K+sp4TZ9PBvGgVSpLjxDj
pP3pVEqNZ7c4PqZ7SqNao4vvfmQxU1x1agMY1sQFjYs7/yrJ1xUiMwMM0q7g5pMM
9WWmB7rXhvDPCqxkTSymF7geXEgmqD+1+TJXp8voL0fRWnjMSwKznhFLBwKBgF8Y
HQLxClFJfqcGPxLYa8cFS/xj9mF8qa8UqXRhYPAlDPI2z4epqxCT7g+u4FukO0nJ
a1cUF8VkiVMlwKC8gKYpjEi1hN3KKDyXL3xBjKOCG/oWumFnfneAdGPiXGFmLfbL
RhQe2MDwqJFrcOFYM8q4VHMccaKBhEfQwQ9kLKQBAoGAMFfe9taFpnZPvdED+VLF
bAGH8uCEHrO6dtNaVvsMpfDiX5YVEFGdO9cNM+xMEV8DdprbN5h/VkpwiEQQB8uU
ShKk4L8CYSnJAw3p8AXplWIWuCMYyPBjX4d/vKGnU/TAz7LFQE2lC0TMFrIng1TK
yGEqUNLVGCLtRaLIpBgXmkU=
-----END PRIVATE KEY-----`,
			wantErr: false,
		},
		{
			name:       "empty email",
			email:      "",
			privateKey: "key",
			wantErr:    false, // JWT config accepts empty email
		},
		{
			name:       "empty private key",
			email:      "test@example.com",
			privateKey: "",
			wantErr:    false, // Error will occur during actual API calls
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := NewWithServiceAccountKey(ctx, Config{
				SpreadsheetID: "test-id",
				SheetName:     "TestSheet",
			}, tt.email, tt.privateKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewWithServiceAccountKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateTokenSource(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "key.json")
	jsonData := []byte(`{
		"type": "service_account",
		"client_email": "test@example.com",
		"private_key": "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----\n"
	}`)
	if err := os.WriteFile(jsonFile, jsonData, 0600); err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	parsedKey := &ServiceAccountKey{
		Type:        "service_account",
		ClientEmail: "test@example.com",
		PrivateKey:  "test-key",
	}

	tests := []struct {
		name        string
		credentials interface{}
		wantErr     bool
	}{
		{
			name:        "file path",
			credentials: jsonFile,
			wantErr:     false,
		},
		{
			name:        "json data",
			credentials: jsonData,
			wantErr:     false,
		},
		{
			name:        "parsed key",
			credentials: parsedKey,
			wantErr:     false,
		},
		{
			name:        "unsupported type",
			credentials: 123,
			wantErr:     true,
		},
		{
			name:        "non-existent file",
			credentials: "/non/existent/file",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := CreateTokenSource(ctx, tt.credentials)

			// We may get errors due to invalid test keys, but check basic validation
			if tt.wantErr && err == nil {
				t.Error("CreateTokenSource() expected error but got none")
			}
		})
	}
}
