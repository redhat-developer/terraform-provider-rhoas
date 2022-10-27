package utils_test

import (
	factories "github.com/redhat-developer/terraform-provider-rhoas/rhoas/factory"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize/goi18n"
	"testing"

	"github.com/pkg/errors"

	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/utils"
	"github.com/stretchr/testify/assert"
)

func TestAsMap(t *testing.T) {
	t.Run("non unmarshable input", func(t *testing.T) {
		_, err := utils.AsMap("invalid object")
		assert.Error(t, err, "expected an error while converting an invalid JSON struct into a map")
	})

	t.Run("invalid input", func(t *testing.T) {
		_, err := utils.AsMap(make(chan int))
		assert.Error(t, err, "expected an error while converting an invalid JSON struct into a map")
	})

	t.Run("valid input", func(t *testing.T) {
		type testStruct struct {
			Name string `json:"name"`
		}
		testJSON := testStruct{
			Name: "test",
		}
		want := map[string]interface{}{
			"name": "test",
		}
		got, err := utils.AsMap(testJSON)
		assert.NoError(t, err, "got unexpected error while converting a valid JSON struct into a map")
		assert.Equal(t, want, got, "unexpected value was returned")
	})

}

func TestGetAPIError(t *testing.T) {
	var (
		testAPIError = errors.New("test")
	)

	localizer, _ := goi18n.New(nil)
	factory := factories.NewDefaultFactory(nil, nil, nil, localizer)

	t.Run("no response and no api error", func(t *testing.T) {
		err := utils.GetAPIError(factory, nil, nil)
		assert.NoError(t, err, "unexpected error if we have no response nor API error")
	})

	t.Run("api error is returned", func(t *testing.T) {
		err := utils.GetAPIError(factory, nil, testAPIError)
		assert.Equal(t, testAPIError, err, "GetAPIError should return the same error passed in")
	})
}
