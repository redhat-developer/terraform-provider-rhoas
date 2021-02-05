package serviceaccounts

import (
	"context"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/connection"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"strconv"
	"time"
)

func DataSourceServiceAccounts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKafkasRead,
		Schema: map[string]*schema.Schema{
			"service_accounts": &schema.Schema{
				Type: schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"href": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"description": &schema.Schema{
							Type: schema.TypeString,
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
						"name": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceKafkasRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	data, _, apiErr := api.ListServiceAccounts(ctx).Execute()
	if apiErr.Error() != "" {
		return diag.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
	}

	obj, err := utils.AsMap(data)
	if err != nil {
		return diag.FromErr(apiErr)
	}

	var raw []map[string]interface{}

	// coerce the type
	for _, item := range obj["items"].([]interface{}) {
		raw = append(raw, item.(map[string]interface{}))
	}

	items := fixClientIDAndClientSecret(raw)

	if err := d.Set("service_accounts", items); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
