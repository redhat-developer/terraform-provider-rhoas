package serviceaccounts

import (
	"context"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	serviceAccounts "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

const (
	DescriptionField = "description"
	NameField        = "name"
	ClientIDField    = "client_id"
	ClientSecret     = "client_secret"
	IDField          = "id"
)

func ServiceAccountSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		IDField: {
			Description: "The unique id fir the service account",
			Type:        schema.TypeString,
			Computed:    true,
			ForceNew:    true,
		},
		DescriptionField: {
			Description: "A description of the service account",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			ForceNew:    true,
		},
		NameField: {
			Description: "The name of the service account",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		ClientIDField: {
			Description: "The client id associated with the service account",
			Type:        schema.TypeString,
			Computed:    true,
		},
		ClientSecret: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The client secret associated with the service account. It must be stored by the client as the server will not return it after creation",
		},
	}
}

func ResourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "`rhoas_service_account` manages a service account in Red Hat OpenShift Streams for Apache Kafka.",
		CreateContext: serviceAccountCreate,
		ReadContext:   serviceAccountRead,
		DeleteContext: serviceAccountDelete,
		Schema:        ServiceAccountSchema(),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
	}
}

func serviceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	id, ok := d.Get(IDField).(string)
	if !ok {
		return diag.FromErr(errors.Errorf("There was a problem getting the id value in the schema resource"))
	}

	resp, err := factory.ServiceAccountMgmt().DeleteServiceAccount(ctx, id).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
			return diag.FromErr(apiErr)
		}
	}

	d.SetId("")
	return diags
}

func serviceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory)", m)
	}

	// the resource data ID field is the same as the service account id which is set when the
	// service account is created
	serviceAccount, resp, err := factory.ServiceAccountMgmt().GetServiceAccount(ctx, d.Id()).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
			return diag.FromErr(apiErr)
		}
	}

	err = setResourceDataFromServiceAccountData(d, &serviceAccount)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func serviceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory)", m)
	}

	tflog.Error(ctx, factory.Localizer().MustLocalize("service_account.error.no_name"))

	request, err := mapResourceDataToServiceAccountCreateRequest(d)
	if err != nil {
		return diag.FromErr(err)
	}

	serviceAccount, resp, err := factory.ServiceAccountMgmt().CreateServiceAccount(ctx).ServiceAccountCreateRequestData(*request).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
			return diag.FromErr(apiErr)
		}
	}

	d.SetId(serviceAccount.GetId())

	err = setResourceDataFromServiceAccountData(d, &serviceAccount)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func mapResourceDataToServiceAccountCreateRequest(d *schema.ResourceData) (*serviceAccounts.ServiceAccountCreateRequestData, error) {

	// we only set these values from the resource data as all the rest are set as
	// computed in the schema and for us the computed values are assigned when we
	// get the create request object back from the API
	description, ok := d.Get(DescriptionField).(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the description value in the schema resource")
	}

	name, ok := d.Get(NameField).(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the name value in the schema resource")
	}

	request := serviceAccounts.NewServiceAccountCreateRequestData(name)
	request.SetDescription(description)

	return request, nil
}

func setResourceDataFromServiceAccountData(d *schema.ResourceData, serviceAccount *serviceAccounts.ServiceAccountData) error {
	var err error

	if err = d.Set(IDField, serviceAccount.GetId()); err != nil {
		return err
	}

	if err = d.Set(ClientIDField, serviceAccount.GetClientId()); err != nil {
		return err
	}

	if err = d.Set(DescriptionField, serviceAccount.GetDescription()); err != nil {
		return err
	}

	if err = d.Set(NameField, serviceAccount.GetName()); err != nil {
		return err
	}

	return nil
}
