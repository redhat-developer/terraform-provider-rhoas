package cloudproviders

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

func DataSourceCloudProviderRegions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudProviderRegionsRead,
		Description: "`rhoas_cloud_provider_regions` provides a list of the regions available for Red Hat OpenShift Streams for Apache Kafka.",
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

	api, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	val := d.Get("id")
	id, ok := val.(string)
	if !ok {
		return diag.Errorf("unable to cast %v to string", val)
	}

	data, resp, err := api.KafkaMgmt().GetCloudProviderRegions(ctx, id).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
			return diag.FromErr(apiErr)
		}
	}

	if err := d.Set("regions", flattenRegions(data.Items)); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenRegions(cloudRegionList []kafkamgmtclient.CloudRegion) []interface{} {
	if cloudRegionList != nil {
		cloudRegions := make([]interface{}, len(cloudRegionList), len(cloudRegionList))

		for i := range cloudRegionList {
			cloudRegion := make(map[string]interface{})

			cloudRegion["display_name"] = cloudRegionList[i].GetDisplayName()
			cloudRegion["enabled"] = cloudRegionList[i].GetEnabled()
			cloudRegion["id"] = cloudRegionList[i].GetId()
			cloudRegion["kind"] = cloudRegionList[i].GetKind()

			cloudRegions[i] = cloudRegion
		}

		return cloudRegions
	}

	return make([]interface{}, 0)
}
