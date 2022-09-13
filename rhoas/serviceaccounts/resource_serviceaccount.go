package serviceaccounts

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	serviceAccounts "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
	rhoasClients "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/clients"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"time"
)

func ResourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "`rhoas_service_account` manages a service account in Red Hat OpenShift Streams for Apache Kafka.",
		CreateContext: serviceAccountCreate,
		ReadContext:   serviceAccountRead,
		DeleteContext: serviceAccountDelete,
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "A description of the service account",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the service account",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"client_id": {
				Description: "The client id associated with the service account",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"owner": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The username of the Red Hat account that owns the service account",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The client secret associated with the service account. It must be stored by the client as the server will not return it after creation",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
	}
}

func serviceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, ok := m.(*rhoasClients.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	resp, err := c.ServiceAccountClient.ServiceAccountsApi.DeleteServiceAccount(ctx, d.Id()).Execute()
	if err != nil {
		apiError, err1 := utils.GetAPIError(resp, err)
		if err1 != nil {
			return diag.FromErr(err1)
		}

		return diag.FromErr(apiError)
	}

	d.SetId("")
	return diags
}

func serviceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	c, ok := m.(*rhoasClients.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	// the resource data ID field is the same as the service account id which is set when the
	// service account is created
	serviceAccount, resp, err := c.ServiceAccountClient.ServiceAccountsApi.GetServiceAccount(ctx, d.Id()).Execute()
	if err != nil {
		apiError, err1 := utils.GetAPIError(resp, err)
		if err1 != nil {
			return diag.FromErr(err1)
		}

		return diag.FromErr(apiError)
	}

	serviceAccountData, err := utils.AsMap(serviceAccount)
	if err != nil {
		return diag.FromErr(errors.WithStack(err))
	}

	err = setResourceDataFromServiceAccountData(d, &serviceAccountData)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func serviceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, ok := m.(*rhoasClients.Clients)
	if !ok {
		return diag.Errorf("unable to cast %v to *rhoasClients.Clients", m)
	}

	request, err := mapResourceDataToServiceAccountCreateRequest(d)
	if err != nil {
		return diag.FromErr(err)
	}

	srr, resp, err := c.ServiceAccountClient.ServiceAccountsApi.CreateServiceAccount(ctx).ServiceAccountCreateRequestData(*request).Execute()
	if err != nil {
		apiError, err1 := utils.GetAPIError(resp, err)
		if err1 != nil {
			return diag.FromErr(err1)
		}

		return diag.FromErr(apiError)
	}

	d.SetId(srr.GetId())

	serviceAccountData, err := utils.AsMap(srr)
	if err != nil {
		return diag.FromErr(errors.WithStack(err))
	}

	err = setResourceDataFromServiceAccountData(d, &serviceAccountData)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func mapResourceDataToServiceAccountCreateRequest(d *schema.ResourceData) (*serviceAccounts.ServiceAccountCreateRequestData, error) {

	// we only set these values from the resource data as all the rest are set as
	// computed in the schema and for us the computed values are assigned when we
	// get the create request object back from the API
	description, ok := d.Get("description").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the description value in the schema resource")
	}

	name, ok := d.Get("name").(string)
	if !ok {
		return nil, errors.Errorf("There was a problem getting the name value in the schema resource")
	}

	request := serviceAccounts.NewServiceAccountCreateRequestData(name)
	request.SetDescription(description)

	return request, nil
}

func setResourceDataFromServiceAccountData(d *schema.ResourceData, serviceAccountData *map[string]interface{}) error {
	var err error

	if err = d.Set("client_id", (*serviceAccountData)["client_id"]); err != nil {
		return err
	}

	if err = d.Set("description", (*serviceAccountData)["description"]); err != nil {
		return err
	}

	if err = d.Set("name", (*serviceAccountData)["name"]); err != nil {
		return err
	}

	if err = d.Set("owner", (*serviceAccountData)["owner"]); err != nil {
		return err
	}

	return nil
}
