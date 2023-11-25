package storage

import (
	"github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/identity"
	"github.com/BytemanD/skyman/openstack/keystoneauth"
)

type StorageClientV2 struct {
	identity.IdentityClientV3

	BaseHeaders map[string]string
	endpoint    string
}

func (client StorageClientV2) UpdateVersion() {

}
func (client *StorageClientV2) Index() (*common.Response, error) {
	return client.Request(common.NewIndexRequest(client.endpoint, nil, client.BaseHeaders))
}
func (client *StorageClientV2) GetCurrentVersion() (*identity.ApiVersion, error) {
	resp, err := client.Index()
	if err != nil {
		return nil, err
	}
	versions := map[string]identity.ApiVersions{"versions": {}}
	resp.BodyUnmarshal(&versions)
	return versions["versions"].Current(), nil
}

func GetStorageClientV2(authClient identity.IdentityClientV3) (*StorageClientV2, error) {
	endpoint, err := authClient.Auth.GetServiceEndpoint(
		keystoneauth.TYPE_VOLUME_V2, "", keystoneauth.INTERFACE_PUBLIC)
	if err != nil {
		return nil, err
	}
	return &StorageClientV2{
		IdentityClientV3: authClient,
		endpoint:         endpoint,
		BaseHeaders:      map[string]string{"Content-Type": "application/json"},
	}, nil
}
