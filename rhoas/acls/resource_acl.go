package acls

import (
	"context"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	rhoasAPI "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/api"
)

const PrincipalPrefix = "User:"

func ResourceAcl() *schema.Resource {
	return &schema.Resource{
		Description:   "`rhoas_acl` manages an ACL binding for a Kafka instance in Red Hat OpenShift Streams for Apache Kafka.",
		CreateContext: aclCreate,
		ReadContext:   aclRead,
		DeleteContext: aclDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"principal": {
				Description: "ID of the User or Service Account to bind created ACLs to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"kafka_id": {
				Description: "The ID of the kafka instance",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"resource_type": {
				Description: "Resource type of ACL, full list of possible values can be found here: https://github.com/redhat-developer/app-services-sdk-python/blob/main/sdks/kafka_instance_sdk/docs/AclResourceType.md",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"resource_name": {
				Description: "Resource name of topic for the ACL",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"pattern_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Pattern type of ACL, full list of possible values can be found here: https://github.com/redhat-developer/app-services-sdk-python/blob/main/sdks/kafka_instance_sdk/docs/AclPatternType.md",
				ForceNew:    true,
			},
			"operation_type": {
				Description: "Operation type of ACL, full list of possible values can be found here: https://github.com/redhat-developer/app-services-sdk-python/blob/main/sdks/kafka_instance_sdk/docs/AclOperation.md",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"permission_type": {
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

	api, ok := m.(rhoasAPI.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Clients", m)
	}

	kafkaID, ok := d.Get("kafka_id").(string)
	if !ok {
		return diag.FromErr(errors.Errorf("There was a problem getting the kafka ID value in the schema resource"))
	}

	instanceApi, _, err := api.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setResourceDataFromAclData(ctx, d, instanceApi)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func aclCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	api, ok := m.(rhoasAPI.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Clients", m)
	}

	kafkaID, ok := d.Get("kafka_id").(string)
	if !ok {
		return diag.FromErr(errors.Errorf("There was a problem getting the kafka ID value in the schema resource"))
	}

	instanceAPI, _, err := api.KafkaAdmin(&ctx, kafkaID)
	if err != nil {
		return diag.FromErr(err)
	}

	binding, err := mapResourceDataToAclBinding(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = instanceAPI.AclsApi.CreateAcl(ctx).AclBinding(*binding).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	idNumber := rand.Intn(1_000_000_000)
	d.SetId(kafkaID + strconv.Itoa(idNumber))

	return diags
}

func setResourceDataFromAclData(ctx context.Context, d *schema.ResourceData, instanceApi *kafkainstanceclient.APIClient) error {
	var err error

	binding, err := mapResourceDataToAclBinding(d)
	if err != nil {
		return err
	}

	aclList, _, err := instanceApi.AclsApi.
		GetAcls(ctx).
		Principal(binding.GetPrincipal()).
		ResourceType(kafkainstanceclient.AclResourceTypeFilter(binding.GetResourceType())).
		ResourceName(binding.GetResourceName()).
		PatternType(kafkainstanceclient.AclPatternTypeFilter(binding.GetPatternType())).
		Operation(kafkainstanceclient.AclOperationFilter(binding.GetOperation())).
		Permission(kafkainstanceclient.AclPermissionTypeFilter(binding.GetPermission())).
		Execute()

	if err != nil {
		return err
	}

	if len(aclList.GetItems()) < 1 {
		return errors.Errorf("No ACLs matched the ACL trying to be read")
	}

	acl := aclList.GetItems()[0]

	// the acl binding principal needs to be prefixed with User:
	// we support adding this is in the provider so we need to remove it
	// when setting back our
	principal := acl.GetPrincipal()
	if strings.HasPrefix(principal, PrincipalPrefix) {
		principal = strings.TrimPrefix(principal, PrincipalPrefix)
	}

	if err = d.Set("principal", principal); err != nil {
		return err
	}

	if err = d.Set("resource_type", acl.GetResourceType()); err != nil {
		return err
	}

	if err = d.Set("resource_name", acl.GetResourceName()); err != nil {
		return err
	}

	if err = d.Set("pattern_type", acl.GetPatternType()); err != nil {
		return err
	}

	if err = d.Set("operation_type", acl.GetOperation()); err != nil {
		return err
	}

	if err = d.Set("permission_type", acl.GetPermission()); err != nil {
		return err
	}

	return nil
}

func mapResourceDataToAclBinding(d *schema.ResourceData) (*kafkainstanceclient.AclBinding, error) {

	// we only set these values from the resource data as all the rest are set as
	// computed in the schema and for us the computed values are assigned when we
	// get the kafka request object back from the API
	principal, ok := d.Get("principal").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the principal value in the schema resource")
	}

	principal = PrincipalPrefix + principal

	resourceType, ok := d.Get("resource_type").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the resource type value in the schema resource")
	}

	resourceName, ok := d.Get("resource_name").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the resource name value in the schema resource")
	}

	patternType, ok := d.Get("pattern_type").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the pattern type value in the schema resource")
	}

	operationType, ok := d.Get("operation_type").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the operation type value in the schema resource")
	}

	permissionType, ok := d.Get("permission_type").(string)
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
