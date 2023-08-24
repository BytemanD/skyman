// Openstack 认证客户端

package identity

import "strings"

const DEFAULT_TOKEN_EXPIRE_SECOND = 3600

type IdentityClientV3 struct {
	RestfuleClient
	BaseHeaders map[string]string
}

func GetIdentityClientV3(authClient V3AuthClient) (*IdentityClientV3, error) {
	if authClient.RegionName == "" {
		authClient.RegionName = "RegionOne"
	}

	endpoint, err := authClient.GetEndpointFromCatalog(
		TYPE_IDENTITY, INTERFACE_PUBLIC, authClient.RegionName)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(endpoint, "/v3") {
		endpoint += "/v3"
	}
	return &IdentityClientV3{
		RestfuleClient: RestfuleClient{
			V3AuthClient: authClient,
			Endpoint:     endpoint,
		},
		BaseHeaders: map[string]string{"Content-Type": "application/json"},
	}, nil
}
