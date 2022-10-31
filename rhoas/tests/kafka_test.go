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

	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
)

const (
	kafkaID           = "test_kafka"
	kafkaPath         = "rhoas_kafka.test_kafka"
	planInput         = "developer.x1"
	billingModelInput = "standard"
)

// TestAccRHOASKafka_Basic checks that this provider is able to spin up a
// Kafka cluster and then destroy it.
func TestAccRHOASKafka_Basic(t *testing.T) {
	var kafka kafkamgmtclient.KafkaRequest
	randomName := fmt.Sprintf("test-%s", randomString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKafkaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaBasic(kafkaID, randomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKafkaExists(kafkaPath, &kafka),
					resource.TestCheckResourceAttr(
						kafkaPath, "name", randomName),
				),
			},
		},
	})

}

// TestAccRHOASKafka_Update checks that this provider is able to spin up a
// Kafka cluster and then update it. Finally, it destroys the resource.
func TestAccRHOASKafka_Update(t *testing.T) {
	// TODO: FIXME
	t.Skip("FIXME")

	var (
		// Used to compare the pre and post IDs
		prekafka  kafkamgmtclient.KafkaRequest
		postkafka kafkamgmtclient.KafkaRequest
	)

	randomName := fmt.Sprintf("test-%s", randomString(10))
	preName := fmt.Sprintf("%s-pre", randomName)
	postName := fmt.Sprintf("%s-post", randomName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKafkaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaBasic(kafkaID, preName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKafkaExists(kafkaPath, &prekafka),
					resource.TestCheckResourceAttr(
						kafkaPath, "name", preName),
				),
			},
			{
				Config: testAccKafkaBasic(kafkaID, postName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKafkaExists(kafkaPath, &postkafka),
					resource.TestCheckResourceAttr(
						kafkaPath, "name", postName),
					testCheckKafkaPreAndPostIDs(&prekafka, &postkafka),
				),
			},
		},
	})
}

// TestAccRHOASKafka_Error checks that this provider returns an error if
// some field is missconfigured
func TestAccRHOASKafka_Error(t *testing.T) {
	randomName := fmt.Sprintf("test-%s", randomString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKafkaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaWithCloudProvider(
					kafkaID, randomName, "not-a-cloud-provider"),
				ExpectError: regexp.MustCompile(
					".",
				),
			},
		},
	})

}

// testAccCheckKafkaDestroy verifies the Kafka cluster has been destroyed
func testAccCheckKafkaDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	factory, ok := testAccRHOAS.Meta().(rhoasAPI.Factory)
	if !ok {
		return errors.Errorf("unable to cast %v to rhoasAPI.Factory)", testAccRHOAS.Meta())
	}

	// loop through the resources in state, verifying each widget of type rhoas_kafka is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rhoas_kafka" {
			continue
		}

		// Retrieve the kafka struct by referencing it's state ID for API lookup
		kafka, resp, err := factory.KafkaMgmt().GetKafkaById(context.Background(), rs.Primary.ID).Execute()
		if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
			return apiErr
		}

		return errors.Errorf("expected a 404 but found a kafka instance: %v", kafka)
	}

	return nil
}

func testAccCheckKafkaExists(resource string, kafka *kafkamgmtclient.KafkaRequest) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		factory, ok := testAccRHOAS.Meta().(rhoasAPI.Factory)
		if !ok {
			return errors.Errorf("unable to cast %v to rhoasAPI.Factory)", testAccRHOAS.Meta())
		}
		gotKafka, resp, err := factory.KafkaMgmt().GetKafkaById(context.Background(), rs.Primary.ID).Execute()
		if err != nil {
			if apiErr := utils.GetAPIError(factory, resp, err); apiErr != nil {
				return apiErr
			}
		}

		if *gotKafka.Status != "ready" {
			return errors.Errorf("error provisioning kafka. Status is %s", *gotKafka.Status)
		}

		*kafka = gotKafka

		return nil
	}
}

// needed in order to pass linting until we unskip the test that uses this function
var _ = testCheckKafkaPreAndPostIDs

func testCheckKafkaPreAndPostIDs(pre, post *kafkamgmtclient.KafkaRequest) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if pre.Id != post.Id {
			return errors.Errorf("expected the id to be the same - before update the id was %s, after it was %s)", pre.Id, post.Id)
		}
		return nil
	}
}

func testAccKafkaBasic(id, name string) string {
	return fmt.Sprintf(`
resource "rhoas_kafka" "%s" {
  name = "%s"
  plan = "%s"
  billing_model = "%s"
}
`, id, name, planInput, billingModelInput)
}

func testAccKafkaWithCloudProvider(id, name, cloudProvider string) string {
	return fmt.Sprintf(`
resource "rhoas_kafka" "%s" {
  name = "%s"
  cloud_provider = "%s"
  plan = "%s"
  billing_model = "%s"
}
`, id, name, cloudProvider, planInput, billingModelInput)
}

func Test_testAccKafkaBasic(t *testing.T) {
	assert.Equal(
		t, fmt.Sprintf(`
resource "rhoas_kafka" "test_id" {
  name = "test_name"
  plan = "%s"
  billing_model = "%s"
}
`, planInput, billingModelInput),
		testAccKafkaBasic("test_id", "test_name"),
	)
}
