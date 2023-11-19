package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/common"
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
func LogError(err error, message string, exit bool) {
	if err == nil {
		return
	}
	if httpError, ok := err.(*common.HttpError); ok {
		logging.Error("%s, %s: %s", message, httpError.Reason, httpError.Message)
	} else {
		logging.Error("%s, %v", message, err)
	}
	if exit {
		os.Exit(1)
	}
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

func Uname() string {
	utsname := &syscall.Utsname{}
	syscall.Uname(utsname)

	toString := func(ints [65]int8) string {
		out := make([]byte, 0, 64)
		for _, v := range ints {
			if v == 0 {
				break
			}
			out = append(out, uint8(v))
		}
		return string(out)
	}

	return fmt.Sprintf("%s %s %s %s",
		toString(utsname.Sysname), toString(utsname.Release),
		toString(utsname.Version), toString(utsname.Machine),
	)
}
