package kafkas

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

func DataSourceKafkas() *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_kafkas` provides a list of the Kafkas accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceKafkasRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The id of Kafka instance",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"kafkas": {
				Description: "The list of Kafka instances",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloud_provider": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cloud provider to use. A list of available cloud providers can be obtained using `data.rhoas_cloud_providers`.",
						},
						"multi_az": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the Kafka instance should be highly available by supporting multi-az",
						},
						"region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region to use. A list of available regions can be obtained using `data.rhoas_cloud_providers_regions`.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the Kafka instance",
						},
						"href": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path to the Kafka instance in the REST Factory",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the Kafka instance",
						},
						"owner": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The username of the Red Hat account that owns the Kafka instance",
						},
						"bootstrap_server_host": {
							Description: "The bootstrap server (host:port)",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"created_at": {
							Description: "The RFC3339 date and time at which the Kafka instance was created",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"updated_at": {
							Description: "The RFC3339 date and time at which the Kafka instance was last updated",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"id": {
							Description: "The unique identifier for the Kafka instance",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The kind of resource in the Factory",
						},
						"version": {
							Description: "The version of Kafka the instance is using",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"failed_reason": {
							Description: "The reason the instance failed",
							Type:        schema.TypeString,
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

	val := d.Get("id")
	id, ok := val.(string)
	if !ok {
		return diag.Errorf("unable to cast %v to string", val)
	}

	kafkas, resp, err := factory.KafkaMgmt().GetKafkas(ctx).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
			return diag.FromErr(apiErr)
		}
	}

	if err := d.Set("kafkas", flattenOrderItemsData(kafkas.Items)); err != nil {
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

func flattenOrderItemsData(kafkas []kafkamgmtclient.KafkaRequest) []interface{} {
	if kafkas != nil {
		ks := make([]interface{}, len(kafkas), len(kafkas))

		for i := range kafkas {
			k := make(map[string]interface{})

			k["cloud_provider"] = kafkas[i].GetCloudProvider()
			k["region"] = kafkas[i].GetRegion()
			k["name"] = kafkas[i].GetName()
			k["href"] = kafkas[i].GetHref()
			k["status"] = kafkas[i].GetStatus()
			k["owner"] = kafkas[i].GetOwner()
			k["bootstrap_server_host"] = kafkas[i].GetBootstrapServerHost()
			k["created_at"] = kafkas[i].GetCreatedAt().Format(time.RFC3339)
			k["updated_at"] = kafkas[i].GetUpdatedAt().Format(time.RFC3339)
			k["id"] = kafkas[i].GetId()
			k["kind"] = kafkas[i].GetKind()
			k["version"] = kafkas[i].GetVersion()

			ks[i] = k
		}

		return ks
	}

	return make([]interface{}, 0)
}
