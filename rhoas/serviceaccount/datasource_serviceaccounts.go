package serviceaccount

import (
	"context"
	"io"
	"log"
	"strconv"
	"time"

	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	serviceaccountsclient "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
)

func DataSourceServiceAccounts(localizer localize.Localizer) *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_service_accounts` provides a list of the service accounts accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceKafkasRead,
		Schema: map[string]*schema.Schema{
			"service_accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
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
				},
			},
		},
	}
}

func dataSourceKafkasRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasAPI.Factory", m)
	}

	data, resp, err := factory.ServiceAccountMgmt().GetServiceAccounts(ctx).Execute()
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

			s[ClientIDField] = serviceAccounts[i].GetClientId()
			s[DescriptionField] = serviceAccounts[i].GetDescription()
			s[IDField] = serviceAccounts[i].GetId()
			s[NameField] = serviceAccounts[i].GetName()
			s[CreatedByField] = serviceAccounts[i].GetCreatedBy()
			s[CreatedAtField] = serviceAccounts[i].GetCreatedAt()

			sas[i] = s
		}

		return sas
	}

	return make([]interface{}, 0)
}
