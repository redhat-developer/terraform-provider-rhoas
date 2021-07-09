package kafkas

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	"io/ioutil"
	"log"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"strconv"
	"time"
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

	c, ok := m.(*kafkamgmtclient.APIClient)
	if !ok {
		return diag.Errorf("unable to cast %v to *connection.KeycloakConnection", m)
	}

	var raw []map[string]interface{}

	val := d.Get("id")
	id, ok := val.(string)
	if !ok {
		return diag.Errorf("unable to cast %v to string", val)
	}

	data, resp, err := c.DefaultApi.GetKafkas(ctx).Execute()
	if err != nil {
		bodyBytes, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			log.Fatal(ioErr)
		}
		return diag.Errorf("%s%s", err.Error(), string(bodyBytes))
	}
	obj, err := utils.AsMap(data)
	if err != nil {
		return diag.FromErr(errors.WithStack(err))
	}

	// coerce the type
	for _, item := range obj["items"].([]interface{}) {
		raw = append(raw, item.(map[string]interface{}))
	}

	if err := d.Set("kafkas", raw); err != nil {
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
