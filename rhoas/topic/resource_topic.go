package topic

import (
	"context"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

const (
	NameField       = "name"
	PartitionsField = "partitions"
	KafkaIDField    = "kafka_id"
)

func ResourceTopic(localizer localize.Localizer) *schema.Resource {
	return &schema.Resource{
		Description:   "`rhoas_topic` manages a topic in a  Kafka instance in Red Hat OpenShift Streams for Apache Kafka.",
		CreateContext: topicCreate,
		ReadContext:   topicRead,
		DeleteContext: topicDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			NameField: {
				Description: localizer.MustLocalize("topic.resource.field.description.name"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			PartitionsField: {
				Description: localizer.MustLocalize("topic.resource.field.description.partitions"),
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			KafkaIDField: {
				Description: localizer.MustLocalize("topic.resource.field.description.kafkaID"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func topicDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	kafkaID, ok := d.Get(KafkaIDField).(string)
	if !ok {
		return diag.FromErr(factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", KafkaIDField)))

	}

	topicName, ok := d.Get(NameField).(string)
	if !ok {
		return diag.FromErr(factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", NameField)))
	}

	instanceAPI, _, err := factory.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := instanceAPI.TopicsApi.DeleteTopic(ctx, topicName).Execute()
	if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
		return diag.FromErr(apiErr)
	}

	d.SetId("")
	return diags
}

func topicRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	kafkaID, ok := d.Get(KafkaIDField).(string)
	if !ok {
		return diag.FromErr(factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", KafkaIDField)))
	}

	topicName, ok := d.Get(NameField).(string)
	if !ok {
		return diag.FromErr(factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", NameField)))
	}

	instanceAPI, _, err := factory.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	topic, resp, err := instanceAPI.TopicsApi.GetTopic(ctx, topicName).Execute()
	if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
		return diag.FromErr(apiErr)
	}

	err = setResourceDataFromTopic(d, &topic)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func topicCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
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

	topicRequest := instanceAPI.TopicsApi.CreateTopic(ctx)

	err = mapResourceDataToTopicRequest(factory, d, &topicRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	topic, resp, err := topicRequest.Execute()
	if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
		return diag.FromErr(apiErr)
	}

	err = setResourceDataFromTopic(d, &topic)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(topic.GetId())

	if err = d.Set(KafkaIDField, kafkaID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setResourceDataFromTopic(d *schema.ResourceData, topic *kafkainstanceclient.Topic) error {
	var err error

	if err = d.Set(NameField, topic.GetName()); err != nil {
		return err
	}

	if err = d.Set(PartitionsField, len(topic.GetPartitions())); err != nil {
		return err
	}

	return nil
}

func mapResourceDataToTopicRequest(factory rhoasAPI.Factory, d *schema.ResourceData, request *kafkainstanceclient.ApiCreateTopicRequest) error {

	name, ok := d.Get(NameField).(string)
	if !ok {
		return factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", NameField))
	}

	partitions, ok := d.Get(PartitionsField).(int)
	if !ok {
		return factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", PartitionsField))
	}

	// terraform stores int types as just ints and fails if you
	// try get a value as int32, so need to cast to int32 here
	// as SDK requires int32
	partitionsInt32 := int32(partitions)

	topicInput := kafkainstanceclient.NewTopicInput{
		Name: name,
		Settings: kafkainstanceclient.TopicSettings{
			NumPartitions: &partitionsInt32,
		},
	}

	*request = request.NewTopicInput(topicInput)

	return nil
}
