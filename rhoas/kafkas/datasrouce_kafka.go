package kafkas

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	rhoasAPI "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/api"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
)

func DataSourceKafka() *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_kafka` provides a Kafka accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceKafkaRead,
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
				Description: "The path to the Kafka instance in the REST API",
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
				Required:    true,
			},
			"kind": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The kind of resource in the API",
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
	}
}

func dataSourceKafkaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	api, ok := m.(rhoasAPI.API)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	val := d.Get("id")
	id, ok := val.(string)
	if !ok {
		return diag.Errorf("unable to cast %v to string for use as for kafka id", val)
	}

	kafka, resp, err := api.KafkaMgmt().GetKafkaById(ctx, id).Execute()
	if err != nil {
		apiError, err1 := utils.GetAPIError(resp, err)
		if err1 != nil {
			return diag.FromErr(err1)
		}

		return diag.FromErr(apiError)
	}
	kafkaData, err := utils.AsMap(kafka)
	if err != nil {
		return diag.FromErr(errors.WithStack(err))
	}

	err = setResourceDataFromKafkaData(d, &kafkaData)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
