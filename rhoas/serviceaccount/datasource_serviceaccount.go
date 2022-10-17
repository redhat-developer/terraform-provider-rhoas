package serviceaccount

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

func DataSourceServiceAccount(localizer localize.Localizer) *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_service_account` provides a service account accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceServiceAccountRead,
		Schema: map[string]*schema.Schema{
			IDField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.id"),
				Type:        schema.TypeString,
				Required:    true,
			},
			DescriptionField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.description"),
				Type:        schema.TypeString,
				Computed:    true,
			},
			NameField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.name"),
				Type:        schema.TypeString,
				Computed:    true,
			},
			ClientIDField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.clientID"),
				Type:        schema.TypeString,
				Computed:    true,
			},
			CreatedByField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.createdBy"),
				Type:        schema.TypeString,
				Computed:    true,
			},
			CreatedAtField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.createdAt"),
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataSourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory)", m)
	}

	id, ok := d.Get(IDField).(string)
	if !ok {
		return diag.FromErr(factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", IDField)))
	}

	serviceAccount, resp, err := factory.ServiceAccountMgmt().GetServiceAccount(ctx, id).Execute()
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
