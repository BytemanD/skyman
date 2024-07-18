package httpclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/go-resty/resty/v2"
)

const (
	CONTENT_TYPE      = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"
)

func getReqContentType(req *resty.Request) string {
	return req.Header.Get(CONTENT_TYPE)
}
func getRespContentType(resp *resty.Response) string {
	return resp.Header().Get(CONTENT_TYPE)
}

type RESTClient struct {
	session     *resty.Client
	Timeout     time.Duration
	AuthPlugin  AuthPluginInterface
	BaseHeaders map[string]string
}

func (c RESTClient) getSession() *resty.Client {
	if c.session == nil {
		c.session = resty.New()
		c.session.SetHeaders(c.BaseHeaders).SetTimeout(c.Timeout)
	}
	return c.session
}
func (c RESTClient) getRequest(method, url string) *resty.Request {
	req := c.getSession().SetHeaders(c.BaseHeaders).R()
	req.Method, req.URL = method, url
	return req
}
func (c RESTClient) logReq(req *resty.Request) {
	encodedHeader := ""
	if c.AuthPlugin != nil {
		encodedHeader = encodeHeaders(c.AuthPlugin.GetSafeHeader(req.Header))
	} else {
		encodedHeader = encodeHeaders(req.Header)
	}
	body := ""
	if getReqContentType(req) == CONTENT_TYPE_JSON {
		data, _ := json.Marshal(&req.Body)
		body = string(data)
	} else {
		body = "<Omitted, octet-stream>"
	}
	logging.Debug("REQ: %s %s, \n    Headers: %v \n    Body: %v",
		req.Method, req.URL, encodedHeader, body)
}
func (c RESTClient) logResp(resp *resty.Response) {
	respBody := ""
	if getRespContentType(resp) == CONTENT_TYPE_JSON {
		respBody = string(resp.Body())
	} else {
		respBody = "<octet-steam>"
	}
	logging.Debug("RESP: [%d], \n    Headers: %v\n    Body: %s",
		resp.StatusCode(), resp.Header(), respBody)
}
func (c RESTClient) Request(req *resty.Request) (*resty.Response, error) {
	if c.AuthPlugin != nil {
		if err := c.AuthPlugin.AuthRequest(req); err != nil {
			return nil, fmt.Errorf("auth request failed: %s", err)
		}
	}
	c.logReq(req)
	resp, err := req.Send()
	if err != nil {
		return nil, err
	}
	c.logResp(resp)
	return resp, MustNotError(resp)
}

func (c RESTClient) Get(url string, params, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodGet, url).SetHeaders(headers).SetQueryParams(params)
	return c.Request(req)
}
func (c RESTClient) Post(url string, body interface{}, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodPost, url).SetHeaders(c.BaseHeaders).SetHeaders(headers).SetBody(body)
	return c.Request(req)
}
func (c RESTClient) Put(url string, body []byte, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodPut, url).SetHeaders(headers).SetBody(body)
	return c.Request(req)
}
func (c RESTClient) Delete(url string, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodDelete, url).SetHeaders(headers)
	return c.Request(req)
}
func (c RESTClient) Patch(url string, body []byte, params, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodPatch, url).SetHeaders(headers).SetPathParams(params)
	return c.Request(req)
}
func (c *RESTClient) SetTimeout(timeout time.Duration) *RESTClient {
	c.Timeout = timeout
	return c
}

func New() *RESTClient {
	return &RESTClient{
		BaseHeaders: map[string]string{"Content-Type": "application/json"},
	}
}
