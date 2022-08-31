package serviceaccounts

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	"io/ioutil"
	"log"
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
			"service_account": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
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
				},
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

	c, ok := m.(*kafkamgmtclient.APIClient)
	if !ok {
		return diag.Errorf("unable to cast %v to *connection.KeycloakConnection", m)
	}

	_, resp, err := c.SecurityApi.DeleteServiceAccountById(ctx, d.Id()).Execute()
	if err != nil && err.Error() == "404 " {
		// the resource is deleted already
		d.SetId("")
		return diags
	}
	if err != nil {
		bodyBytes, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			log.Fatal(ioErr)
		}
		return diag.Errorf("%s%s", err.Error(), string(bodyBytes))
	}

	d.SetId("")
	return diags
}

func serviceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	c, ok := m.(*kafkamgmtclient.APIClient)
	if !ok {
		return diag.Errorf("unable to cast %v to *connection.KeycloakConnection", m)
	}

	var raw []map[string]interface{}

	// Store the existing client id
	existingServiceAccounts := d.Get("service_account")
	var existingClientSecret *string
	if existingServiceAccounts != nil {
		d, ok := existingServiceAccounts.([]interface{})
		if !ok {
			return diag.Errorf("unable to cast %v to []interface{}", existingServiceAccounts)
		}
		if len(d) == 1 {
			e := d[0].(map[string]interface{})["client_secret"]
			if e != nil {
				f, ok := e.(string)
				if !ok {
					return diag.Errorf("unable to cast %v to string", e)
				}
				existingClientSecret = &f
			}
		}
	}

	serviceAccount, resp, err := c.SecurityApi.GetServiceAccountById(ctx, d.Id()).Execute()

	if err != nil && err.Error() == "404 Not Found" {
		d.SetId("")
		return diags
	}

	if err != nil {
		bodyBytes := []byte("empty response")
		if resp != nil {
			var ioErr error
			bodyBytes, ioErr = ioutil.ReadAll(resp.Body)
			if ioErr != nil {
				log.Fatal(ioErr)
			}
		}
		return diag.Errorf("%s %s", err.Error(), string(bodyBytes))
	}

	obj, err := utils.AsMap(serviceAccount)
	if err != nil {
		return diag.FromErr(errors.WithStack(err))
	}
	raw = []map[string]interface{}{obj}

	items := fixClientIDAndClientSecret(raw, existingClientSecret)
	if err != nil {
		return diag.FromErr(err)
	}
	err = applyRead(items, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func applyRead(items []map[string]interface{}, d *schema.ResourceData) error {
	filter := []string{"name", "description", "client_id", "owner", "client_secret"}
	filtered := utils.Filter(items, filter)
	if err := d.Set("service_account", filtered); err != nil {
		return err
	}
	return nil
}

func serviceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, ok := m.(*kafkamgmtclient.APIClient)
	if !ok {
		return diag.Errorf("unable to cast %v to *connection.KeycloakConnection", m)
	}

	val := d.Get("service_account")
	items, ok := val.([]interface{})
	if !ok {
		return diag.Errorf("unable to cast %v to []interface{}", val)
	}

	payload := make([]kafkamgmtclient.ServiceAccountRequest, 0)

	for _, item := range items {
		kafka, ok := item.(map[string]interface{})
		if !ok {
			return diag.Errorf("unable to cast %v to map[string]interface{}", item)
		}

		description, ok := kafka["description"].(string)
		if !ok {
			return diag.Errorf("unable to cast %v to string", kafka["description"])
		}
		name, ok := kafka["name"].(string)
		if !ok {
			return diag.Errorf("unable to cast %v to string", kafka["name"])
		}

		payload = append(payload, kafkamgmtclient.ServiceAccountRequest{
			Description: &description,
			Name:        name,
		})
	}

	srr, resp, err := c.SecurityApi.CreateServiceAccount(ctx).ServiceAccountRequest(payload[0]).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(err)
	}

	if err != nil {
		bodyBytes, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			log.Fatal(ioErr)
		}
		return diag.Errorf("%s%s", err.Error(), string(bodyBytes))
	}

	if srr.Id == "" {
		return diag.Errorf("no id provided")
	}

	d.SetId(srr.Id)

	obj, err := utils.AsMap(srr)
	if err != nil {
		return diag.FromErr(errors.WithStack(err))
	}
	raw := []map[string]interface{}{obj}

	fixed := fixClientIDAndClientSecret(raw, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	err = applyRead(fixed, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
