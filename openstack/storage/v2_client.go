package storage

import (
	"github.com/BytemanD/skyman/openstack/identity"
)

type StorageClientV2 struct {
	identity.RestfuleClient
	BaseHeaders map[string]string
}

func (client StorageClientV2) UpdateVersion() {

}

func GetStorageClientV2(authClient identity.V3AuthClient) (*StorageClientV2, error) {
	if authClient.RegionName == "" {
		authClient.RegionName = "RegionOne"
	}
	endpoint, err := authClient.GetEndpointFromCatalog(
		identity.TYPE_VOLUME_V2, identity.INTERFACE_PUBLIC, authClient.RegionName)
	if err != nil {
		return nil, err
	}
	return &StorageClientV2{
		RestfuleClient: identity.RestfuleClient{
			V3AuthClient: authClient,
			Endpoint:     endpoint,
		},
		BaseHeaders: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}
