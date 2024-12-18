package internal

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/go-resty/resty/v2"
)

const (
	CONTENT_TYPE        = "Content-Type"
	CONTENT_TYPE_JSON   = "application/json"
	CONTENT_TYPE_STREAM = "application/octet-stream"
)

type ServiceClient struct {
	Endpoint   string
	Version    string
	AuthPlugin *auth.PasswordAuthPlugin
	rawClient  *resty.Client
}

func (s ServiceClient) NewRequest() *resty.Request {
	return s.rawClient.R()
}

func (s ServiceClient) Get(url string) (*resty.Response, error) {
	return s.rawClient.R().Get(url)
}

func NewServiceApi(endpoint string, version string, authPlugin *auth.PasswordAuthPlugin) *ServiceClient {
	u, _ := url.Parse(endpoint)
	urlPath := strings.TrimSuffix(u.Path, "/")
	if !strings.HasPrefix(urlPath, fmt.Sprintf("/%s", version)) {
		u.Path = fmt.Sprintf("/%s", version)
	}
	return &ServiceClient{
		Endpoint: u.String(),
		rawClient: resty.New().SetHeader(CONTENT_TYPE, CONTENT_TYPE_JSON).
			OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
				return authPlugin.AuthRequest(r)
			}),
	}
}
