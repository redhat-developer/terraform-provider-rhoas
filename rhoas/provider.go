package rhoas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/cloudproviders"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/kafkas"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/serviceaccounts"
)

const (
	DefaultAuthURL    = "https://sso.redhat.com/auth/realms/redhat-external"
	DefaultMasAuthURL = "https://identity.api.openshift.com/auth/realms/rhoas-client-prod"
	DefaultAPIURL     = "https://api.openshift.com"
	DefaultClientID   = "cloud-services"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"offline_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OFFLINE_TOKEN", nil),
				Description: "The offline token is a refresh token with no expiry and can be used by non-interactive processes to provide an access token for Red Hat OpenShift Application Services. The offline token can be obtained from [https://cloud.redhat.com/openshift/token](https://cloud.redhat.com/openshift/token). As the offline token is a sensitive value that varies between environments it is best specified using the `OFFLINE_TOKEN` environment variable.",
			},
			"auth_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AUTH_URL", DefaultAuthURL),
				Description: fmt.Sprintf("The auth url is used to get an access token for the service by passing the offline token. By default production is used (%s).", DefaultAuthURL),
			},
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLIENT_ID", DefaultClientID),
				Description: fmt.Sprintf("The client id is used to when getting the access token using the offline token. By default %s is used.", DefaultClientID),
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("API_URL", DefaultAPIURL),
				Description: fmt.Sprintf("URL to the RHOAS services API. By default using production API (%s).", DefaultAPIURL),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"rhoas_kafka":           kafkas.ResourceKafka(),
			"rhoas_service_account": serviceaccounts.ResourceServiceAccount(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rhoas_cloud_providers":        cloudproviders.DataSourceCloudProviders(),
			"rhoas_cloud_provider_regions": cloudproviders.DataSourceCloudProviderRegions(),
			"rhoas_kafkas":                 kafkas.DataSourceKafkas(),
			"rhoas_service_accounts":       serviceaccounts.DataSourceServiceAccounts(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	c := BuildKasAPIClient(d.Get("offline_token").(string), d.Get("client_id").(string), d.Get("auth_url").(string))
	return c, diags
}
