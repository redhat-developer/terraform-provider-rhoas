package rhoas_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas"
)

// TestProviderIsValid tests that the provider is correctly configured to
// work with Terraform
func TestProviderIsValid(t *testing.T) {
	if err := rhoas.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// TestProviderConfigure checks that the RHOAS client is created
func TestProviderConfigure(t *testing.T) {
	diag := rhoas.Provider().Configure(context.TODO(), &terraform.ResourceConfig{})
	assert.Empty(t, diag, "got unexpected diagnostics")
}

// TestProviderSchema checks that the RHOAS provider schema is the expected one
func TestProviderSchema(t *testing.T) {
	schemaRequest := terraform.ProviderSchemaRequest{
		ResourceTypes: []string{"rhoas_kafka", "rhoas_topic", "rhoas_service_account"},
		DataSources:   []string{"rhoas_cloud_providers", "rhoas_cloud_provider_regions", "rhoas_kafkas", "rhoas_kafka", "rhoas_service_accounts"},
	}

	providerSchema, err := rhoas.Provider().GetSchema(&schemaRequest)
	assert.NoError(t, err, "unexpected error getting the provider schema")
	t.Run("data sources", func(t *testing.T) {
		sut := providerSchema.DataSources
		assert.Contains(t, sut, "rhoas_cloud_providers")
		assert.Contains(t, sut, "rhoas_cloud_provider_regions")
		assert.Contains(t, sut, "rhoas_kafkas")
		assert.Contains(t, sut, "rhoas_kafka")
		assert.Contains(t, sut, "rhoas_service_accounts")
	})
	t.Run("resource types", func(t *testing.T) {
		sut := providerSchema.ResourceTypes
		assert.Contains(t, sut, "rhoas_kafka")
		assert.Contains(t, sut, "rhoas_topic")
		assert.Contains(t, sut, "rhoas_service_account")
	})
	t.Run("attributes", func(t *testing.T) {
		sut := providerSchema.Provider.Attributes
		assert.Contains(t, sut, "offline_token")
		assert.Contains(t, sut, "auth_url")
		assert.Contains(t, sut, "client_id")
		assert.Contains(t, sut, "api_url")
	})
}
