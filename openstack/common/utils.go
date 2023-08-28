package common

import (
	"bytes"
	"encoding/json"
)

func GetIndentJson(v interface{}) (string, error) {
	jsonBytes, _ := json.Marshal(v)
	var buffer bytes.Buffer
	json.Indent(&buffer, jsonBytes, "", "  ")
	return buffer.String(), nil
}
