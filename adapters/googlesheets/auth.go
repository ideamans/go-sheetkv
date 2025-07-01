package googlesheets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// ServiceAccountKey represents the structure of a service account JSON key file
type ServiceAccountKey struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// NewWithJSONKeyFile creates a new SheetsAdaptor using a JSON key file
func NewWithJSONKeyFile(ctx context.Context, config Config, jsonPath string) (*SheetsAdaptor, error) {
	// If jsonPath is empty, try GOOGLE_APPLICATION_CREDENTIALS env var
	if jsonPath == "" {
		jsonPath = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if jsonPath == "" {
			return nil, fmt.Errorf("no JSON key file path provided and GOOGLE_APPLICATION_CREDENTIALS not set")
		}
	}

	// Read the JSON key file
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON key file: %w", err)
	}

	// Parse credentials
	creds, err := google.CredentialsFromJSON(ctx, jsonData, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return NewSheetsAdaptor(ctx, config, option.WithCredentials(creds))
}

// NewWithJSONKeyData creates a new SheetsAdaptor using JSON key data
func NewWithJSONKeyData(ctx context.Context, config Config, jsonData []byte) (*SheetsAdaptor, error) {
	// Parse credentials
	creds, err := google.CredentialsFromJSON(ctx, jsonData, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return NewSheetsAdaptor(ctx, config, option.WithCredentials(creds))
}

// NewWithServiceAccountKey creates a new SheetsAdaptor using email and private key
func NewWithServiceAccountKey(ctx context.Context, config Config, email string, privateKey string) (*SheetsAdaptor, error) {
	// Create JWT config
	jwtConfig := &jwt.Config{
		Email:      email,
		PrivateKey: []byte(privateKey),
		Scopes:     []string{sheets.SpreadsheetsScope},
		TokenURL:   google.JWTTokenURL,
	}

	// Create token source
	tokenSource := jwtConfig.TokenSource(ctx)

	return NewSheetsAdaptor(ctx, config, option.WithTokenSource(tokenSource))
}

// NewWithDefaultCredentials creates a new SheetsAdaptor using Application Default Credentials
func NewWithDefaultCredentials(ctx context.Context, config Config) (*SheetsAdaptor, error) {
	// This will use:
	// 1. GOOGLE_APPLICATION_CREDENTIALS environment variable if set
	// 2. gcloud auth application-default credentials if available
	// 3. GCE metadata service if running on Google Cloud

	tokenSource, err := google.DefaultTokenSource(ctx, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get default token source: %w", err)
	}

	return NewSheetsAdaptor(ctx, config, option.WithTokenSource(tokenSource))
}

// ParseServiceAccountJSON parses a service account JSON file or data
func ParseServiceAccountJSON(jsonData []byte) (*ServiceAccountKey, error) {
	var key ServiceAccountKey
	if err := json.Unmarshal(jsonData, &key); err != nil {
		return nil, fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	if key.Type != "service_account" {
		return nil, fmt.Errorf("invalid key type: %s (expected: service_account)", key.Type)
	}

	if key.ClientEmail == "" || key.PrivateKey == "" {
		return nil, fmt.Errorf("missing required fields in service account key")
	}

	return &key, nil
}

// CreateTokenSource creates an oauth2.TokenSource from various credential types
func CreateTokenSource(ctx context.Context, credentials interface{}) (oauth2.TokenSource, error) {
	switch cred := credentials.(type) {
	case string:
		// Assume it's a file path
		return createTokenSourceFromFile(ctx, cred)
	case []byte:
		// JSON data
		return createTokenSourceFromJSON(ctx, cred)
	case *ServiceAccountKey:
		// Parsed service account key
		return createTokenSourceFromKey(ctx, cred)
	default:
		return nil, fmt.Errorf("unsupported credential type: %T", credentials)
	}
}

func createTokenSourceFromFile(ctx context.Context, path string) (oauth2.TokenSource, error) {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}
	return createTokenSourceFromJSON(ctx, jsonData)
}

func createTokenSourceFromJSON(ctx context.Context, jsonData []byte) (oauth2.TokenSource, error) {
	creds, err := google.CredentialsFromJSON(ctx, jsonData, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}
	return creds.TokenSource, nil
}

func createTokenSourceFromKey(ctx context.Context, key *ServiceAccountKey) (oauth2.TokenSource, error) {
	jwtConfig := &jwt.Config{
		Email:      key.ClientEmail,
		PrivateKey: []byte(key.PrivateKey),
		Scopes:     []string{sheets.SpreadsheetsScope},
		TokenURL:   google.JWTTokenURL,
	}
	return jwtConfig.TokenSource(ctx), nil
}
