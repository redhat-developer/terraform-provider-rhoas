package rhoas

import (
	"context"
	kafkamgmt "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/acl"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize/goi18n"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	authAPI "github.com/redhat-developer/app-services-sdk-go/auth/apiv1"
	serviceAccounts "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
	factories "github.com/redhat-developer/terraform-provider-rhoas/rhoas/factory"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/kafka"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/serviceaccount"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/topic"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

const (
	mockAlias       = "mock"
	prodAlias       = "prod"
	stageAlias      = "stage"
	offlineTokenENV = "OFFLINE_TOKEN"
)

var (
	ProductionAPIURL  = "https://api.openshift.com"
	StagingAPIURL     = "https://api.stage.openshift.com"
	ProductionAuthURL = "https://sso.redhat.com/auth/realms/redhat-external"
	StagingAuthURL    = "https://sso.stage.redhat.com/auth/realms/redhat-external"
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
				DefaultFunc: schema.EnvDefaultFunc(offlineTokenENV, nil),
				Description: "The offline token is a refresh token with no expiry and can be used by non-interactive processes to provide an access token for Red Hat OpenShift Application Services. The offline token can be obtained from [https://cloud.redhat.com/openshift/token](https://cloud.redhat.com/openshift/token). As the offline token is a sensitive value that varies between environments it is best specified using the `OFFLINE_TOKEN` environment variable.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"rhoas_kafka":           kafka.ResourceKafka(localizer),
			"rhoas_topic":           topic.ResourceTopic(localizer),
			"rhoas_service_account": serviceaccount.ResourceServiceAccount(localizer),
			"rhoas_acl":             acl.ResourceACL(localizer),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rhoas_kafkas":           kafka.DataSourceKafkas(localizer),
			"rhoas_service_accounts": serviceaccount.DataSourceServiceAccounts(localizer),
			"rhoas_kafka":            kafka.DataSourceKafka(localizer),
			"rhoas_topic":            topic.DataSourceTopic(localizer),
			"rhoas_service_account":  serviceaccount.DataSourceServiceAccount(localizer),
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

	baseURLsForAlias := map[string]string{
		mockAlias:  "http://localhost:8000",
		stageAlias: StagingAPIURL,
		prodAlias:  ProductionAPIURL,
	}

	authURLsForAlias := map[string]string{
		mockAlias:  "",
		stageAlias: StagingAuthURL, // NOTE: stage uses the production auth URL also
		prodAlias:  ProductionAuthURL,
	}

	apiAlias := os.Getenv("API")
	if apiAlias == "" {
		apiAlias = prodAlias
	}

	var offlineToken string

	if token, ok := d.Get("offline_token").(string); ok {
		offlineToken = token
	} else {
		offlineToken = os.Getenv(offlineTokenENV)
	}

	httpClient := &http.Client{}
	if apiAlias != "mock" {
		// nolint: contextcheck
		httpClient = authAPI.BuildAuthenticatedHTTPClientCustom(offlineToken, authAPI.DefaultClientID, authURLsForAlias[prodAlias])
	}

	kafkaClient := kafkamgmt.NewAPIClient(&kafkamgmt.Config{
		HTTPClient: httpClient,
		BaseURL:    baseURLsForAlias[apiAlias],
	})

	serviceAccountConfig := serviceAccounts.NewConfiguration()

	if apiAlias == mockAlias {
		serviceAccountConfig.Servers = serviceAccounts.ServerConfigurations{
			{
				URL:         baseURLsForAlias[apiAlias],
				Description: "RHOAS mock server",
			},
		}
	}

	serviceAccountConfig.HTTPClient = httpClient
	serviceAccountClient := serviceAccounts.NewAPIClient(serviceAccountConfig)

	// package both service account client and kafka client together to be used in the provider
	// these are passed to each action we do and can be use to CRUD kafkas/serviceAccounts
	factory := factories.NewDefaultFactory(kafkaClient, serviceAccountClient, httpClient, localizer)

	return factory, diags
}
