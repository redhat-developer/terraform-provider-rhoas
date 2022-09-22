package utils

import (
	"encoding/json"
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
func GetAPIError(response *http.Response, apiError error) error {

	if apiError == nil && response == nil {
		return nil
	}

	if apiError == nil {
		responseBody, err := stringifyResponse(response)
		if err != nil {
			return err
		}

		return errors.Errorf("%s", responseBody)
	}

	if response == nil {
		return apiError
	}

	responseBody, err := stringifyResponse(response)
	if err != nil {
		return err
	}

	return errors.Errorf("%s%s", apiError.Error(), responseBody)

}

func stringifyResponse(response *http.Response) (string, error) {

	if response == nil {
		return "", nil
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
