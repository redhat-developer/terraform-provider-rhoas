package topic

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

func DataSourceTopic(localizer localize.Localizer) *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_topic` provides a Topic accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceTopicRead,
		Schema: map[string]*schema.Schema{
			NameField: {
				Description: localizer.MustLocalize("topic.resource.field.description.name"),
				Type:        schema.TypeString,
				Required:    true,
			},
			PartitionsField: {
				Description: localizer.MustLocalize("topic.resource.field.description.partitions"),
				Type:        schema.TypeInt,
				Computed:    true,
			},
			KafkaIDField: {
				Description: localizer.MustLocalize("topic.resource.field.description.kafkaID"),
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceTopicRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	kafkaID, ok := d.Get(KafkaIDField).(string)
	if !ok {
		return diag.FromErr(factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", KafkaIDField)))
	}

	instanceAPI, _, err := factory.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	name, ok := d.Get(KafkaIDField).(string)
	if !ok {
		return diag.FromErr(factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", NameField)))
	}

	topic, resp, err := instanceAPI.TopicsApi.GetTopic(ctx, name).Execute()
	if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
		return diag.FromErr(apiErr)
	}

	err = setResourceDataFromTopic(d, &topic)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
