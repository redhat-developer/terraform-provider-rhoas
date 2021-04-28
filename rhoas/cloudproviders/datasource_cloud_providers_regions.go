package cloudproviders

import (
	"context"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/redhat-developer/app-services-cli/pkg/connection"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
)

func DataSourceCloudProviderRegions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudProviderRegionsRead,
		Description: "`rhoas_cloud_providers_regions` provides a list of the regions available for Red Hat OpenShift Streams for Apache Kafka.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"regions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Description: "Describes whether the region is enabled",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kind": {
							Type:     schema.TypeString,
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

	c, ok := m.(*connection.KeycloakConnection)
	if !ok {
		return diag.Errorf("unable to cast %v to *connection.KeycloakConnection", m)
	}

	api := c.API().Kafka()

	val := d.Get("id")
	id, ok := val.(string)
	if !ok {
		return diag.Errorf("unable to cast %v to string", val)
	}

	data, resp, err := api.ListCloudProviderRegions(ctx, id).Execute()
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

	if err := d.Set("regions", obj["items"]); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
