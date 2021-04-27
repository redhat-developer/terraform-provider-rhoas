package serviceaccounts

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io/ioutil"
	"log"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/cli/connection"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"strconv"
	"time"
)

func DataSourceServiceAccounts() *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_service_accounts` provides a list of the service accounts accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
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
							Description: "The client id associated with the service account",
						},
						"href": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"description": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
							Description: "A description of the service account",
						},
						"id": &schema.Schema{
							Description: "The unique identifier for the service account",
							Type: schema.TypeString,
							Computed: true,
						},
						"kind": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
							Description: "The kind of resource in the API",
						},
						"name": &schema.Schema{
							Description: "The name of the service account",
							Type: schema.TypeString,
							Computed: true,
						},
						"owner": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
							Description: "The username of the Red Hat account that owns the service account",
						},
						"created_at": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
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

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	data, resp, err := api.ListServiceAccounts(ctx).Execute()
	if err != nil {
		bodyBytes, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			log.Fatal(ioErr)
		}
		return diag.Errorf("%s%s", err.Error(), string(bodyBytes))
	}

	obj, err := utils.AsMap(data)
	if err != nil {
		return diag.FromErr(err)
	}

	var raw []map[string]interface{}

	// coerce the type
	for _, item := range obj["items"].([]interface{}) {
		raw = append(raw, item.(map[string]interface{}))
	}

	items := fixClientIDAndClientSecret(raw, nil)

	if err := d.Set("service_accounts", items); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
