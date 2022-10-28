package kafka

import (
	"context"
	"strings"
	"time"

	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/acl"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

const (
	NameField                    = "name"
	CloudProviderField           = "cloud_provider"
	RegionField                  = "region"
	ReauthenticationEnabledField = "reauthentication_enabled"
	PlanField                    = "plan"
	BillingCloudAccountIDField   = "billing_cloud_account_id"
	MarketPlaceField             = "marketplace"
	BillingModelField            = "billing_model"
	HrefField                    = "href"
	StatusField                  = "status"
	OwnerField                   = "owner"
	BootstrapServerHostField     = "bootstrap_server_host"
	CreatedAtField               = "created_at"
	UpdatedAtField               = "updated_at"
	IDField                      = "id"
	KindField                    = "kind"
	VersionField                 = "version"
	ACLField                     = "acl"

	DefaultEmptyField = ""
)

// nolint:funlen
func ResourceKafka(localizer localize.Localizer) *schema.Resource {
	return &schema.Resource{
		Description:   "`rhoas_kafka` manages a Kafka instance in Red Hat OpenShift Streams for Apache Kafka.",
		CreateContext: kafkaCreate,
		ReadContext:   kafkaRead,
		DeleteContext: kafkaDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			NameField: {
				Description: localizer.MustLocalize("kafka.resource.field.description.name"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			CloudProviderField: {
				Description: localizer.MustLocalize("kafka.resource.field.description.cloudProvider"),
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "aws",
				ForceNew:    true,
			},
			RegionField: {
				Description: localizer.MustLocalize("kafka.resource.field.description.region"),
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "us-east-1",
				ForceNew:    true,
			},
			ReauthenticationEnabledField: {
				Description: localizer.MustLocalize("kafka.resource.field.description.reauthenticationEnabled"),
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			PlanField: {
				Description: localizer.MustLocalize("kafka.resource.field.description.plan"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			BillingCloudAccountIDField: {
				Description: localizer.MustLocalize("kafka.resource.field.description.billingCloudAccountId"),
				Type:        schema.TypeString,
				Optional:    true,
				Default:     DefaultEmptyField,
				ForceNew:    true,
			},
			MarketPlaceField: {
				Description: localizer.MustLocalize("kafka.resource.field.description.marketplace"),
				Type:        schema.TypeString,
				Optional:    true,
				Default:     DefaultEmptyField,
				ForceNew:    true,
			},
			BillingModelField: {
				Description: localizer.MustLocalize("kafka.resource.field.description.billingModel"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
				Description: localizer.MustLocalize("kafka.resource.field.description.id"),
				Type:        schema.TypeString,
				Computed:    true,
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
			ACLField: {
				Description: localizer.MustLocalize("kafka.resource.field.description.acl"),
				Type:        schema.TypeList,
				ForceNew:    true,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: schema.TypeString,
				},
			},
		},
	}
}

func kafkaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	_, resp, err := factory.KafkaMgmt().DeleteKafkaById(ctx, d.Id()).Async(true).Execute()
	if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
		return diag.FromErr(apiErr)
	}

	deleteStateConf := &resource.StateChangeConf{
		Delay: 5 * time.Second,
		Pending: []string{
			"deprovision", "deleting",
		},
		Refresh: func() (interface{}, string, error) {
			data, resp, err1 := factory.KafkaMgmt().GetKafkaById(ctx, d.Id()).Execute()
			if utils.CheckNotFound(resp) {
				return data, "deleted", nil
			}
			if apiErr := utils.GetAPIError(factory, resp, err1); apiErr != nil {
				return nil, "", apiErr
			}

			return data, *data.Status, nil
		},
		Target: []string{
			"deleted", "",
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

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	kafka, resp, err := factory.KafkaMgmt().GetKafkaById(ctx, d.Id()).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
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

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	requestPayload, err := mapResourceDataToKafkaPayload(factory, d)
	if err != nil {
		return diag.FromErr(err)
	}

	kr, resp, err := factory.KafkaMgmt().CreateKafka(ctx).Async(true).KafkaRequestPayload(*requestPayload).Execute()
	if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
		return diag.FromErr(apiErr)
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
			kafka, resp, err1 := factory.KafkaMgmt().GetKafkaById(ctx, kr.Id).Execute()
			if apiErr := utils.GetAPIError(factory, resp, err1); apiErr != nil {
				return nil, "", apiErr
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
		return diag.FromErr(err)
	}

	kafka, castOk := data.(kafkamgmtclient.KafkaRequest)
	if !castOk {
		return diag.Errorf("Cannot cast data from kafka creation to to map[string]interface{}")
	}

	err = setResourceDataFromKafkaData(d, &kafka)
	if err != nil {
		return diag.FromErr(err)
	}

	// now that kafka is created define the acl
	err = createACLForKafka(ctx, factory, d, &kafka)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func createACLForKafka(ctx context.Context, factory rhoasAPI.Factory, d *schema.ResourceData, kafka *kafkamgmtclient.KafkaRequest) error {

	aclInput := d.Get(ACLField)
	if aclInput == nil {
		// no input was given for acl so do nothing
		return nil
	}

	aclConfig, ok := aclInput.([]interface{})
	if !ok {
		return factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", ACLField))

	}

	for i := 0; i < len(aclConfig); i++ {
		element, ok := aclConfig[i].(map[string]interface{})
		if !ok {
			return factory.Localizer().MustLocalizeError("kafka.errors.unableToRetrieveAclContents")
		}

		principal, ok := element[acl.PrincipalField].(string)
		if !ok {
			return factory.Localizer().MustLocalizeError("kafka.errors.noAclFieldGiven", localize.NewEntry("Field", acl.PrincipalField))
		}

		// required for api, the user id, service account id or * works
		// when appended to User:
		principal = acl.PrincipalPrefix + principal

		resourceType, ok := element[acl.ResourceTypeField].(string)
		if !ok {
			return factory.Localizer().MustLocalizeError("kafka.errors.noAclFieldGiven", localize.NewEntry("Field", acl.ResourceTypeField))
		}

		resourceName, ok := element[acl.ResourceNameField].(string)
		if !ok {
			return factory.Localizer().MustLocalizeError("kafka.errors.noAclFieldGiven", localize.NewEntry("Field", acl.ResourceNameField))
		}

		patternType, ok := element[acl.PatternTypeField].(string)
		if !ok {
			return factory.Localizer().MustLocalizeError("kafka.errors.noAclFieldGiven", localize.NewEntry("Field", acl.PatternTypeField))
		}

		operationType, ok := element[acl.OperationTypeField].(string)
		if !ok {
			return factory.Localizer().MustLocalizeError("kafka.errors.noAclFieldGiven", localize.NewEntry("Field", acl.OperationTypeField))
		}

		permissionType, ok := element[acl.PermissionTypeField].(string)
		if !ok {
			return factory.Localizer().MustLocalizeError("kafka.errors.noAclFieldGiven", localize.NewEntry("Field", acl.PermissionTypeField))
		}

		binding := kafkainstanceclient.NewAclBinding(
			kafkainstanceclient.AclResourceType(strings.ToUpper(resourceType)),
			resourceName,
			kafkainstanceclient.AclPatternType(strings.ToUpper(patternType)),
			principal,
			kafkainstanceclient.AclOperation(strings.ToUpper(operationType)),
			kafkainstanceclient.AclPermissionType(strings.ToUpper(permissionType)),
		)

		instanceAPI, _, err := factory.KafkaAdmin(&ctx, kafka.GetId())
		if err != nil {
			return err
		}

		resp, err := instanceAPI.AclsApi.CreateAcl(ctx).AclBinding(*binding).Execute()
		if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
			return apiErr
		}
	}

	return nil
}

func setResourceDataFromKafkaData(d *schema.ResourceData, kafka *kafkamgmtclient.KafkaRequest) error {
	var err error

	// any computed field is then set here
	if err = d.Set(BillingCloudAccountIDField, kafka.GetBillingCloudAccountId()); err != nil {
		return err
	}

	if err = d.Set(MarketPlaceField, kafka.GetMarketplace()); err != nil {
		return err
	}

	if err = d.Set(HrefField, kafka.GetHref()); err != nil {
		return err
	}

	if err = d.Set(StatusField, kafka.GetStatus()); err != nil {
		return err
	}

	if err = d.Set(OwnerField, kafka.GetOwner()); err != nil {
		return err
	}

	if err = d.Set(BootstrapServerHostField, kafka.GetBootstrapServerHost()); err != nil {
		return err
	}

	if err = d.Set(CreatedAtField, kafka.GetCreatedAt().Format(time.RFC3339)); err != nil {
		return err
	}

	if err = d.Set(UpdatedAtField, kafka.GetUpdatedAt().Format(time.RFC3339)); err != nil {
		return err
	}

	if err = d.Set(IDField, kafka.GetId()); err != nil {
		return err
	}

	if err = d.Set(KindField, kafka.GetKind()); err != nil {
		return err
	}

	if err = d.Set(VersionField, kafka.GetVersion()); err != nil {
		return err
	}

	return nil
}

func mapResourceDataToKafkaPayload(factory rhoasAPI.Factory, d *schema.ResourceData) (*kafkamgmtclient.KafkaRequestPayload, error) {

	// any required fields and optionals with defaults values are set from here
	name, ok := d.Get(NameField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", NameField))
	}

	cloudProvider, ok := d.Get(CloudProviderField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", CloudProviderField))
	}

	region, ok := d.Get(RegionField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", RegionField))
	}

	reauthenticationEnabled, ok := d.Get(ReauthenticationEnabledField).(bool)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", ReauthenticationEnabledField))
	}

	plan, ok := d.Get(PlanField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", PlanField))
	}

	billingCloudAccountID, ok := d.Get(BillingCloudAccountIDField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", BillingCloudAccountIDField))
	}

	marketplace, ok := d.Get(MarketPlaceField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", MarketPlaceField))
	}

	billingModel, ok := d.Get(BillingModelField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", BillingModelField))
	}

	payload := kafkamgmtclient.NewKafkaRequestPayload(name)

	payload.SetCloudProvider(cloudProvider)
	payload.SetRegion(region)
	payload.SetReauthenticationEnabled(reauthenticationEnabled)
	payload.SetPlan(plan)
	payload.SetBillingModel(billingModel)

	// these are optional fields so if they are not set then do not set them
	// in the api call
	if billingCloudAccountID != DefaultEmptyField {
		payload.SetBillingCloudAccountId(billingCloudAccountID)
	}

	if marketplace != DefaultEmptyField {
		payload.SetMarketplace(marketplace)
	}

	return payload, nil
}
