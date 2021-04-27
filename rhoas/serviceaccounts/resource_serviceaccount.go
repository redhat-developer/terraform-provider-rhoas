package serviceaccounts

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	kasclient "github.com/redhat-developer/app-services-cli/pkg/api/kas/client"
	"io/ioutil"
	"log"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/cli/connection"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"time"
)

func ResourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: serviceAccountCreate,
		ReadContext:   serviceAccountRead,
		DeleteContext: serviceAccountDelete,
		Schema: map[string]*schema.Schema{
			"service_account": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
							ForceNew: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"client_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_secret": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
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

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	_, resp, err := api.DeleteServiceAccount(ctx, d.Id()).Execute()
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

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	var raw []map[string]interface{}

	// Store the existing client id
	existingServiceAccounts := d.Get("service_account")
	var existingClientSecret *string
	if existingServiceAccounts != nil {
		d := existingServiceAccounts.([]interface{})
		if len(d) == 1 {
			e := d[0].(map[string]interface{})["client_secret"]
			if e != nil {
				f := e.(string)
				existingClientSecret = &f
			}
		}
	}

	serviceAccount, resp, err := api.GetServiceAccountById(ctx, d.Id()).Execute()

	if err != nil && err.Error() == "404 Not Found" {
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

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	items := d.Get("service_account").([]interface{})

	payload := make([]kasclient.ServiceAccountRequest, 0)

	for _, item := range items {
		kafka := item.(map[string]interface{})

		description := kafka["description"].(string)
		name := kafka["name"].(string)

		payload = append(payload, kasclient.ServiceAccountRequest{
			Description: &description,
			Name:        name,
		})
	}

	srr, resp, err := api.CreateServiceAccount(ctx).ServiceAccountRequest(payload[0]).Execute()

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

	if srr.Id == nil {
		return diag.Errorf("no id provided")
	}

	d.SetId(*srr.Id)

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
