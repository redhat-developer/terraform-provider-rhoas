package utils

import (
	"encoding/json"

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
