package utility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	uuid "github.com/satori/go.uuid"
)

func IsUUID(s string) bool {
	uuid.NewV4()
	if _, err := uuid.FromString(s); err != nil {
		return false
	} else {
		return true
	}
}

func UrlJoin(path ...string) string {
	return strings.Join(path, "/")
}

func StringsContain(stringList []string, s string) bool {
	for _, str := range stringList {
		if s == str {
			return true
		}
	}
	return false
}
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

func LogError(err error, message string, exit bool) {
	if err == nil {
		return
	}
	if httpError, ok := err.(*HttpError); ok {
		logging.Error("%s, %s: %s", message, httpError.Reason, httpError.Message)
	} else {
		logging.Error("%s: %v", message, err)
	}
	if exit {
		os.Exit(1)
	}
}

func VersionUrl(endpoint, version string) string {
	parsedUrl, _ := url.Parse(endpoint)
	if !strings.HasPrefix(parsedUrl.Path, fmt.Sprintf("/%s", version)) {
		return UrlJoin(endpoint, version)
	}
	return endpoint
}
