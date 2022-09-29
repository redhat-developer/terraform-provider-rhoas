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

	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
)

const (
	topicID         = "test_topic"
	topicPath       = "rhoas_topic.test_topic"
	topicPartitions = 3
)

var kafkaName = fmt.Sprintf("test-create-topic-%s", randomString(6))

// TestAccRHOASTopic_Basic checks that this provider is able to create a topic
// and then destroy it.
func TestAccRHOASTopic_Basic(t *testing.T) {
	var topic kafkainstanceclient.Topic
	randomName := fmt.Sprintf("test-%s", randomString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// Check that the Kafka instance is also destroyed
		CheckDestroy: testAccCheckKafkaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicBasic(topicID, randomName, topicPartitions),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(topicPath, &topic),
					resource.TestCheckResourceAttr(
						topicPath, "name", randomName),
				),
			},
			{
				// Apply a configuration where just kafka is available and the topic is destroyed
				Config: testAccTopicDestroyedBasic(),
				Check:  resource.ComposeTestCheckFunc(testAccCheckTopicDestroy),
			},
		},
	})

}

// TestAccRHOASTopic_Update checks that this provider is able to create a
// topic and then update it. Finally, it destroys the resource.
func TestAccRHOASTopic_Update(t *testing.T) {
	// TODO: FIXME
	t.Skip("FIXME")

	var (
		// Used to compare the pre and post IDs
		pretopic  kafkainstanceclient.Topic
		posttopic kafkainstanceclient.Topic
	)

	randomName := fmt.Sprintf("test-%s", randomString(10))
	preName := fmt.Sprintf("%s-pre", randomName)
	postName := fmt.Sprintf("%s-post", randomName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// Check that the Kafka instance is destroyed
		CheckDestroy: testAccCheckKafkaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicBasic(topicID, preName, topicPartitions),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(topicPath, &pretopic),
					resource.TestCheckResourceAttr(
						topicPath, "name", preName),
				),
			},
			{
				Config: testAccTopicBasic(topicID, postName, topicPartitions),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(topicPath, &posttopic),
					resource.TestCheckResourceAttr(
						topicPath, "name", postName),
					testCheckTopicPreAndPostIDs(&pretopic, &posttopic),
				),
			},
		},
	})
}

// TestAccRHOASTopic_Error checks that this provider returns an error if
// some field is missconfigured
func TestAccRHOASTopic_Error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKafkaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicBasic(
					topicID, "", topicPartitions),
				ExpectError: regexp.MustCompile(
					"name must not be blank",
				),
			},
		},
	})

}

// testAccCheckTopicDestroy verifies the topic has been destroyed
func testAccCheckTopicDestroy(s *terraform.State) error {
	ctx := context.Background()
	// retrieve the connection established in Provider configuration
	api, ok := testAccRHOAS.Meta().(rhoasAPI.Clients)
	if !ok {
		return errors.Errorf("unable to cast %v to rhoasAPI.Clients)", testAccRHOAS.Meta())
	}

	// loop through the resources in state, verifying each widget of type rhoas_kafka is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rhoas_topic" {
			continue
		}

		kafkaID, ok := rs.Primary.Attributes["kafka_id"]
		if !ok {
			return errors.Errorf("kafka_id is not set for topic %s", rs.Primary.String())
		}
		name, ok := rs.Primary.Attributes["name"]
		if !ok {
			return errors.Errorf("name is not set for topic %s", rs.Primary.String())
		}

		instanceAPI, _, err := api.KafkaAdmin(&ctx, kafkaID)
		if err != nil {
			return err
		}

		topic, resp, err := instanceAPI.TopicsApi.GetTopic(ctx, name).Execute()
		if err != nil {
			if err.Error() == "404 Not Found" {
				return nil
			}
			if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
				return apiErr
			}
		}

		return errors.Errorf("expected a 404 but found a topic: %v", topic)
	}

	return nil
}

func testAccCheckTopicExists(resource string, topic *kafkainstanceclient.Topic) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

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

		kafkaID, ok := rs.Primary.Attributes["kafka_id"]
		if !ok {
			return errors.Errorf("kafka_id is not set for topic %s", rs.Primary.String())
		}
		name, ok := rs.Primary.Attributes["name"]
		if !ok {
			return errors.Errorf("name is not set for topic %s", rs.Primary.String())
		}

		instanceAPI, _, err := api.KafkaAdmin(&ctx, kafkaID)
		if err != nil {
			return err
		}

		gotTopic, resp, err := instanceAPI.TopicsApi.GetTopic(ctx, name).Execute()
		if err != nil {
			if apiErr := utils.GetAPIError(resp, err); apiErr != nil {
				return apiErr
			}
		}

		*topic = gotTopic

		return nil
	}
}

// needed in order to pass linting until we unskip the test that uses this function
var _ = testCheckTopicPreAndPostIDs

func testCheckTopicPreAndPostIDs(pre, post *kafkainstanceclient.Topic) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if pre.Id != post.Id {
			return errors.Errorf("expected the id to be the same - before update the id was %s, after it was %s)", *pre.Id, *post.Id)
		}
		return nil
	}
}

func testAccTopicBasic(id, name string, partitions int) string {
	return fmt.Sprintf(`
resource "rhoas_kafka" "test_kafka_create_topic" {
  name = "%s"
}

resource "rhoas_topic" "%s" {
  name = "%s"
  partitions = %d
  kafka_id   = rhoas_kafka.test_kafka_create_topic.id
}
`, kafkaName, id, name, partitions)
}

func testAccTopicDestroyedBasic() string {
	return fmt.Sprintf(`
resource "rhoas_kafka" "test_kafka_create_topic" {
  name = "%s"
}
`, kafkaName)
}

func Test_testAccTopicBasic(t *testing.T) {
	assert.Equal(
		t, fmt.Sprintf(`
resource "rhoas_kafka" "test_kafka_create_topic" {
  name = "%s"
}

resource "rhoas_topic" "test_id" {
  name = "test-name"
  partitions = 1
  kafka_id   = rhoas_kafka.test_kafka_create_topic.id
}
`, kafkaName),
		testAccTopicBasic("test_id", "test-name", 1),
	)
}
