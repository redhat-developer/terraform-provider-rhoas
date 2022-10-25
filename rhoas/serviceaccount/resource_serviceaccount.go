package serviceaccount

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	serviceAccounts "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
)

const (
	DescriptionField = "description"
	NameField        = "name"
	ClientIDField    = "client_id"
	ClientSecret     = "client_secret"
	IDField          = "id"
	CreatedByField   = "created_by"
	CreatedAtField   = "created_at"
)

func ResourceServiceAccount(localizer localize.Localizer) *schema.Resource {
	return &schema.Resource{
		Description:   "`rhoas_service_account` manages a service account in Red Hat OpenShift Streams for Apache Kafka.",
		CreateContext: serviceAccountCreate,
		ReadContext:   serviceAccountRead,
		DeleteContext: serviceAccountDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			IDField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.id"),
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			DescriptionField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.description"),
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
			},
			NameField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.name"),
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			ClientIDField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.clientID"),
				Type:        schema.TypeString,
				Computed:    true,
			},
			ClientSecret: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.clientSecret"),
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			CreatedByField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.createdBy"),
				Type:        schema.TypeString,
				Computed:    true,
			},
			CreatedAtField: {
				Description: localizer.MustLocalize("serviceaccount.resource.field.description.createdAt"),
				Type:        schema.TypeInt,
				Computed:    true,
			},
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
		return diag.FromErr(factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", IDField)))
	}

	resp, err := factory.ServiceAccountMgmt().DeleteServiceAccount(ctx, id).Execute()
	if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
		return diag.FromErr(apiErr)
	}

	d.SetId("")
	return diags
}

func serviceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	factory, ok := m.(rhoasAPI.Factory)
	if !ok {
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	// the resource data ID field is the same as the service account id which is set when the
	// service account is created
	serviceAccount, resp, err := factory.ServiceAccountMgmt().GetServiceAccount(ctx, d.Id()).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
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
		return diag.Errorf("unable to cast %v to rhoasAPI.Factory", m)
	}

	request, err := mapResourceDataToServiceAccountCreateRequest(factory, d)
	if err != nil {
		return diag.FromErr(err)
	}

	serviceAccount, resp, err := factory.ServiceAccountMgmt().CreateServiceAccount(ctx).ServiceAccountCreateRequestData(*request).Execute()
	if err != nil {
		if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
			return diag.FromErr(apiErr)
		}
	}

	d.SetId(serviceAccount.GetId())

	err = setResourceDataFromServiceAccountData(d, &serviceAccount)
	if err != nil {
		return diag.FromErr(err)
	}

	// This is only valid when creating, so running it out of setResourceDataFromServiceAccountData
	if err = d.Set(ClientSecret, serviceAccount.GetSecret()); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func mapResourceDataToServiceAccountCreateRequest(factory rhoasAPI.Factory, d *schema.ResourceData) (*serviceAccounts.ServiceAccountCreateRequestData, error) {

	// we only set these values from the resource data as all the rest are set as
	// computed in the schema and for us the computed values are assigned when we
	// get the create request object back from the API
	description, ok := d.Get(DescriptionField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", DescriptionField))

	}

	name, ok := d.Get(NameField).(string)
	if !ok {
		return nil, factory.Localizer().MustLocalizeError("common.errors.fieldNotFoundInSchema", localize.NewEntry("Field", NameField))

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

	if err = d.Set(CreatedByField, serviceAccount.GetCreatedBy()); err != nil {
		return err
	}

	if err = d.Set(CreatedAtField, serviceAccount.GetCreatedAt()); err != nil {
		return err
	}

	return nil
}
