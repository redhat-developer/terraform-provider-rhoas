package tests

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"redhat.com/rhoas/rhoas-terraform-provider/m/rhoas"
)

var (
	testAccProviders map[string]*schema.Provider
	testAccRHOAS     *schema.Provider
)

func init() {
	testAccRHOAS = rhoas.Provider()
	testAccProviders = map[string]*schema.Provider{
		"rhoas": testAccRHOAS,
	}
}

func randomString(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))] // #nosec G404
	}
	return string(b)
}

// testAccPreCheck validates the necessary test API keys exist
// in the testing environment
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("OFFLINE_TOKEN"); v == "" {
		t.Fatal("OFFLINE_TOKEN must be set for acceptance tests")
	}
}
