package rhoas_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas"
	"github.com/stretchr/testify/assert"
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

// TestProviderConfigureLocalServer checks that the RHOAS provider can be configured to work with a local server
func TestProviderConfigureLocalServer(t *testing.T) {
	os.Setenv("API", "mock")
	defer os.Setenv("API", "")

	diag := rhoas.Provider().Configure(context.TODO(), &terraform.ResourceConfig{})
	assert.Empty(t, diag, "got unexpected diagnostics")
}

// TestProviderSchema checks that the RHOAS provider schema is the expected one
func TestProviderSchema(t *testing.T) {
	schemaRequest := terraform.ProviderSchemaRequest{
		ResourceTypes: []string{"rhoas_kafka", "rhoas_topic", "rhoas_service_account", "rhoas_acl"},
		DataSources:   []string{"rhoas_kafka", "rhoas_topic", "rhoas_service_account"},
	}

	providerSchema, err := rhoas.Provider().GetSchema(&schemaRequest)
	assert.NoError(t, err, "unexpected error getting the provider schema")

	t.Run("data sources", func(t *testing.T) {
		sut := providerSchema.DataSources
		assert.Contains(t, sut, "rhoas_kafka")
		assert.Contains(t, sut, "rhoas_topic")
		assert.Contains(t, sut, "rhoas_service_account")
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
	})
}
