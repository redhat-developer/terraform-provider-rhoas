package kafkas

import (
	"context"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/connection"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"strconv"
	"time"
)

func DataSourceKafkas() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKafkasRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default: "",
			},
			"kafkas": &schema.Schema{
				Type: schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"href": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"status": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"cloud_provider": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"multi_az": &schema.Schema{
							Type: schema.TypeBool,
							Computed: true,
						},
						"region": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"owner": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"bootstrap_server_host": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"created_at": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"updated_at": &schema.Schema{
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

	var raw []map[string]interface{}

	id := d.Get("id").(string)

	data, _, apiErr := api.ListKafkas(ctx).Execute()
	if apiErr.Error() != "" {
		return diag.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
	}
	obj, err := utils.AsMap(data)
	if err != nil {
		return diag.FromErr(errors.WithStack(err))
	}

	// coerce the type
	for _, item := range obj["items"].([]interface{}) {
		raw = append(raw, item.(map[string]interface{}))
	}

	items := fixBootstrapServerHosts(raw)

	if err := d.Set("kafkas", items); err != nil {
		return diag.FromErr(err)
	}

	if id == "" {
		// use the current timestamp for a list request to force a refresh
		id = strconv.FormatInt(time.Now().Unix(), 10)
	}

	// always run
	d.SetId(id)

	return diags
}
