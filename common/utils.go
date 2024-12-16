package common

import (
	"encoding/json"
	"fmt"
	"net"
	"path"
	"strconv"
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

func RepeatFunc(nums int, function func()) {
	for i := 0; i < nums; i++ {
		function()
	}
}
func ValidIpv4(cidr string, maxMask ...int) bool {
	values := strings.Split(cidr, "/")
	if len(values) == 2 {
		validMaxkMask := 32
		if len(maxMask) > 0 {
			validMaxkMask = maxMask[0]
		}
		if mask, err := strconv.Atoi(values[1]); err != nil {
			return false
		} else if mask <= 0 || mask > validMaxkMask {
			return false
		}
	}
	parsed := net.ParseIP(values[0])
	if parsed == nil || parsed.To4() == nil {
		return false
	}
	return true
}

func OneOfString(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return strs[len(strs)-1]
}

func OneOfNumber[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | float32 | float64](values ...T) T {
	if len(values) == 0 {
		return 0
	}
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return values[len(values)-1]
}
func OneOfBoolean(values ...bool) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if value {
			return value
		}
	}
	return values[len(values)-1]
}
func OneOfStringArrays(arraysList ...[]string) []string {
	if len(arraysList) == 0 {
		return []string{}
	}
	for _, arrays := range arraysList {
		if len(arrays) != 0 {
			return arrays
		}
	}
	return arraysList[len(arraysList)-1]
}
