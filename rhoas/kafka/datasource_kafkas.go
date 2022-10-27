package kafka

import (
	"context"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

func DataSourceKafkas(localizer localize.Localizer) *schema.Resource {
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
						CloudProviderField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.cloudProvider"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						RegionField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.region"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						NameField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.name"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						HrefField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.href"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						StatusField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.status"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						OwnerField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.owner"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						BootstrapServerHostField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.bootstrapServerHost"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						CreatedAtField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.createdAt"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						UpdatedAtField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.updatedAt"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						IDField: {
							Description: localizer.MustLocalize("kafka.datasource.field.description.id"),
							Type:        schema.TypeString,
							Required:    true,
						},
						KindField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.kind"),
							Type:        schema.TypeString,
							Computed:    true,
						},
						VersionField: {
							Description: localizer.MustLocalize("kafka.resource.field.description.version"),
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
	if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
		return diag.FromErr(apiErr)
	}

	if err := d.Set("kafkas", flattenKafkas(kafkas.Items)); err != nil {
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

func flattenKafkas(kafkas []kafkamgmtclient.KafkaRequest) []interface{} {
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
