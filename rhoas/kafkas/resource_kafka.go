package kafkas

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	rhoasAPI "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/api"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
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

	api, ok := m.(rhoasAPI.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Clients)", m)
	}

	apiErr, _, err := api.KafkaMgmt().DeleteKafkaById(ctx, d.Id()).Async(true).Execute()
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
			data, resp, err1 := api.KafkaMgmt().GetKafkaById(ctx, d.Id()).Execute()
			if err1 != nil {
				if err1.Error() == "404 Not Found" {
					return data, "404", nil
				}
				if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
					return nil, "", apiErr
				}
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
		if !strings.Contains(err.Error(), "not found") {
			return diag.FromErr(errors.Wrapf(err, "Error waiting for example instance (%s) to be deleted", d.Id()))
		}
	}

	d.SetId("")
	return diags
}

func kafkaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	api, ok := m.(rhoasAPI.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Clients)", m)
	}

	kafka, resp, err := api.KafkaMgmt().GetKafkaById(ctx, d.Id()).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
			return diag.FromErr(apiErr)
		}
	}

	err = setResourceDataFromKafkaData(d, &kafka)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func kafkaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	api, ok := m.(rhoasAPI.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Clients)", m)
	}

	requestPayload, err := mapResourceDataToKafkaPayload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	kr, resp, err := api.KafkaMgmt().CreateKafka(ctx).Async(true).KafkaRequestPayload(*requestPayload).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
			return diag.FromErr(apiErr)
		}
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
			api, ok := m.(rhoasAPI.Clients)
			if !ok {
				return nil, "", errors.Errorf("unable to cast %v to rhoasAPI.Clients)", m)
			}

			kafka, resp, err1 := api.KafkaMgmt().GetKafkaById(ctx, kr.Id).Execute()
			if err1 != nil {
				if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
					return nil, "", apiErr
				}
			}

			return kafka, kafka.GetStatus(), nil
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

	kafka, castOk := data.(kafkamgmtclient.KafkaRequest)
	if !castOk {
		return diag.Errorf("Cannot cast data from kafka creation to to map[string]interface{}")
	}

	err = setResourceDataFromKafkaData(d, &kafka)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setResourceDataFromKafkaData(d *schema.ResourceData, kafka *kafkamgmtclient.KafkaRequest) error {
	var err error

	if err = d.Set("cloud_provider", kafka.GetCloudProvider()); err != nil {
		return err
	}

	if err = d.Set("region", kafka.GetRegion()); err != nil {
		return err
	}

	if err = d.Set("name", kafka.GetName()); err != nil {
		return err
	}

	if err = d.Set("href", kafka.GetHref()); err != nil {
		return err
	}

	if err = d.Set("status", kafka.GetStatus()); err != nil {
		return err
	}

	if err = d.Set("owner", kafka.GetOwner()); err != nil {
		return err
	}

	if err = d.Set("bootstrap_server_host", kafka.GetBootstrapServerHost()); err != nil {
		return err
	}

	if err = d.Set("created_at", kafka.GetCreatedAt().Format(time.RFC3339)); err != nil {
		return err
	}

	if err = d.Set("updated_at", kafka.GetUpdatedAt().Format(time.RFC3339)); err != nil {
		return err
	}

	if err = d.Set("id", kafka.GetId()); err != nil {
		return err
	}

	if err = d.Set("kind", kafka.GetKind()); err != nil {
		return err
	}

	if err = d.Set("version", kafka.GetVersion()); err != nil {
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
