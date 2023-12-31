package networking

import (
	"strings"

	"github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/identity"
	"github.com/BytemanD/skyman/openstack/keystoneauth"
)

type NeutronClientV2 struct {
	identity.IdentityClientV3
	endpoint    string
	BaseHeaders map[string]string
}

func (client *NeutronClientV2) Index() (*common.Response, error) {
	return client.Request(common.NewIndexRequest(client.endpoint, nil, client.BaseHeaders))
}
func (client *NeutronClientV2) GetCurrentVersion() (*identity.ApiVersion, error) {
	resp, err := client.Index()
	if err != nil {
		return nil, err
	}
	versions := map[string]identity.ApiVersions{"versions": {}}
	resp.BodyUnmarshal(&versions)
	return versions["versions"].Current(), nil
}

func GetNeutronClientV2(authClient identity.IdentityClientV3) (*NeutronClientV2, error) {
	endpoint, err := authClient.GetServiceEndpoint(
		keystoneauth.TYPE_NETWORK, "", keystoneauth.INTERFACE_PUBLIC)

	if err != nil {
		return nil, err
	}
	if !strings.Contains(endpoint, "/v2") {
		endpoint += "/v2.0"
	}
	return &NeutronClientV2{
		IdentityClientV3: authClient,
		endpoint:         endpoint,
		BaseHeaders:      map[string]string{},
	}, nil
}
