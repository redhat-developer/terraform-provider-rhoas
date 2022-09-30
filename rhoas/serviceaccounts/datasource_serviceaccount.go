package serviceaccounts

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rhoasAPI "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/api"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
)

func DataSourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_service_account` provides a service account accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceServiceAccountRead,
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The client id associated with the service account",
			},
			"href": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A description of the service account",
			},
			"id": {
				Description: "The unique identifier for the service account",
				Type:        schema.TypeString,
				Required:    true,
			},
			"kind": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The kind of resource in the API",
			},
			"name": {
				Description: "The name of the service account",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"owner": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The username of the Red Hat account that owns the service account",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The RFC3339 date and time at which the service account was created",
			},
		},
	}
}

func dataSourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	api, ok := m.(rhoasAPI.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Clients)", m)
	}

	val := d.Get("id")
	id, ok := val.(string)
	if !ok {
		return diag.Errorf("unable to cast %v to string for use as for service account id", val)
	}

	serviceAccount, resp, err := api.ServiceAccountMgmt().GetServiceAccount(ctx, id).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
			return diag.FromErr(apiErr)
		}
	}

	err = setResourceDataFromServiceAccountData(d, &serviceAccount)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serviceAccount.GetId())

	return diags
}
