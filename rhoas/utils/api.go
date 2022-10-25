package utils

import (
	"encoding/json"
	"fmt"
	rhoasAPI "github.com/redhat-developer/terraform-provider-rhoas/rhoas/api"
	"io"
	"net/http"

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
		return fmt.Errorf("%v : %v", parseResponse(response).Error(), apiError.Error())
	}

	switch response.StatusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("%v : %v", factory.Localizer().MustLocalize("common.errors.api.badRequest"), apiError.Error())
	case http.StatusUnauthorized:
		return fmt.Errorf("%v : %v", factory.Localizer().MustLocalize("common.errors.api.unauthorized"), apiError.Error())
	case http.StatusForbidden:
		return fmt.Errorf("%v : %v", factory.Localizer().MustLocalize("common.errors.api.forbidden"), apiError.Error())
	case http.StatusInternalServerError:
		return fmt.Errorf("%v : %v", factory.Localizer().MustLocalize("common.errors.api.internalServerError"), apiError.Error())
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%v : %v", factory.Localizer().MustLocalize("common.errors.api.serviceUnavailable"), apiError.Error())
	case http.StatusConflict:
		return fmt.Errorf("%v : %v", factory.Localizer().MustLocalize("common.errors.api.conflict"), apiError.Error())
	case http.StatusNotFound:
		return fmt.Errorf("%v : %v", factory.Localizer().MustLocalize("common.errors.api.notFound"), apiError.Error())
	}

	return apiError
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
