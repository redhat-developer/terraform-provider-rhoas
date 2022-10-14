package rhoas

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/acl"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize/goi18n"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	authAPI "github.com/redhat-developer/app-services-sdk-go/auth/apiv1"
	kafkamgmt "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1"
	serviceAccounts "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
	factories "github.com/redhat-developer/terraform-provider-rhoas/rhoas/factory"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/kafka"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/serviceaccount"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/topic"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

const (
	DefaultAPIURL = "https://api.openshift.com"
)

// Provider -
func Provider() *schema.Provider {

	localizer, err := goi18n.New(nil)
	if err != nil {
		return nil
	}

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
				DefaultFunc: schema.EnvDefaultFunc("AUTH_URL", authAPI.DefaultAuthURL),
				Description: fmt.Sprintf("The auth url is used to get an access token for the service by passing the offline token. By default production is used (%s).", authAPI.DefaultAuthURL),
			},
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLIENT_ID", authAPI.DefaultClientID),
				Description: fmt.Sprintf("The client id is used to when getting the access token using the offline token. By default %s is used.", authAPI.DefaultClientID),
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("API_URL", DefaultAPIURL),
				Description: fmt.Sprintf("URL to the RHOAS services API. By default using production API (%s).", DefaultAPIURL),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"rhoas_kafka":           kafka.ResourceKafka(localizer),
			"rhoas_topic":           topic.ResourceTopic(localizer),
			"rhoas_service_account": serviceaccount.ResourceServiceAccount(localizer),
			"rhoas_acl":             acl.ResourceACL(localizer),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rhoas_kafka":           kafka.DataSourceKafka(localizer),
			"rhoas_topic":           topic.DataSourceTopic(localizer),
			"rhoas_service_account": serviceaccount.DataSourceServiceAccount(localizer),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	localizer, err := goi18n.New(nil)
	if err != nil {
		tflog.Error(ctx, err.Error())
		os.Exit(1)
	}

	httpClient := authAPI.BuildAuthenticatedHTTPClient(d.Get("offline_token").(string))

	kafkaClient := kafkamgmt.NewAPIClient(&kafkamgmt.Config{
		HTTPClient: httpClient,
	})

	config := serviceAccounts.NewConfiguration()
	config.HTTPClient = httpClient
	serviceAccountClient := serviceAccounts.NewAPIClient(config)

	// package both service account client and kafka client together to be used in the provider
	// these are passed to each action we do and can be use to CRUD kafkas/serviceAccounts
	factory := factories.NewDefaultFactory(kafkaClient, serviceAccountClient, httpClient, localizer)

	return factory, diags
}
