package acls

import (
	"context"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	rhoasAPI "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/api"
)

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
				Optional:    true,
				ForceNew:    true,
			},
			"kafka_admin_url": {
				Description: "URL of the kafka instance to connect to",
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

	return diags
}

func setResourceDataFromAclData(d *schema.ResourceData, kafka *kafkamgmtclient.KafkaRequest) error {
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

func mapResourceDataToAclBinding(d *schema.ResourceData) (*kafkainstanceclient.AclBinding, error) {

	// we only set these values from the resource data as all the rest are set as
	// computed in the schema and for us the computed values are assigned when we
	// get the kafka request object back from the API
	principal, ok := d.Get("principal").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the principal value in the schema resource")
	}

	resource_type, ok := d.Get("resource_type").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the resource type value in the schema resource")
	}

	resource_name, ok := d.Get("resource_name").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the resource name value in the schema resource")
	}

	pattern_type, ok := d.Get("pattern_type").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the pattern type value in the schema resource")
	}

	operation_type, ok := d.Get("operation_type").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the operation type value in the schema resource")
	}

	permission_type, ok := d.Get("permission_type").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the permission type value in the schema resource")
	}

	binding := kafkainstanceclient.NewAclBinding(
		kafkainstanceclient.AclResourceType(resource_type),
		resource_name, kafkainstanceclient.AclPatternType(pattern_type),
		principal, kafkainstanceclient.AclOperation(operation_type),
		kafkainstanceclient.AclPermissionType(permission_type),
	)

	return binding, nil
}
