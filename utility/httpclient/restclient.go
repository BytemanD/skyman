package httpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/go-resty/resty/v2"
)

const (
	CONTENT_TYPE        = "Content-Type"
	CONTENT_TYPE_JSON   = "application/json"
	CONTENT_TYPE_STREAM = "application/octet-stream"
)

func getReqContentType(req *resty.Request) string {
	return req.Header.Get(CONTENT_TYPE)
}
func getRespContentType(resp *resty.Response) string {
	return resp.Header().Get(CONTENT_TYPE)
}
func encodeHeaders(headers http.Header) string {
	headersString := []string{}
	for k, v := range headers {
		headersString = append(headersString, fmt.Sprintf("'%s: %s'", k, strings.Join(v, ",")))
	}
	return strings.Join(headersString, " ")
}

type RESTClient struct {
	session       *resty.Client
	Timeout       time.Duration
	RetryWaitTime time.Duration
	RetryCount    int
	AuthPlugin    AuthPluginInterface
	BaseHeaders   map[string]string
	sessionLock   *sync.Mutex
}

func (c *RESTClient) getSession() *resty.Client {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()

	if c.session == nil {
		c.session = resty.New()
		c.session.SetHeaders(c.BaseHeaders).SetTimeout(c.Timeout).
			SetRetryCount(c.RetryCount).SetRetryWaitTime(c.RetryWaitTime)
	}
	return c.session
}
func (c *RESTClient) getRequest(method, url string) *resty.Request {
	req := c.getSession().SetHeaders(c.BaseHeaders).R()
	req.Method, req.URL = method, url
	return req
}
func (c *RESTClient) GetRequest(method, url string) *resty.Request {
	return c.getRequest(method, url)
}
func (c *RESTClient) logReq(req *resty.Request) {
	encodedHeader := ""
	if c.AuthPlugin != nil {
		encodedHeader = encodeHeaders(c.AuthPlugin.GetSafeHeader(req.Header))
	} else {
		encodedHeader = encodeHeaders(req.Header)
	}
	body := ""
	if getReqContentType(req) == CONTENT_TYPE_STREAM {
		body = "<Omitted, octet-stream>"
	} else {
		data, _ := json.Marshal(&req.Body)
		body = string(data)
	}
	logging.Debug("REQ: %s %s, Query: %v\n    Headers: %v \n    Body: %v",
		req.Method, req.URL, req.QueryParam.Encode(), encodedHeader, body)
}
func (c *RESTClient) logResp(resp *resty.Response) {
	respBody := ""
	if getRespContentType(resp) == CONTENT_TYPE_STREAM {
		respBody = "<octet-steam>"
	} else {
		respBody = string(resp.Body())
	}
	logging.Debug("RESP: [%d], \n    Headers: %v\n    Body: %s",
		resp.StatusCode(), resp.Header(), respBody)
}
func (c *RESTClient) Request(req *resty.Request) (*resty.Response, error) {
	if c.AuthPlugin != nil {
		if err := c.AuthPlugin.AuthRequest(req); err != nil {
			return nil, fmt.Errorf("auth request failed: %s", err)
		}
	}
	c.logReq(req)

	var (
		resp *resty.Response
		err  error
	)
	resp, err = req.Send()
	if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
		logging.Error("connectoin reset by peer, Timeout: %d RetryWaitTime:%d RetryCount:%d",
			c.Timeout, c.RetryWaitTime, c.RetryCount)
	}
	if err != nil {
		return nil, err
	}
	c.logResp(resp)
	return resp, MustNotError(resp)
}

func (c *RESTClient) Get(url string, query url.Values, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodGet, url).SetHeaders(c.BaseHeaders).SetQueryParamsFromValues(query)
	if headers != nil {
		req.SetHeaders(headers)
	}
	return c.Request(req)
}
func (c *RESTClient) Post(url string, body interface{}, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodPost, url).SetHeaders(c.BaseHeaders).SetBody(body)
	if headers != nil {
		req.SetHeaders(headers)
	}
	return c.Request(req)
}
func (c *RESTClient) Put(url string, body interface{}, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodPut, url).SetHeaders(c.BaseHeaders).SetBody(body)
	if headers != nil {
		req.SetHeaders(headers)
	}
	return c.Request(req)
}
func (c *RESTClient) Delete(url string, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodDelete, url).SetHeaders(c.BaseHeaders)
	if headers != nil {
		req.SetHeaders(headers)
	}
	return c.Request(req)
}
func (c *RESTClient) Patch(url string, body interface{}, query url.Values, headers map[string]string) (*resty.Response, error) {
	req := c.getRequest(resty.MethodPatch, url).SetHeaders(c.BaseHeaders).
		SetBody(body).SetQueryParamsFromValues(query)
	if headers != nil {
		req.SetHeaders(headers)
	}
	return c.Request(req)
}

func (c *RESTClient) SetAuthPlugin(authPlugin AuthPluginInterface) *RESTClient {
	c.AuthPlugin = authPlugin
	return c
}

func New() *RESTClient {
	return &RESTClient{
		BaseHeaders: map[string]string{"Content-Type": "application/json"},
		sessionLock: &sync.Mutex{},
	}
}
