package tests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
	"github.com/stretchr/testify/assert"

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
						&serviceAccount),
					resource.TestCheckResourceAttr(
						serviceAccountPath, "name", randomName),
					resource.TestCheckResourceAttr(
						serviceAccountPath, "description", ""),
					resource.TestCheckResourceAttrSet(serviceAccountPath, "client_id"),
					resource.TestCheckResourceAttrSet(serviceAccountPath, "client_secret"),
					resource.TestCheckResourceAttrSet(serviceAccountPath, "created_by"),
					resource.TestCheckResourceAttrSet(serviceAccountPath, "created_at"),
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
						&preServiceAccount),
					resource.TestCheckResourceAttr(
						serviceAccountPath, "name", preName),
				),
			},
			{
				Config: testAccServiceAccountBasic(serviceAccountID, postName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountExists(
						&postServiceAccount),
					resource.TestCheckResourceAttr(
						serviceAccountPath, "name", postName),
					testCheckServiceAccountPreAndPostIDs(&preServiceAccount, &postServiceAccount),
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
					".",
				),
			},
		},
	})

}

// TestAccRHOASServiceAccount_DataSource checks the rhoas_service_account data source behavior
func TestAccRHOASServiceAccount_DataSource(t *testing.T) {
	var serviceAccount saclient.ServiceAccountData
	randomName := fmt.Sprintf("test-%s", randomString(10))
	config := fmt.Sprintf(`
resource "rhoas_service_account" "%[1]s" {
  name = "%[2]s"
}

data "rhoas_service_account" "test" {
	id = rhoas_service_account.%[1]s.id
}`, serviceAccountID, randomName)

	dataSourcePath := "data.rhoas_service_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountExists(&serviceAccount),
					resource.TestCheckResourceAttrPair(
						serviceAccountPath, "name",
						dataSourcePath, "name",
					),
					resource.TestCheckResourceAttrPair(
						serviceAccountPath, "client_id",
						dataSourcePath, "client_id",
					),
					resource.TestCheckResourceAttrPair(
						serviceAccountPath, "description",
						dataSourcePath, "description",
					),
					resource.TestCheckResourceAttrPair(
						serviceAccountPath, "id",
						dataSourcePath, "id",
					),
					// The secret is only retrieved in the creation, not in a data source
					resource.TestCheckResourceAttrSet(serviceAccountPath, "client_secret"),
					resource.TestCheckResourceAttrPair(
						serviceAccountPath, "created_by",
						dataSourcePath, "created_by",
					),
					resource.TestCheckResourceAttrPair(
						serviceAccountPath, "created_at",
						dataSourcePath, "created_at",
					),
				),
			},
		},
	})
}

// testAccCheckServiceAccountDestroy verifies the service account has been destroyed
func testAccCheckServiceAccountDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	factory, ok := testAccRHOAS.Meta().(rhoasAPI.Factory)
	if !ok {
		return errors.Errorf("unable to cast %v to rhoasAPI.Factory)", testAccRHOAS.Meta())
	}

	// loop through the resources in state, verifying each widget of type rhoas_kafka is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rhoas_service_account" {
			continue
		}

		// Retrieve the service account struct by referencing it's state ID for API lookup
		serviceAccount, resp, err := factory.ServiceAccountMgmt().GetServiceAccount(context.Background(), rs.Primary.ID).Execute()
		if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
			return apiErr
		}

		return errors.Errorf("expected a 404 but found a service account: %v", serviceAccount)
	}

	return nil
}

// testAccCheckServiceAccountDestroy verifies the service account at "rhoas_service_account.test_service_account" exists
func testAccCheckServiceAccountExists(serviceAccount *saclient.ServiceAccountData) resource.TestCheckFunc {

	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[serviceAccountPath]
		if !ok {
			return fmt.Errorf("Not found: %s", serviceAccountPath)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		api, ok := testAccRHOAS.Meta().(rhoasAPI.Factory)
		if !ok {
			return errors.Errorf("unable to cast %v to rhoasAPI.Factory)", testAccRHOAS.Meta())
		}
		gotServiceAccount, resp, err := api.ServiceAccountMgmt().GetServiceAccount(context.Background(), rs.Primary.ID).Execute()
		if apiErr := utils.GetAPIError(nil, resp, err); apiErr != nil {
			return apiErr
		}

		*serviceAccount = gotServiceAccount

		return nil
	}
}

// needed in order to pass linting until we unskip the test that uses this function
var _ = testCheckServiceAccountPreAndPostIDs

func testCheckServiceAccountPreAndPostIDs(pre, post *saclient.ServiceAccountData) resource.TestCheckFunc {
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
