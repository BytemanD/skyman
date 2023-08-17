// Openstack 认证客户端

package identity

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
	return &IdentityClientV3{
		RestfuleClient: RestfuleClient{
			V3AuthClient: authClient,
			Endpoint:     endpoint,
		},
		BaseHeaders: map[string]string{"Content-Type": "application/json"},
	}, nil
}

// func (client IdentityClientV3) ServiceList() (*common.Response, error) {
// 	url := fmt.Sprintf("%s%s", client.AuthUrl, "/services")
// 	client.List("services")
// 	return client.Get(url, nil, map[string]string{})
// }

// func (client *V3AuthClient) UserList() (*common.Response, error) {
// 	url := fmt.Sprintf("%s%s", client.AuthUrl, "/users")
// 	return client.Get(url, nil, map[string]string{})
// }
