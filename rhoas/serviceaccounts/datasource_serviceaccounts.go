package serviceaccounts

import (
	"context"
	"io"
	"log"
	"strconv"
	"time"

	rhoasAPI "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	serviceaccountsclient "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
)

func DataSourceServiceAccounts() *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_service_accounts` provides a list of the service accounts accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceKafkasRead,
		Schema: map[string]*schema.Schema{
			"service_accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The client id associated with the service account",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A description of the service account",
						},
						"id": {
							Description: "The unique identifier for the service account",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "The name of the service account",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"created_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The username of the Red Hat account that owns the service account",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The RFC3339 date and time at which the service account was created",
						},
					},
				},
			},
		},
	}
}

func dataSourceKafkasRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	api, ok := m.(rhoasAPI.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	data, resp, err := api.ServiceAccountMgmt().GetServiceAccounts(ctx).Execute()
	if err != nil {
		bodyBytes, ioErr := io.ReadAll(resp.Body)
		if ioErr != nil {
			log.Fatal(ioErr)
		}
		return diag.Errorf("%s%s", err.Error(), string(bodyBytes))
	}

	if err := d.Set("service_accounts", flattenServiceAccountData(data)); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenServiceAccountData(serviceAccounts []serviceaccountsclient.ServiceAccountData) []interface{} {
	if serviceAccounts != nil {
		sas := make([]interface{}, len(serviceAccounts), len(serviceAccounts))

		for i := range serviceAccounts {
			s := make(map[string]interface{})

			s["client_id"] = serviceAccounts[i].GetClientId()
			s["description"] = serviceAccounts[i].GetDescription()
			s["id"] = serviceAccounts[i].GetId()
			s["name"] = serviceAccounts[i].GetName()
			s["created_by"] = serviceAccounts[i].GetCreatedBy()
			s["created_at"] = serviceAccounts[i].GetCreatedAt()

			sas[i] = s
		}

		return sas
	}

	return make([]interface{}, 0)
}
