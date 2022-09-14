package kafkas

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	rhoasClients "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/clients"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"time"
)

func ResourceKafka() *schema.Resource {
	return &schema.Resource{
		Description:   "`rhoas_kafka` manages a Kafka instance in Red Hat OpenShift Streams for Apache Kafka.",
		CreateContext: kafkaCreate,
		ReadContext:   kafkaRead,
		DeleteContext: kafkaDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"cloud_provider": {
				Description: "The cloud provider to use. A list of available cloud providers can be obtained using `data.rhoas_cloud_providers`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "aws",
				ForceNew:    true,
			},
			"region": {
				Description: "The region to use. A list of available regions can be obtained using `data.rhoas_cloud_providers_regions`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "us-east-1",
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the Kafka instance",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
				Computed:    true,
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
		},
	}
}

func kafkaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, ok := m.(*rhoasClients.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	apiErr, _, err := c.KafkaClient.DefaultApi.DeleteKafkaById(ctx, d.Id()).Async(true).Execute()
	if err != nil && err.Error() == "404 " {
		// the resource is deleted already
		d.SetId("")
		return diags
	}
	if err != nil {
		if apiErr.Reason != "" {
			return diag.Errorf("%s%s", err.Error(), apiErr.Reason)
		}
		return diag.Errorf("%s", err.Error())
	}

	deleteStateConf := &resource.StateChangeConf{
		Delay: 5 * time.Second,
		Pending: []string{
			"deprovision", "deleting",
		},
		Refresh: func() (interface{}, string, error) {
			data, resp, err1 := c.KafkaClient.DefaultApi.GetKafkaById(ctx, d.Id()).Execute()
			if err1 != nil {
				apiError, err2 := utils.GetAPIError(resp, err1)
				if err2 != nil {
					return nil, "", err2
				}

				return nil, "", apiError
			}
			return data, *data.Status, nil
		},
		Target: []string{
			"deleted", "404",
		},
		Timeout:                   d.Timeout(schema.TimeoutCreate),
		MinTimeout:                5 * time.Second,
		NotFoundChecks:            0,
		ContinuousTargetOccurence: 0,
	}

	_, err = deleteStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "Error waiting for example instance (%s) to be deleted", d.Id()))
	}

	d.SetId("")
	return diags
}

func kafkaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	c, ok := m.(*rhoasClients.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	kafka, resp, err := c.KafkaClient.DefaultApi.GetKafkaById(ctx, d.Id()).Execute()
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

func kafkaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, ok := m.(*rhoasClients.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	requestPayload, err := mapResourceDataToKafkaPayload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	kr, resp, err := c.KafkaClient.DefaultApi.CreateKafka(ctx).Async(true).KafkaRequestPayload(*requestPayload).Execute()
	if err != nil {
		apiError, err1 := utils.GetAPIError(resp, err)
		if err1 != nil {
			return diag.FromErr(err1)
		}

		return diag.FromErr(apiError)
	}

	if kr.Id == "" {
		return diag.Errorf("no id provided")
	}

	d.SetId(kr.Id)

	createStateConf := &resource.StateChangeConf{
		Delay: 5 * time.Second,
		Pending: []string{
			"accepted",
			"preparing",
			"provisioning",
		},
		Refresh: func() (interface{}, string, error) {
			c, ok := m.(*rhoasClients.Clients)
			if !ok {
				return nil, "", errors.Errorf("unable to cast %v to *rhoasClients.Clients", m)
			}

			data, resp, err1 := c.KafkaClient.DefaultApi.GetKafkaById(ctx, kr.Id).Execute()
			if err1 != nil {
				apiError, err2 := utils.GetAPIError(resp, err1)
				if err2 != nil {
					return nil, "", err2
				}

				return nil, "", apiError
			}
			obj, err1 := utils.AsMap(data)
			if err1 != nil {
				return nil, "", errors.WithStack(err1)
			}
			// raw := []map[string]interface{}{obj}

			return obj, *data.Status, nil
		},
		Target: []string{
			"ready",
		},
		Timeout:                   d.Timeout(schema.TimeoutCreate),
		MinTimeout:                5 * time.Second,
		NotFoundChecks:            0,
		ContinuousTargetOccurence: 0,
	}

	data, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "Error waiting for instance (%s) to be created", d.Id()))
	}

	kafkaData, castOk := data.(map[string]interface{})
	if !castOk {
		return diag.Errorf("Cannot cast data from kafka creation to to map[string]interface{}")
	}

	err = setResourceDataFromKafkaData(d, &kafkaData)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setResourceDataFromKafkaData(d *schema.ResourceData, kafkaData *map[string]interface{}) error {
	var err error

	if err = d.Set("cloud_provider", (*kafkaData)["cloud_provider"]); err != nil {
		return err
	}

	if err = d.Set("region", (*kafkaData)["region"]); err != nil {
		return err
	}

	if err = d.Set("name", (*kafkaData)["name"]); err != nil {
		return err
	}

	if err = d.Set("href", (*kafkaData)["href"]); err != nil {
		return err
	}

	if err = d.Set("status", (*kafkaData)["status"]); err != nil {
		return err
	}

	if err = d.Set("owner", (*kafkaData)["owner"]); err != nil {
		return err
	}

	if err = d.Set("bootstrap_server_host", (*kafkaData)["bootstrap_server_host"]); err != nil {
		return err
	}

	if err = d.Set("created_at", (*kafkaData)["created_at"]); err != nil {
		return err
	}

	if err = d.Set("updated_at", (*kafkaData)["updated_at"]); err != nil {
		return err
	}

	if err = d.Set("id", (*kafkaData)["id"]); err != nil {
		return err
	}

	if err = d.Set("kind", (*kafkaData)["kind"]); err != nil {
		return err
	}

	if err = d.Set("version", (*kafkaData)["version"]); err != nil {
		return err
	}

	return nil
}

func mapResourceDataToKafkaPayload(d *schema.ResourceData) (*kafkamgmtclient.KafkaRequestPayload, error) {

	// we only set these values from the resource data as all the rest are set as
	// computed in the schema and for us the computed values are assigned when we
	// get the kafka request object back from the API
	cloudProvider, ok := d.Get("cloud_provider").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the cloud provider value in the schema resource")
	}

	region, ok := d.Get("region").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the region value in the schema resource")
	}

	name, ok := d.Get("name").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the name value in the schema resource")
	}

	payload := kafkamgmtclient.NewKafkaRequestPayload(name)

	payload.SetCloudProvider(cloudProvider)
	payload.SetRegion(region)

	return payload, nil
}
