package topics

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

func DataSourceTopic() *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_topic` provides a Topic accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceTopicRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Kafka topic",
			},
			"partitions": {
				Description: "The number of partitions in the topic",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"kafka_id": {
				Description: "The unique identifier for the Kafka instance",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceTopicRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	api, ok := m.(rhoasAPI.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Clients)", m)
	}

	val := d.Get("kafka_id")
	kafkaID, ok := val.(string)
	if !ok {
		return diag.Errorf("unable to cast %v to string for use as for kafka_id", val)
	}

	instanceAPI, _, err := api.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	val = d.Get("name")
	name, ok := val.(string)
	if !ok {
		return diag.Errorf("unable to cast %v to string for use as for topic name", val)
	}

	topic, resp, err := instanceAPI.TopicsApi.GetTopic(ctx, name).Execute()
	if err != nil {
		err2 := utils.GetAPIError(resp, err)
		if err2 != nil {
			return diag.FromErr(err2)
		}
	}

	err = setResourceDataFromTopic(d, &topic)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
