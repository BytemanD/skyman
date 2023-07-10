package storage

import (
	// "github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/openstack/common"
	"github.com/BytemanD/stackcrud/openstack/identity"
)

type StorageClientV2 struct {
	common.ResourceClient
}

func (client StorageClientV2) UpdateVersion() {
	client.BaseHeaders["User-Agent"] = "go-stackcurd"
}

func GetStorageClientV2(authClient identity.V3AuthClient) (StorageClientV2, error) {
	if authClient.RegionName == "" {
		authClient.RegionName = "RegionOne"
	}
	endpoint, err := authClient.GetEndpointFromCatalog(
		identity.TYPE_VOLUME_V2, identity.INTERFACE_PUBLIC, authClient.RegionName)
	if err != nil {
		return StorageClientV2{}, err
	}
	return StorageClientV2{
		ResourceClient: common.ResourceClient{
			AuthClient: authClient, Endpoint: endpoint,
			BaseHeaders: map[string]string{},
		},
	}, nil
}
