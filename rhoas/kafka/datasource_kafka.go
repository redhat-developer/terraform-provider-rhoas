package kafka

import (
	"context"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

func DataSourceKafka(localizer localize.Localizer) *schema.Resource {
	return &schema.Resource{
		Description: "`rhoas_kafka` provides a Kafka accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.",
		ReadContext: dataSourceKafkaRead,
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
	}
}

func dataSourceKafkaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	id, ok := d.Get(IDField).(string)
	if !ok {
		return diag.FromErr(factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", IDField)))
	}

	kafka, resp, err := factory.KafkaMgmt().GetKafkaById(ctx, id).Execute()
	if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
		return diag.FromErr(apiErr)
	}

	err = setResourceDataFromKafkaData(d, &kafka)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
