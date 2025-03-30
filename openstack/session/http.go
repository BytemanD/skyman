package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
)

const (
	CONTENT_TYPE        = "Content-Type"
	CONTENT_LENGTH      = "Content-Length"
	CONTENT_TYPE_JSON   = "application/json"
	CONTENT_TYPE_STREAM = "application/octet-stream"

	HEADER_REQUEST_ID = "X-Openstack-Request-Id"

	DEFAULT_RETRY_COUNT         = 0
	DEFAULT_RETRY_WAIT_TIME     = time.Second
	DEFAULT_RETRY_MAX_WAIT_TIME = time.Second * 5
	DEFAULT_TIMEOUT             = time.Second * 60

	CODE_404 = 404
)

var ErrHTTPStatus = errors.New("http error")
var ErrHTTP404 = fmt.Errorf("%w: %d", ErrHTTPStatus, CODE_404)

func EncodeHeaders(reqHeader, clientHeader http.Header) []string {
	allKeys := lo.Uniq(append(lo.Keys(reqHeader), lo.Keys(clientHeader)...))
	sort.Strings(allKeys)
	return lo.FilterMap(allKeys, func(key string, _ int) (string, bool) {
		if key == "X-Auth-Token" {
			return fmt.Sprintf("%s: %s", key, "<token>"), true
		}
		if lo.HasKey(reqHeader, key) {
			return fmt.Sprintf("%s: %s", key, strings.Join(reqHeader[key], ", ")), true
		}
		if lo.HasKey(clientHeader, key) {
			return fmt.Sprintf("%s: %s", key, strings.Join(clientHeader[key], ", ")), true
		}
		return "", false
	})
}

func LogBeforeRequest(c *resty.Client, r *resty.Request) error {
	body := []byte{}
	if r.Header.Get(CONTENT_TYPE) != CONTENT_TYPE_STREAM && r.Body != nil {
		body, _ = json.Marshal(r.Body)
	}
	var u string
	if strings.HasPrefix(r.URL, "http://") || strings.HasPrefix(r.URL, "https://") {
		u = r.URL
	} else {
		u, _ = url.JoinPath(c.BaseURL, r.URL)
	}
	console.Debug("---- REQ ----: %s %s?%s\n    Header: %v\n    Body: %s",
		r.Method, u, r.QueryParam.Encode(),
		strings.Join(EncodeHeaders(r.Header, c.Header), "\n            "),
		body)
	return nil
}

func LogRespAfterResponse(c *resty.Client, r *resty.Response) error {
	console.Debug("---- RESP ----: [%d]\n    Header: %s\n    Body: %s",
		r.StatusCode(),
		strings.Join(EncodeHeaders(r.Header(), nil), "\n            "),
		string(r.Body()))
	return nil
}
func CheckStatusAfterResponse(c *resty.Client, r *resty.Response) error {
	if !r.IsError() {
		return nil
	}
	switch r.StatusCode() {
	case CODE_404:
		return fmt.Errorf("%w: %s", ErrHTTP404, r.Body())
	default:
		if r.IsError() {
			return fmt.Errorf("%w: [%d], %s", ErrHTTPStatus, r.StatusCode(), string(r.Body()))
		}
		return nil
	}
}

// 默认的 Client, 设置content-type=application/json
func DefaultRestyClient(baseUrl string) *resty.Client {
	return resty.New().SetBaseURL(baseUrl).
		SetHeader(CONTENT_TYPE, CONTENT_TYPE_JSON).
		SetRetryCount(DEFAULT_RETRY_COUNT).
		SetRetryWaitTime(DEFAULT_RETRY_WAIT_TIME).
		SetRetryMaxWaitTime(DEFAULT_RETRY_MAX_WAIT_TIME).
		OnAfterResponse(LogRespAfterResponse).
		OnAfterResponse(CheckStatusAfterResponse)
}
