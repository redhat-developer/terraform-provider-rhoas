package utils

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

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

func GetAPIError(response *http.Response, apiError error) (error, error) {

	if apiError == nil && response == nil {
		return nil, nil
	}

	if apiError == nil {
		responseBody, err := stringifyResponse(response)
		if err != nil {
			return nil, err
		}

		return errors.Errorf("%s", responseBody), nil
	}

	if response == nil {
		return apiError, nil
	}

	responseBody, err := stringifyResponse(response)
	if err != nil {
		return nil, err
	}

	return errors.Errorf("%s%s", apiError.Error(), responseBody), nil

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
