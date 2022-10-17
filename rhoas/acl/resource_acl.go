package acl

import (
	"context"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"
	"math/rand"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
)

const (
	PrincipalPrefix = "User:"

	PrincipalField      = "principal"
	KafkaIDField        = "kafka_id"
	ResourceTypeField   = "resource_type"
	ResourceNameField   = "resource_name"
	PatternTypeField    = "pattern_type"
	OperationTypeField  = "operation_type"
	PermissionTypeField = "permission_type"
)

func ResourceACL(localizer localize.Localizer) *schema.Resource {
	return &schema.Resource{
		Description:   "`rhoas_acl` manages an ACL binding for a Kafka instance in Red Hat OpenShift Streams for Apache Kafka.",
		CreateContext: aclCreate,
		ReadContext:   aclRead,
		DeleteContext: aclDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			PrincipalField: {
				Description: localizer.MustLocalize("acl.resource.field.description.principal"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			KafkaIDField: {
				Description: localizer.MustLocalize("acl.resource.field.description.kafkaID"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			ResourceTypeField: {
				Description: localizer.MustLocalize("acl.resource.field.description.resourceType"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			ResourceNameField: {
				Description: localizer.MustLocalize("acl.resource.field.description.resourceName"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			PatternTypeField: {
				Description: localizer.MustLocalize("acl.resource.field.description.patternType"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			OperationTypeField: {
				Description: localizer.MustLocalize("acl.resource.field.description.operationType"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			PermissionTypeField: {
				Description: localizer.MustLocalize("acl.resource.field.description.permissionType"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func aclDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func aclRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func aclCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	binding, err := mapResourceDataToACLBinding(factory, d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = instanceAPI.AclsApi.CreateAcl(ctx).AclBinding(*binding).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	// acls have no id so we need to create a new random one for it
	idNumber := rand.Intn(1_000_000_000) //nolint:gosec
	d.SetId(kafkaID + strconv.Itoa(idNumber))

	return diags
}

func mapResourceDataToACLBinding(factory rhoasAPI.Factory, d *schema.ResourceData) (*kafkainstanceclient.AclBinding, error) {

	// we only set these values from the resource data as all the rest are set as
	// computed in the schema and for us the computed values are assigned when we
	// get the kafka request object back from the API
	principal, ok := d.Get(PrincipalField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", PrincipalField))
	}

	principal = PrincipalPrefix + principal

	resourceType, ok := d.Get(ResourceTypeField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", ResourceTypeField))
	}

	resourceName, ok := d.Get(ResourceNameField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", ResourceNameField))
	}

	patternType, ok := d.Get(PatternTypeField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", PatternTypeField))
	}

	operationType, ok := d.Get(OperationTypeField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", OperationTypeField))
	}

	permissionType, ok := d.Get(PermissionTypeField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", PermissionTypeField))
	}

	binding := kafkainstanceclient.NewAclBinding(
		kafkainstanceclient.AclResourceType(resourceType),
		resourceName,
		kafkainstanceclient.AclPatternType(patternType),
		principal,
		kafkainstanceclient.AclOperation(operationType),
		kafkainstanceclient.AclPermissionType(permissionType),
	)

	return binding, nil
}
