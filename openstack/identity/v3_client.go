// Openstack 认证客户端

package identity

import (
	"strings"

	"github.com/BytemanD/skyman/openstack/keystoneauth"
)

const DEFAULT_TOKEN_EXPIRE_SECOND = 3600

type IdentityClientV3 struct {
	RestfuleClient
	BaseHeaders map[string]string
	Auth        keystoneauth.PasswordAuthPlugin
}

func GetIdentityClientV3(auth keystoneauth.PasswordAuthPlugin) (*IdentityClientV3, error) {
	endpoint, err := authClient.GetEndpointFromCatalog(
		TYPE_IDENTITY, INTERFACE_PUBLIC, authClient.RegionName)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(endpoint, "/v3") {
		endpoint += "/v3"
	}
	return &IdentityClientV3{
		RestfuleClient: RestfuleClient{
			Endpoint: endpoint,
		},
		Auth:        auth,
		BaseHeaders: map[string]string{"Content-Type": "application/json"},
	}, nil
}
