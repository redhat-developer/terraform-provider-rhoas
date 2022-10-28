package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"

	"github.com/pkg/errors"
)

// AsMap converts a JSON-tagged struct into a map
func AsMap(original interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(original)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	var obj map[string]interface{}
	err = json.Unmarshal(data, &obj)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	return obj, nil
}

// GetAPIError converts an http.Response and a RHOAS apiError into golang errors
func GetAPIError(factory rhoasAPI.Factory, response *http.Response, apiError error) error {

	if apiError == nil {
		return nil
	}

	if response == nil {
		return apiError
	}

	// this to support terraform acceptance tests which make it impossible to pass factory around
	// testing code should never affect actual code but no one actually about make good software
	if factory == nil {
		//nolint
		return fmt.Errorf("%v : %v", parseResponse(response).Error(), apiError.Error())
	}

	switch response.StatusCode {
	case http.StatusBadRequest:
		return fmt.Errorf(buildErrorString(factory.Localizer().MustLocalize("common.errors.api.badRequest"), response, apiError))
	case http.StatusUnauthorized:
		return fmt.Errorf(buildErrorString(factory.Localizer().MustLocalize("common.errors.api.unauthorized"), response, apiError))
	case http.StatusForbidden:
		return fmt.Errorf(buildErrorString(factory.Localizer().MustLocalize("common.errors.api.forbidden"), response, apiError))
	case http.StatusInternalServerError:
		return fmt.Errorf(buildErrorString(factory.Localizer().MustLocalize("common.errors.api.internalServerError"), response, apiError))
	case http.StatusServiceUnavailable:
		return fmt.Errorf(buildErrorString(factory.Localizer().MustLocalize("common.errors.api.serviceUnavailable"), response, apiError))
	case http.StatusConflict:
		return fmt.Errorf(buildErrorString(factory.Localizer().MustLocalize("common.errors.api.conflict"), response, apiError))
	case http.StatusNotFound:
		return fmt.Errorf(buildErrorString(factory.Localizer().MustLocalize("common.errors.api.notFound"), response, apiError))
	}

	return apiError
}

func buildErrorString(message string, response *http.Response, apiError error) string {
	return fmt.Sprintf("%v :: %v :: %v :: %v", message, apiError.Error(), response.Request.URL.Host+response.Request.URL.Path, response.Request.Method)
}

func parseResponse(response *http.Response) error {
	if response == nil {
		return nil
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "unable to read response body")
	}

	return errors.New(string(bodyBytes))
}

// CheckNotFound checks whether the response status code is not found
func CheckNotFound(response *http.Response) bool {
	return response.StatusCode == http.StatusNotFound
}
