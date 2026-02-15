package utils

import (
	"context"
	"encoding/json"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// BootstrapConfig represents the bootstrap configuration
type BootstrapConfig struct {
	GCPSecretName string `json:"GCPSecretName"`
}

// LoadBootstrapConfig reads and parses the bootstrap.json file
func LoadBootstrapConfig(filename string) (*BootstrapConfig, error) {
	data, err := LoadValuesFromFile(filename)
	if err != nil {
		return nil, err
	}

	// Convert map to struct
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bootstrap config: %w", err)
	}

	var config BootstrapConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bootstrap config: %w", err)
	}

	return &config, nil
}

// LoadGCPSecret loads a secret from GCP Secret Manager and returns it as a map
func LoadGCPSecret(ctx context.Context, secretName string) (map[string]interface{}, error) {
	if secretName == "" {
		// Return empty map if no secret is configured
		return make(map[string]interface{}), nil
	}

	// Create the Secret Manager client
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Secret Manager client (verify GCP credentials are configured): %w", err)
	}
	defer client.Close()

	// Build the request to access the secret version
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	// Access the secret version
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to access secret version '%s': %w", secretName, err)
	}

	// Parse the secret payload as JSON
	var secretValues map[string]interface{}
	if err := json.Unmarshal(result.Payload.Data, &secretValues); err != nil {
		return nil, fmt.Errorf("failed to parse secret as JSON: %w", err)
	}

	return secretValues, nil
}
