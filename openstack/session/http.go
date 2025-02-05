package session

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/go-resty/resty/v2"
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
)

func EncodeHeaders(reqHeader, clientHeader http.Header) string {
	allHeaders := map[string][]string{}
	for k, v := range clientHeader {
		allHeaders[k] = v
	}
	for k, v := range reqHeader {
		allHeaders[k] = v
	}
	headerList := []string{}
	for k, v := range allHeaders {
		headerList = append(headerList, fmt.Sprintf("'%s: %s'", k, strings.Join(v, ",")))
	}
	return strings.Join(headerList, ", ")
}
func LogRequestPre(c *resty.Client, r *http.Request) error {
	console.Debug("REQ: %s %s\n    Header: %v", r.Method, r.URL, EncodeHeaders(r.Header, c.Header))
	return nil
}
func LogBeforeRequest(c *resty.Client, r *resty.Request) error {
	body := []byte{}
	if r.Body != nil {
		body, _ = json.Marshal(r.Body)
	}
	console.Debug("REQ: %s %s?%s\n    Header: %v\n    Body: %s",
		r.Method, r.URL, r.QueryParam.Encode(), EncodeHeaders(r.Header, c.Header), body)
	return nil
}

func LogRespAfterResponse(c *resty.Client, r *resty.Response) error {
	console.Debug("RESP: [%d]\n    Header: %s\n    Body: %s",
		r.StatusCode(), EncodeHeaders(r.Header(), nil), string(r.Body()))
	return nil
}

// 默认的 Client
//
// 记录请求日志，设置content-type=application/json
func DefaultRestyClient(baseUrl string) *resty.Client {
	return resty.New().SetBaseURL(baseUrl).
		SetHeader(CONTENT_TYPE, CONTENT_TYPE_JSON).
		SetRetryCount(DEFAULT_RETRY_COUNT).
		SetRetryWaitTime(DEFAULT_RETRY_WAIT_TIME).
		SetRetryMaxWaitTime(DEFAULT_RETRY_MAX_WAIT_TIME).
		SetTimeout(DEFAULT_TIMEOUT).
		OnBeforeRequest(LogBeforeRequest).
		OnAfterResponse(LogRespAfterResponse)
}
