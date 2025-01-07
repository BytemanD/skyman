package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/openstack/internal/auth_plugin"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/go-resty/resty/v2"
)

const (
	HEADER_REQUEST_ID = "X-Openstack-Request-Id"
)

type Response struct{ *resty.Response }

func (resp Response) RequestId() string {
	return resp.Header().Get(HEADER_REQUEST_ID)
}
func (resp Response) UnmarshalBody(v interface{}) error {
	return json.Unmarshal(resp.Body(), &v)
}

type ServiceClient struct {
	Url        string
	AuthPlugin auth_plugin.AuthPlugin
	rawClient  *resty.Client
}

func (c *ServiceClient) AddBaseHeader(k, v string) {
	if c.rawClient == nil {
		return
	}
	c.rawClient.SetHeader(k, v)
}
func (c *ServiceClient) Header() http.Header {
	return c.rawClient.Header
}
func (c *ServiceClient) IndexUrl() (string, error) {
	if c.Url == "" {
		return "", fmt.Errorf("endpoint is required")
	}
	parsed, err := url.Parse(c.Url)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint: %s", c.Url)
	}
	parsed.Path = ""
	parsed.RawQuery = ""
	return parsed.String(), err
}

func (c *ServiceClient) Index(result interface{}) (*Response, error) {
	indexUrl, err := c.IndexUrl()
	if err != nil {
		return nil, fmt.Errorf("get index url failed: %s", err)
	}

	resp, err := c.rawClient.R().SetResult(result).Get(indexUrl)
	return &Response{resp}, err
}

func NewServiceApi[T ServiceClient](endpoint string, version string, authPlugin auth_plugin.AuthPlugin) *T {
	u, _ := url.Parse(endpoint)
	urlPath := strings.TrimSuffix(u.Path, "/")
	if !strings.HasPrefix(urlPath, fmt.Sprintf("/%s", version)) {
		u.Path = fmt.Sprintf("/%s", version)
	}
	return &T{
		Url:        u.String(),
		AuthPlugin: authPlugin,
		rawClient: session.DefaultRestyClient().
			OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
				return authPlugin.AuthRequest(r)
			}),
	}
}