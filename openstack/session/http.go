package session

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/go-resty/resty/v2"
)

const (
	CONTENT_TYPE        = "Content-Type"
	CONTENT_LENGTH      = "Content-Length"
	CONTENT_TYPE_JSON   = "application/json"
	CONTENT_TYPE_STREAM = "application/octet-stream"

	HEADER_REQUEST_ID = "X-Openstack-Request-Id"
)

func EncodeHeaders(headers http.Header) string {
	headersString := []string{}
	for k, v := range headers {
		headersString = append(headersString, fmt.Sprintf("'%s: %s'", k, strings.Join(v, ",")))
	}
	return strings.Join(headersString, ", ")
}
func LogRequestPre(c *resty.Client, r *http.Request) error {
	console.Debug("REQ: %s %s\n    Header: %v", r.Method, r.URL, EncodeHeaders(r.Header))
	return nil
}
func LogRespAfterResponse(c *resty.Client, r *resty.Response) error {
	console.Debug("RESP: [%d] content-length: %s\n    Body: %s",
		r.StatusCode(), r.Header().Get(CONTENT_LENGTH), string(r.Body()))
	return nil
}

// 默认的 Client
//
// 记录请求日志，设置content-type=application/json
func DefaultRestyClient() *resty.Client {
	return resty.New().
		SetHeader(CONTENT_TYPE, CONTENT_TYPE_JSON).
		SetPreRequestHook(LogRequestPre).
		OnAfterResponse(LogRespAfterResponse)
}
