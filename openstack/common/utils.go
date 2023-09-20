package common

import (
	"bytes"
	"encoding/json"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

func GetIndentJson(v interface{}) (string, error) {
	jsonBytes, _ := json.Marshal(v)
	var buffer bytes.Buffer
	json.Indent(&buffer, jsonBytes, "", "  ")
	return buffer.String(), nil
}

func RaiseIfError(err error, msg string) {
	if err == nil {
		return
	}
	if httpError, ok := err.(*HttpError); ok {
		logging.Fatal("%s, %s, %s", msg, httpError.Reason, httpError.Message)
	} else {
		logging.Fatal("%s, %v", msg, err)
	}
}

func ContainsString(stringList []string, s string) bool {
	for _, str := range stringList {
		if s == str {
			return true
		}
	}
	return false
}
