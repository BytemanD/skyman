package internal

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/openstack/auth"
	internal "github.com/BytemanD/skyman/openstack/internal/utility"
	"github.com/go-resty/resty/v2"
)

type ServiceClient struct {
	Endpoint   string
	AuthPlugin auth.AuthPlugin
	rawClient  *resty.Client
}

func (c *ServiceClient) Index(result interface{}) (*resty.Response, error) {
	if c.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	parsed, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %s", c.Endpoint)
	}
	for k := range parsed.Query() {
		delete(parsed.Query(), k)
	}
	return c.rawClient.R().SetResult(result).Get(parsed.String())
}

func NewServiceApi[T ServiceClient](endpoint string, version string, authPlugin auth.AuthPlugin) *T {
	u, _ := url.Parse(endpoint)
	urlPath := strings.TrimSuffix(u.Path, "/")
	if !strings.HasPrefix(urlPath, fmt.Sprintf("/%s", version)) {
		u.Path = fmt.Sprintf("/%s", version)
	}
	return &T{
		Endpoint:   u.String(),
		AuthPlugin: authPlugin,
		rawClient: internal.DefaultRestyClient().
			OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
				return authPlugin.AuthRequest(r)
			}),
	}
}
