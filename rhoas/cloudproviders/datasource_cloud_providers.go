package cloudproviders

import (
	"context"
	"io"
	"log"
	"strconv"
	"time"

	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

func DataSourceCloudProviders() *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_cloud_providers` provides a list of the cloud providers available for Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceCloudProvidersRead,
		Schema: map[string]*schema.Schema{
			"cloud_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kind": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudProvidersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	c, ok := m.(*kafkamgmtclient.APIClient)
	if !ok {
		return diag.Errorf("unable to cast %v to *connection.KeycloakConnection", m)
	}

	data, resp, err := c.DefaultApi.GetCloudProviders(ctx).Execute()
	if err != nil {
		bodyBytes, ioErr := io.ReadAll(resp.Body)
		if ioErr != nil {
			log.Fatal(ioErr)
		}
		return diag.Errorf("%s%s", err.Error(), string(bodyBytes))
	}

	obj, err := utils.AsMap(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cloud_providers", obj["items"]); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
