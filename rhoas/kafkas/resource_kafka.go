package kafkas

import (
	"context"
	kasclient "github.com/bf2fc6cc711aee1a0c2a/cli/pkg/api/kas/client"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/connection"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"strconv"
	"time"
)

func ResourceKafka() *schema.Resource {
	return &schema.Resource{
		CreateContext: kafkaCreate,
		ReadContext: kafkaRead,
		DeleteContext: kafkaDelete,
		Schema: map[string]*schema.Schema{
			"kafka": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloud_provider": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default: "aws",
							ForceNew: true,
						},
						"multi_az": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default: true,
							ForceNew: true,
						},
						"region": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default: "us-east-1",
							ForceNew: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(20 * time.Minute),
		},
	}
}

func filter (items []map[string]interface{}, fields []string) []map[string]interface{} {
	answer := make([]map[string]interface{}, 0)
	for _, item := range items {
		filtered := make(map[string]interface{})
		for k, v := range item {
			keep := false
			for _, f := range fields {
				if f == k {
					keep = true
				}
			}
			if keep {
				filtered[k] = v
			}
		}
		answer = append(answer, filtered)
	}
	return answer
}

func kafkaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	_, _, apiErr := api.DeleteKafkaById(ctx, d.Id()).Async(true).Execute()
	if apiErr.Error() == "404 " {
		// the resource is deleted already
		d.SetId("")
		return diags
	}
	if apiErr.Error() != "" {
		return diag.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
	}

	deleteStateConf := &resource.StateChangeConf{
		Delay:                     5 * time.Second,
		Pending:                   []string{
			"deprovision",
		},
		Refresh:                   func() (interface{}, string, error) {
			data, _, apiErr := api.GetKafkaById(ctx, d.Id()).Execute()
			if apiErr.Error() == "404 " {
				return data, "404",nil
			}
			if apiErr.Error() != "" {
				return nil, "", errors.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
			}
			return data, *data.Status, nil
		},
		Target:                    []string{
			"404",
		},
		Timeout:                   d.Timeout(schema.TimeoutCreate),
		MinTimeout:                5 * time.Second,
		NotFoundChecks:            0,
		ContinuousTargetOccurence: 0,
	}

	_, err := deleteStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "Error waiting for example instance (%s) to be created: %s", d.Id()))
	}

	d.SetId("")
	return diags
}

func kafkaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	var raw []map[string]interface{}

	data, _, apiErr := api.GetKafkaById(ctx, d.Id()).Execute()
	if apiErr.Error() == "404 " {
		d.SetId("")
		return diags
	}
	if apiErr.Error() != "" {
		return diag.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
	}
	obj, err := utils.AsMap(data)
	if err != nil {
		return diag.FromErr(errors.WithStack(err))
	}
	raw = []map[string]interface{}{obj}

	items := fixBootstrapServerHosts(raw)
	if err != nil {
		diag.FromErr(err);
	}
	err = applyRead(items, d)
	if err != nil {
		diag.FromErr(err);
	}

	return diags
}

func applyRead(items []map[string]interface{}, d *schema.ResourceData) error {
	items = filter(items, []string{"cloud_provider", "multi_az", "region","name"})



	if err := d.Set("kafkas", items); err != nil {
		return err
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func kafkaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	items := d.Get("kafka").([]interface{})

	payload := make([]kasclient.KafkaRequestPayload, 0)

	for _, item := range items {
		kafka := item.(map[string]interface{})

		cloudProvider := kafka["cloud_provider"].(string)
		name := kafka["name"].(string)
		multiAz := kafka["multi_az"].(bool)
		region := kafka["region"].(string)

		payload = append(payload, kasclient.KafkaRequestPayload{
			CloudProvider: &cloudProvider,
			MultiAz:       &multiAz,
			Name:          name,
			Region:        &region,
		})
	}

	kr,_, apiErr := api.CreateKafka(ctx).Async(true).KafkaRequestPayload(payload[0]).Execute()

	if apiErr.Error() != "" {
		return diag.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
	}

	if kr.Id == nil {
		return diag.Errorf("no id provided")
	}

	d.SetId(*kr.Id)

	createStateConf := &resource.StateChangeConf{
		Delay:                     5 * time.Second,
		Pending:                   []string{
			"accepted",
			"preparing",
			"provisioning",
		},
		Refresh:                   func() (interface{}, string, error) {
			c := m.(*connection.KeycloakConnection)

			api := c.API().Kafka()

			var raw []map[string]interface{}

			data, _, apiErr := api.GetKafkaById(ctx, *kr.Id).Execute()
			if apiErr.Error() != "" {
				return nil, "", errors.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
			}
			obj, err := utils.AsMap(data)
			if err != nil {
				return nil, "", errors.WithStack(err)
			}
			raw = []map[string]interface{}{obj}

			items := fixBootstrapServerHosts(raw)
			filtered := filter(items, []string{"cloud_provider", "multi_az", "region","name"})
			return filtered, *data.Status, nil
		},
		Target:                    []string{
			"ready",
		},
		Timeout:                   d.Timeout(schema.TimeoutCreate),
		MinTimeout:                5 * time.Second,
		NotFoundChecks:            0,
		ContinuousTargetOccurence: 0,
	}

	data, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "Error waiting for example instance (%s) to be created: %s", d.Id()))
	}
	err = applyRead(data.([]map[string]interface{}), d)
	if err != nil {
		diag.FromErr(err);
	}
	return diags
}
