package common

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GetYaml(v interface{}) (string, error) {
	jsonString, err := stringutils.JsonDumpsIndent(v)
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

func LastItems(items []interface{}, last int) []interface{} {
	return items[max(len(items)-max(last, 0), 0):]
}
