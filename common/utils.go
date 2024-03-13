package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GetIndentJson(v interface{}) (string, error) {
	jsonBytes, _ := json.Marshal(v)
	var buffer bytes.Buffer
	json.Indent(&buffer, jsonBytes, "", "    ")
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

func SplitKeyValue(kv string) ([]string, error) {
	kvList := strings.Split(kv, "=")
	if len(kvList) != 2 {
		return nil, fmt.Errorf("invalid key value %s", kv)
	} else {
		return kvList, nil
	}
}

func RegistryLongFlag(cmd ...*cobra.Command) {
	for _, c := range cmd {
		c.Flags().BoolP("long", "l", false, "List additional fields in output")
	}
}

func MarshalModel(obj interface{}, indent bool) string {
	var m []byte
	if indent {
		m, _ = json.MarshalIndent(obj, "", "  ")
	} else {
		m, _ = json.Marshal(obj)
	}
	return string(m)
}

func PathExtSplit(file string) (string, string) {
	ext := path.Ext(path.Base(file))
	name := strings.TrimSuffix(path.Base(file), ext)
	return name, ext
}
