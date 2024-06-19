package utility

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/BytemanD/easygo/pkg/compare"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/utility/httpclient"
	"github.com/fatih/color"
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

func VersionUrl(endpoint, version string) string {
	parsedUrl, _ := url.Parse(endpoint)
	if !strings.HasPrefix(parsedUrl.Path, fmt.Sprintf("/%s", version)) {
		return UrlJoin(endpoint, version)
	}
	return endpoint
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
