package tests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	rhoasAPI "redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/api"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas/utils"

	saclient "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
)

const (
	serviceAccountID   = "test_service_account"
	serviceAccountPath = "rhoas_service_account.test_service_account"
)

// TestAccRHOASServiceAccount_Basic checks that this provider is able to create a
// service account and then destroy it.
func TestAccRHOASServiceAccount_Basic(t *testing.T) {
	var serviceAccount saclient.ServiceAccountData
	randomName := fmt.Sprintf("test-%s", randomString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountBasic(serviceAccountID, randomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountExists(
						serviceAccountPath, &serviceAccount),
					resource.TestCheckResourceAttr(
						serviceAccountPath, "name", randomName),
					resource.TestCheckResourceAttr(
						serviceAccountPath, "description", ""),
					// TODO: Add more checks?
				),
			},
		},
	})

}

// TestAccRHOASServiceAccount_Update checks that this provider is able create a
// service account cluster and then update it. Finally, it destroys the resource.
func TestAccRHOASServiceAccount_Update(t *testing.T) {
	// TODO: FIXME
	t.Skip("FIXME")

	randomName := fmt.Sprintf("test-%s", randomString(10))
	preName := fmt.Sprintf("%s-pre", randomName)
	postName := fmt.Sprintf("%s-post", randomName)

	var (
		// Used to compare the pre and post IDs
		preServiceAccount  saclient.ServiceAccountData
		postServiceAccount saclient.ServiceAccountData
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountBasic(serviceAccountID, preName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountExists(
						serviceAccountPath, &preServiceAccount),
					resource.TestCheckResourceAttr(
						serviceAccountPath, "name", preName),
				),
			},
			{
				Config: testAccServiceAccountBasic(serviceAccountID, postName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountExists(
						serviceAccountPath, &postServiceAccount),
					resource.TestCheckResourceAttr(
						serviceAccountPath, "name", postName),
					testCheckPreAndPostIDs(&preServiceAccount, &postServiceAccount),
				),
			},
		},
	})
}

// TestAccRHOASServiceAccount_Error checks that this provider returns an error if
// some field is misconfigured
func TestAccRHOASServiceAccount_Error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountBasic(
					serviceAccountID, ""),
				ExpectError: regexp.MustCompile(
					"Request failed field validation",
				),
			},
		},
	})

}

// testAccCheckServiceAccountDestroy verifies the service account has been destroyed
func testAccCheckServiceAccountDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	api, ok := testAccRHOAS.Meta().(rhoasAPI.Clients)
	if !ok {
		return errors.Errorf("unable to cast %v to rhoasAPI.Clients)", testAccRHOAS.Meta())
	}

	// loop through the resources in state, verifying each widget of type rhoas_kafka is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rhoas_service_account" {
			continue
		}

		// Retrieve the service account struct by referencing it's state ID for API lookup
		serviceAccount, resp, err := api.ServiceAccountMgmt().GetServiceAccount(context.Background(), rs.Primary.ID).Execute()
		if err != nil {
			if err.Error() == "404 Not Found" {
				return nil
			}
			if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
				return apiErr
			}
		}

		return errors.Errorf("expected a 404 but found a service account: %v", serviceAccount)
	}

	return nil
}

// testAccCheckServiceAccountDestroy verifies the service account exists
func testAccCheckServiceAccountExists(resource string, serviceAccount *saclient.ServiceAccountData) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		api, ok := testAccRHOAS.Meta().(rhoasAPI.Clients)
		if !ok {
			return errors.Errorf("unable to cast %v to rhoasAPI.Clients)", testAccRHOAS.Meta())
		}
		gotServiceAccount, resp, err := api.ServiceAccountMgmt().GetServiceAccount(context.Background(), rs.Primary.ID).Execute()
		if err != nil {
			if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
				return apiErr
			}
		}

		*serviceAccount = gotServiceAccount

		return nil
	}
}

// needed in order to pass linting until we unskip the test that uses this function
var _ = testCheckPreAndPostIDs

func testCheckPreAndPostIDs(pre, post *saclient.ServiceAccountData) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if *pre.Id != *post.Id {
			return errors.Errorf("expected the id to be the same - before update the id was %s, after it was %s)", *pre.Id, *post.Id)
		}
		return nil
	}
}

func testAccServiceAccountBasic(id, name string) string {
	return fmt.Sprintf(`
resource "rhoas_service_account" "%s" {
  name = "%s"
}
`, id, name)
}

func Test_testAccServiceAccountBasic(t *testing.T) {
	assert.Equal(
		t, `
resource "rhoas_service_account" "test_id" {
  name = "test_name"
}
`,
		testAccServiceAccountBasic("test_id", "test_name"),
	)
}