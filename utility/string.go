package utility

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/BytemanD/easygo/pkg/compare"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/utility/httpclient"
	"github.com/fatih/color"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
)

func UrlJoin(path ...string) string {
	return strings.Join(path, "/")
}

func RaiseIfError(err error, msg string) {
	if err == nil {
		return
	}
	if compare.IsType[httpclient.HttpError](err) {
		httpError, _ := err.(httpclient.HttpError)
		logging.Fatal("%s, %s: %s", msg, httpError.Reason, httpError.Message)
	} else {
		logging.Fatal("%s, %v", msg, err)
	}
}

func LogError(err error, message string, exit bool) {
	if err == nil {
		return
	}
	if compare.IsType[httpclient.HttpError](err) {
		httpError, _ := err.(httpclient.HttpError)
		logging.Error("%s, %s: %s", message, httpError.Reason, httpError.Message)
	} else {
		logging.Error("%s: %v", message, err)
	}
	if exit {
		os.Exit(1)
	}
}
func LogIfError(err error, exit bool, format string, args ...interface{}) {
	if err == nil {
		return
	}
	if compare.IsType[httpclient.HttpError](err) {
		httpError, _ := err.(httpclient.HttpError)
		logging.Error(fmt.Sprintf(format, args...)+": [%s] %s", httpError.Reason, httpError.Message)
	} else {
		logging.Error(fmt.Sprintf(format, args...)+": %v", err)
	}
	if exit {
		os.Exit(1)
	}
}
func VersionUrl(endpoint, version string) string {
	u, _ := url.Parse(endpoint)
	u.Path = strings.TrimSuffix(u.Path, "/")
	if !strings.HasPrefix(u.Path, fmt.Sprintf("/%s", version)) {
		u.Path = fmt.Sprintf("/%s", version)
	}
	result, _ := url.JoinPath(fmt.Sprintf("%s://%s", u.Scheme, u.Host), u.Path)
	return result
}

func GreenString(s string) string {
	return color.New(color.FgGreen).Sprintf(s)
}

func BlueString(s string) string {
	return color.New(color.FgBlue).Sprintf(s)
}
func RedString(s string) string {
	return color.New(color.FgRed).Sprintf(s)
}

func HumanBytes(value int) string {
	switch {
	case value >= TB:
		return fmt.Sprintf("%.2f TB", float32(value)/TB)
	case value >= GB:
		return fmt.Sprintf("%.2f GB", float32(value)/GB)
	case value >= MB:
		return fmt.Sprintf("%.2f MB", float32(value)/MB)
	case value >= KB:
		return fmt.Sprintf("%.2f KB", float32(value)/KB)
	default:
		return fmt.Sprintf("%d B", value)
	}
}

func MatchPingResult(text string) []string {
	reg := regexp.MustCompile(`(\d+) packets transmitted, (\d+) received,[^\n]+`)
	return reg.FindStringSubmatch(text)
}

func UnmarshalJsonKey(bytes []byte, key string, v any) error {
	tmp := map[string]interface{}{}
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return err
	}
	if tmpBytes, err := json.Marshal(tmp[key]); err != nil {
		return err
	} else {
		return json.Unmarshal(tmpBytes, &v)
	}
}
