package serviceaccounts

import (
	"context"
	"io/ioutil"
	"log"
	rhoasAPI "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/api"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
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
							Computed:    true,
						},
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The kind of resource in the Clients",
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
