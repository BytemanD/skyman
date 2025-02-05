package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/openstack/internal/auth_plugin"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/go-resty/resty/v2"
)

type ServiceClient struct{ *resty.Client }

func (c *ServiceClient) BaserUrl() string {
	return c.BaseURL
}
func (c *ServiceClient) AddBaseHeader(k, v string) {
	if c.Client == nil {
		return
	}
	c.Client.SetHeader(k, v)
}
func (c *ServiceClient) Header() http.Header {
	return c.Client.Header
}
func (c *ServiceClient) IndexUrl() (string, error) {
	if c.Client.BaseURL == "" {
		return "", fmt.Errorf("endpoint is required")
	}
	parsed, err := url.Parse(c.BaserUrl())
	if err != nil {
		return "", fmt.Errorf("invalid endpoint: %s", c.BaserUrl())
	}
	parsed.Path = ""
	parsed.RawQuery = ""
	return parsed.String(), err
}

func (c *ServiceClient) Index(result interface{}) (*session.Response, error) {
	indexUrl, err := c.IndexUrl()
	if err != nil {
		return nil, fmt.Errorf("get index url failed: %s", err)
	}

	resp, err := c.Client.R().SetResult(result).Get(indexUrl)
	return &session.Response{Response: resp}, err
}

func NewServiceApi(endpoint string, version string, authPlugin auth_plugin.AuthPlugin) *ServiceClient {
	u, _ := url.Parse(endpoint)
	urlPath := strings.TrimSuffix(u.Path, "/")
	if !strings.HasPrefix(urlPath, fmt.Sprintf("/%s", version)) {
		u.Path = fmt.Sprintf("/%s", version)
	}
	return &ServiceClient{
		Client: session.DefaultRestyClient(u.String()).
			OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
				return authPlugin.AuthRequest(r)
			}),
	}
}
