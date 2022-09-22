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
	switch {
	case apiError == nil:
		return parseResponse(response)
	case response == nil:
		return apiError
	default:
		return errors.Errorf("API error: %v, response error: %v", apiError, parseResponse(response))
	}
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
