package acls

import (
	"context"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
	"math/rand"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
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

func ResourceACL() *schema.Resource {
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
				Description: "ID of the User or Service Account to bind created ACLs to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			KafkaIDField: {
				Description: "The ID of the kafka instance",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			ResourceTypeField: {
				Description: "Resource type of ACL, full list of possible values can be found here: https://github.com/redhat-developer/app-services-sdk-python/blob/main/sdks/kafka_instance_sdk/docs/AclResourceType.md",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			ResourceNameField: {
				Description: "Resource name of topic for the ACL",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			PatternTypeField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Pattern type of ACL, full list of possible values can be found here: https://github.com/redhat-developer/app-services-sdk-python/blob/main/sdks/kafka_instance_sdk/docs/AclPatternType.md",
				ForceNew:    true,
			},
			OperationTypeField: {
				Description: "Operation type of ACL, full list of possible values can be found here: https://github.com/redhat-developer/app-services-sdk-python/blob/main/sdks/kafka_instance_sdk/docs/AclOperation.md",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			PermissionTypeField: {
				Description: "Permission type of ACL, full list of possible values can be found here: https://github.com/redhat-developer/app-services-sdk-python/blob/main/sdks/kafka_instance_sdk/docs/AclPermissionType.md",
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

	api, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	kafkaID, ok := d.Get(KafkaIDField).(string)
	if !ok {
		return diag.FromErr(errors.Errorf("There was a problem getting the kafka ID value in the schema resource"))
	}

	instanceAPI, _, err := api.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	binding, err := mapResourceDataToACLBinding(d)
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

func mapResourceDataToACLBinding(d *schema.ResourceData) (*kafkainstanceclient.AclBinding, error) {

	// we only set these values from the resource data as all the rest are set as
	// computed in the schema and for us the computed values are assigned when we
	// get the kafka request object back from the API
	principal, ok := d.Get(PrincipalField).(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the principal value in the schema resource")
	}

	principal = PrincipalPrefix + principal

	resourceType, ok := d.Get(ResourceTypeField).(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the resource type value in the schema resource")
	}

	resourceName, ok := d.Get(ResourceNameField).(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the resource name value in the schema resource")
	}

	patternType, ok := d.Get(PatternTypeField).(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the pattern type value in the schema resource")
	}

	operationType, ok := d.Get(OperationTypeField).(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the operation type value in the schema resource")
	}

	permissionType, ok := d.Get(PermissionTypeField).(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the permission type value in the schema resource")
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
