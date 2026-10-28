package jsonutils

import (
	"encoding/json"
)

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func MarshalString(v any) (string, error) {
	data, err := json.Marshal(v)
	return string(data), err
}
