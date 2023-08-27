package common

import (
	"bytes"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func GetIndentJson(v interface{}) (string, error) {
	jsonBytes, _ := json.Marshal(v)
	var buffer bytes.Buffer
	json.Indent(&buffer, jsonBytes, "", "\t")
	return buffer.String(), nil
}
func GetYaml(v interface{}) (string, error) {
	jsonString, err := GetIndentJson(v)
	if err != nil {
		return "", nil
	}
	bytes := []byte(jsonString)
	var out interface{}
	yaml.Unmarshal(bytes, &out)
	yamlBytes, err := yaml.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}
func splitTitle(s string) string {
	newStr := ""
	for _, c := range s {
		if c < 91 && newStr != "" {
			newStr += " " + string(c)
		} else {
			newStr += string(c)
		}
	}
	return newStr
}
