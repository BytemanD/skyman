package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/internal/auth_plugin"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
)

type ReqBody map[string]map[string]any

type ServiceClient struct {
	*resty.Client
	IsAdmin bool
}

func (c *ServiceClient) BaserUrl() string {
	return c.BaseURL
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

func QueryResource[T any](c *ServiceClient, u string, query url.Values, bodyKey string) ([]T, error) {
	result := map[string]any{}
	_, err := c.R().SetQueryParamsFromValues(query).SetResult(&result).Get(u)
	if err != nil {
		return nil, err
	}
	if itemBytes, err := json.Marshal(result[bodyKey]); err != nil {
		return nil, err
	} else {
		items := []T{}
		return items, json.Unmarshal(itemBytes, &items)
	}
}

func GetResource[T any](c *ServiceClient, u string, bodyKey string) (*T, error) {
	result := map[string]*T{}
	_, err := c.R().SetResult(&result).Get(u)
	return result[bodyKey], err
}
func DeleteResource(c *ServiceClient, u string, query ...url.Values) error {
	_, err := DeleteResourceWithResp(c, u, query...)
	return err
}
func DeleteResourceWithResp(c *ServiceClient, u string, query ...url.Values) (*resty.Response, error) {
	req := c.R()
	if len(query) > 0 {
		req.SetQueryParamsFromValues(query[0])
	}
	return req.Delete(u)
}
func FindResource[T any](c *ServiceClient, u string, bodyKey string) (*T, error) {

	result := map[string]*T{}
	_, err := c.R().SetResult(&result).Get(u)
	return result[bodyKey], err
}

func QueryByIdOrName[T any](
	idOrName string,
	showFunc func(id string) (*T, error),
	listFunc func(query url.Values) ([]T, error),
) (*T, error) {
	if item, err := showFunc(idOrName); err == nil {
		return item, nil
	} else if !errors.Is(err, session.ErrHTTP404) && !errors.Is(err, ErrResourceNotFound) {
		return nil, err
	}
	if items, err := listFunc(url.Values{"name": []string{idOrName}}); err != nil {
		return nil, err
	} else {
		fileted := lo.Filter(items, func(item T, _ int) bool {
			valueName := reflect.ValueOf(item).FieldByName("Name")
			return valueName.Kind() == reflect.String && valueName.String() == idOrName
		})
		switch len(fileted) {
		case 0:
			return nil, fmt.Errorf("%w with id or name %s", ErrResourceNotFound, idOrName)
		case 1:
			return &fileted[0], nil
		default:
			return nil, fmt.Errorf("%w with id or name: %s", ErrResourceMulti, idOrName)
		}
	}
}
