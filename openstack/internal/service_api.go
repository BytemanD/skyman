package internal

import (
	"encoding/json"
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

func (c *ServiceClient) Index(result any) (*resty.Response, error) {
	resp, err := c.Client.R().SetResult(nil).Get("{index}")
	if err == nil && !resp.IsError() && result != nil {
		err = json.Unmarshal([]byte(resp.Body()), result)
	}
	return resp, err
}

func urlWithVersion(urlPath, version string) (string, error) {
	if strings.HasPrefix(urlPath, "/"+version) {
		return urlPath, nil
	}
	return url.JoinPath(version, urlPath)
}
func urlWithoutVersion(u string) (string, error) {
	if u, err := url.Parse(u); err != nil {
		return "", err
	} else {
		u.Path = ""
		return u.String(), nil
	}
}
func NewServiceClient(regionName string, sType, sName, sInterface string, version string, authPlugin auth_plugin.AuthPlugin) *ServiceClient {
	client := &ServiceClient{
		IsAdmin: authPlugin.IsAdmin(),
		Client:  session.DefaultRestyClient(""),
	}
	client.Client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		if c.BaseURL != "" {
			if u, err := url.Parse(c.BaseURL); err != nil {
				return err
			} else {
				if newPath, err := urlWithVersion(u.Path, version); err != nil {
					return err
				} else {
					u.Path = newPath
				}
				c.BaseURL = u.String()
			}
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
		c.SetBaseURL(u.String())
		return nil
	}).OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		if r.URL == "{index}" {
			if u, err := urlWithoutVersion(c.BaseURL); err != nil {
				return err
			} else {
				r.URL = u
				return nil
			}
		}
		return authPlugin.AuthRequest(r)
	}).OnBeforeRequest(session.LogBeforeRequest)
	return client
}
