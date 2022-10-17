package utils_test

import (
	"io"
	"net/http"
	"strings"
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
		testAPIError  = errors.New("test")
		responseError = errors.New("Internal Server Error")
		testResponse  = http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(responseError.Error())),
		}
	)

	t.Run("no response and no api error", func(t *testing.T) {
		err := utils.GetAPIError(nil, nil)
		assert.NoError(t, err, "unexpected error if we have no response nor API error")
	})
	t.Run("no response and api error", func(t *testing.T) {
		err := utils.GetAPIError(nil, testAPIError)
		assert.Error(t, err, "expecting an error if we have no response but an API error")
		assert.Equal(t, testAPIError, err, "got an unexpected error")
	})
	t.Run("response and no api error", func(t *testing.T) {
		testResponse.Body = io.NopCloser(strings.NewReader(responseError.Error())) // needed to reset the reader

		err := utils.GetAPIError(&testResponse, nil)
		assert.Error(t, err, "expecting an error if we have a response and no API error")
		assert.Equal(t, responseError.Error(), err.Error(), "got an unexpected error")
	})
	t.Run("response and api error", func(t *testing.T) {
		testResponse.Body = io.NopCloser(strings.NewReader(responseError.Error())) // needed to reset the reader

		err := utils.GetAPIError(&testResponse, testAPIError)
		assert.Error(t, err, "expecting an error if we have a response and an API error")

		want := errors.Errorf("API error: %v, response error: %v", testAPIError, responseError)
		assert.Equal(t, want.Error(), err.Error(), "got an unexpected error")
	})
	t.Run("unreadable response and no api error", func(t *testing.T) {
		testResponse.Body = erroringBuffer{} // needed to reset the reader

		err := utils.GetAPIError(&testResponse, nil)
		assert.Error(t, err, "expecting an error if we have a response and an API error")

		want := errors.New("unable to read response body: error reading body")
		assert.Equal(t, want.Error(), err.Error(), "got an unexpected error")
	})
}

type erroringBuffer struct {
}

func (mb erroringBuffer) Close() error {
	return nil
}

func (mb erroringBuffer) Read(p []byte) (n int, err error) {
	return 0, errors.New("error reading body")
}
