package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/internal/auth_plugin"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/go-resty/resty/v2"
)

type ServiceClient struct {
	*resty.Client
	IsAdmin bool
}

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

func NewServiceClient(regionName string, sType, sName, sInterface string, version string, authPlugin auth_plugin.AuthPlugin) *ServiceClient {
	client := &ServiceClient{
		IsAdmin: authPlugin.IsAdmin(),
		Client:  session.DefaultRestyClient(""),
	}
	client.Client.OnBeforeRequest(func(c *resty.Client, _ *resty.Request) error {
		if c.BaseURL != "" {
			return nil
		}
		console.Debug("get endpoint for %s:%s:%s", sType, sName, sInterface)
		endpoint, err := authPlugin.GetEndpoint(regionName, sType, sName, sInterface)
		if err != nil {
			return err
		}
		u, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("parse endpoint falied: %w", err)
		}
		// u.Path 可能是 / 或 /vX.Y 或 /vX.Y/xxxxxxxxxx
		if !strings.HasPrefix(u.Path, "/"+version) {
			u = u.JoinPath(version)
		}
		c.BaseURL = u.String()
		return nil
	}).OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		return authPlugin.AuthRequest(r)
	}).OnBeforeRequest(session.LogBeforeRequest)
	return client
}
