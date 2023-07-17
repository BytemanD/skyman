// Openstack 认证客户端

package identity

const DEFAULT_TOKEN_EXPIRE_SECOND = 3600

type IdentityClientV3 struct {
	RestfuleClient
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
