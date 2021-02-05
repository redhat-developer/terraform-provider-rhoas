package cloudproviders

import (
	"context"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/connection"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"strconv"
	"time"
)

func DataSourceCloudProviderRegions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudProviderRegionsRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"regions": &schema.Schema{
				Type: schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"enabled": &schema.Schema{
							Type: schema.TypeBool,
							Computed: true,
						},
						"id": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"kind": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudProviderRegionsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	id := d.Get("id").(string)

	data, _, apiErr := api.ListCloudProviderRegions(ctx, id).Execute()
	if apiErr.Error() != "" {
		return diag.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
	}

	obj, err := utils.AsMap(data)
	if err != nil {
		return diag.FromErr(apiErr)
	}

	if err := d.Set("regions", obj["items"]); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
