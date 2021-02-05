package serviceaccounts

import (
	"context"
	"fmt"
	kasclient "github.com/bf2fc6cc711aee1a0c2a/cli/pkg/api/kas/client"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/connection"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"io/ioutil"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"
	"strconv"
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
							Default: "",
							ForceNew: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(20 * time.Minute),
		},
	}
}

func serviceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	_, _, apiErr := api.DeleteServiceAccount(ctx, d.Id()).Execute()
	if apiErr.Error() == "404 " {
		// the resource is deleted already
		d.SetId("")
		return diags
	}
	if apiErr.Error() != "" {
		return diag.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
	}

	d.SetId("")
	return diags
}

func serviceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	c := m.(*connection.KeycloakConnection)

	api := c.API().Kafka()

	var raw []map[string]interface{}

	data, resp, apiErr := api.ListServiceAccounts(ctx).Execute()
	if apiErr.Error() != "" {
		return diag.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
	}

	var serviceAccount *kasclient.ServiceAccountListItem

	body, _ := ioutil.ReadAll(resp.Body)

	diags = append(diags, diag.Diagnostic{
		Severity:      diag.Warning,
		Summary:       fmt.Sprintf("id: %s; Response: +%s", d.Id(), body),
		Detail:        "",
		AttributePath: nil,
	})

	// Find the service account as we don't have a get for it
	for _, sa := range *data.Items {
		if *sa.Id == d.Id() {
			serviceAccount = &sa
		}
	}

	if serviceAccount == nil {
		// the service account doesn't exist any more
		d.SetId("")
		return diags
	}

	obj, err := utils.AsMap(serviceAccount)
	if err != nil {
		return diag.FromErr(errors.WithStack(err))
	}
	raw = []map[string]interface{}{obj}

	items := fixClientIDAndClientSecret(raw)
	if err != nil {
		diag.FromErr(err);
	}
	err = applyRead(items, d)
	if err != nil {
		diag.FromErr(err);
	}

	return diags
}

func applyRead(items []map[string]interface{}, d *schema.ResourceData) error {
	if err := d.Set("service_account", items); err != nil {
		return err
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
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
			Name:          name,
		})
	}

	srr,resp, apiErr := api.CreateServiceAccount(ctx).ServiceAccountRequest(payload[0]).Execute()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary: resp.Status,
		Detail: string(body),
	})
	if apiErr.Error() != "" {
		return diag.Errorf("%s%s", apiErr.Error(), string(apiErr.Body()))
	}

	if srr.Id == nil {
		return diag.Errorf("no id provided")
	}

	d.SetId(*srr.Id)

	diags = append(diags, serviceAccountRead(ctx, d, m)...)
	return diags
}
