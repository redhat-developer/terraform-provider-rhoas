package rhoas

import (
	"context"
	"fmt"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/connection"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/cloudproviders"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/kafkas"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/serviceaccounts"
)

const (
	DefaultAuthURL  = "https://sso.redhat.com/auth/realms/redhat-external"
	DefaultApiUrl = "https://api.openshift.com"
	DefaultClientID = "cloud-services"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema {
			"offline_token" : &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("OFFLINE_TOKEN", nil),
			},
			"auth_url": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("AUTH_URL", DefaultAuthURL),
			},
			"client_id": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("CLIENT_ID", DefaultClientID),
			},
			"api_url": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("API_URL", DefaultApiUrl),
			},
		},
		ResourcesMap:   map[string]*schema.Resource{
			"rhoas_kafka": kafkas.ResourceKafka(),
			"rhoas_service_account": serviceaccounts.ResourceServiceAccount(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rhoas_cloud_providers": cloudproviders.DataSourceCloudProviders(),
			"rhoas_cloud_provider_regions": cloudproviders.DataSourceCloudProviderRegions(),
			"rhoas_kafkas": kafkas.DataSourceKafkas(),
			"rhoas_service_accounts": serviceaccounts.DataSourceServiceAccounts(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	fmt.Printf("HERE!!!")

	c, err := connection.
		NewBuilder().
		WithAuthURL(d.Get("auth_url").(string)).
		WithRefreshToken(d.Get("offline_token").(string)).
		WithClientID(d.Get("client_id").(string)).
		WithURL(d.Get("api_url").(string)).
		Build()

	if err != nil {
		return nil, diag.FromErr(err)
	}

	_, _, err = c.RefreshTokens(ctx)

	if err != nil {
		return nil, diag.FromErr(err)
	}

	return c, diags
}
