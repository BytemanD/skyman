package utility

import (
	"net/http"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/utility/httpclient"
	"github.com/go-resty/resty/v2"
)

const (
	CONTENT_TYPE        = "Content-Type"
	CONTENT_LENGTH      = "Content-Length"
	CONTENT_TYPE_JSON   = "application/json"
	CONTENT_TYPE_STREAM = "application/octet-stream"
)

func DefaultRestyClient() *resty.Client {
	return resty.New().
		SetHeader(CONTENT_TYPE, CONTENT_TYPE_JSON).
		SetPreRequestHook(func(c *resty.Client, r *http.Request) error {
			logging.Debug("REQ: %s %s    \nHeader: %v", r.Method, r.URL, httpclient.EncodeHeaders(r.Header))
			return nil
		}).
		OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			logging.Debug("RESP: [%d] content-length: %s", r.StatusCode(), r.Header().Get(CONTENT_LENGTH))
			return nil
		})
}
