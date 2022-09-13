package topics

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
	"io/ioutil"
	"log"
	rhoasAPI "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/api"
	"time"
)

func ResourceTopic() *schema.Resource {
	return &schema.Resource{
		Description:   "`rhoas_topic` manages a topic from a  Kafka instance in Red Hat OpenShift Streams for Apache Kafka.",
		CreateContext: topicCreate,
		ReadContext:   topicRead,
		DeleteContext: topicDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Kafka instance",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"partitions": {
				Description: "The number of partition in the topic",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"kafka_id": {
				Description: "The number of partition in the topic",
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

	api, ok := m.(rhoasAPI.API)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	kafkaID, ok := d.Get("kafka_id").(string)
	if !ok {
		return diag.FromErr(errors.Errorf("There was a problem getting the kafka ID value in the schema resource"))
	}

	topicName, ok := d.Get("name").(string)
	if !ok {
		return diag.FromErr(errors.Errorf("There was a problem getting the topic name value in the schema resource"))
	}

	instanceAPI, _, err := api.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := instanceAPI.TopicsApi.DeleteTopic(ctx, topicName).Execute()
	if err != nil {
		bodyBytes := []byte("empty response")
		if resp != nil {
			var ioErr error
			bodyBytes, ioErr = ioutil.ReadAll(resp.Body)
			if ioErr != nil {
				log.Fatal(ioErr)
			}
		}
		return diag.Errorf("%s %s", err.Error(), string(bodyBytes))
	}

	d.SetId("")
	return diags
}

func topicRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	api, ok := m.(rhoasAPI.API)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasAPI.API", m)
	}

	kafkaID, ok := d.Get("kafka_id").(string)
	if !ok {
		return diag.FromErr(errors.Errorf("There was a problem getting the kafka ID value in the schema resource"))
	}

	topicName, ok := d.Get("name").(string)
	if !ok {
		return diag.FromErr(errors.Errorf("There was a problem getting the topic name value in the schema resource"))
	}

	instanceAPI, _, err := api.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	topic, resp, err := instanceAPI.TopicsApi.GetTopic(ctx, topicName).Execute()
	if err != nil {
		bodyBytes := []byte("empty response")
		if resp != nil {
			var ioErr error
			bodyBytes, ioErr = ioutil.ReadAll(resp.Body)
			if ioErr != nil {
				log.Fatal(ioErr)
			}
		}
		return diag.Errorf("%s %s", err.Error(), string(bodyBytes))
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

	api, ok := m.(rhoasAPI.API)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasAPI.API", m)
	}

	kafkaID, ok := d.Get("kafka_id").(string)
	if !ok {
		return diag.FromErr(errors.Errorf("There was a problem getting the kafka ID value in the schema resource"))
	}

	instanceAPI, _, err := api.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	topicRequest := instanceAPI.TopicsApi.CreateTopic(ctx)

	err = mapResourceDataToTopicRequest(d, &topicRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	topic, resp, err := topicRequest.Execute()
	if err != nil {
		bodyBytes := []byte("empty response")
		if resp != nil {
			var ioErr error
			bodyBytes, ioErr = ioutil.ReadAll(resp.Body)
			if ioErr != nil {
				log.Fatal(ioErr)
			}
		}
		return diag.Errorf("%s%s", err.Error(), string(bodyBytes))
	}

	err = setResourceDataFromTopic(d, &topic)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(topic.GetId())

	if err = d.Set("kafka_id", kafkaID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setResourceDataFromTopic(d *schema.ResourceData, topic *kafkainstanceclient.Topic) error {
	var err error

	if err = d.Set("name", topic.GetName()); err != nil {
		return err
	}

	if err = d.Set("partitions", len(topic.GetPartitions())); err != nil {
		return err
	}

	return nil
}

func mapResourceDataToTopicRequest(d *schema.ResourceData, request *kafkainstanceclient.ApiCreateTopicRequest) error {

	name, ok := d.Get("name").(string)
	if !ok {
		return errors.Errorf("There was a problem getting the name value in the schema resource")
	}

	partitions, ok := d.Get("partitions").(int)
	if !ok {
		return errors.Errorf("There was a problem getting the partition value in the schema resource")
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
