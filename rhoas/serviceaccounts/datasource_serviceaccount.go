package serviceaccounts

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

func DataSourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_service_account` provides a service account accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceServiceAccountRead,
		Schema: map[string]*schema.Schema{
			IDField: {
				Description: "The unique id fir the service account",
				Type:        schema.TypeString,
				Required:    true,
			},
			DescriptionField: {
				Description: "A description of the service account",
				Type:        schema.TypeString,
				Computed:    true,
			},
			NameField: {
				Description: "The name of the service account",
				Type:        schema.TypeString,
				Computed:    true,
			},
			ClientIDField: {
				Description: "The client id associated with the service account",
				Type:        schema.TypeString,
				Computed:    true,
			},
			ClientSecret: {
				Description: "The client secret associated with the service account. It must be stored by the client as the server will not return it after creation",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	api, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory)", m)
	}

	id, ok := d.Get(IDField).(string)
	if !ok {
		return diag.FromErr(errors.Errorf("Could not retrieve client id in service account data source"))
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

	d.SetId(id)

	return diags
}
