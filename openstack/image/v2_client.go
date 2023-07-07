package image

import (
	"github.com/BytemanD/stackcrud/openstack/common"
	"github.com/BytemanD/stackcrud/openstack/identity"
)

type ImageClientV2 struct {
	common.ResourceClient
}

// X-OpenStack-Nova-API-Version
func (client ImageClientV2) UpdateVersion() {
	if client.APiVersion == "" {
		client.APiVersion = "v2"
	}
	client.BaseHeaders["User-Agent"] = "go-stackcurd"
}

func GetImageClientV2(authClient identity.V3AuthClient) (ImageClientV2, error) {
	if authClient.RegionName == "" {
		authClient.RegionName = "RegionOne"
	}
	endpoint, err := authClient.GetEndpointFromCatalog(
		identity.TYPE_IMAGE, identity.INTERFACE_PUBLIC, authClient.RegionName)

	if err != nil {
		return ImageClientV2{}, err
	}
	return ImageClientV2{
		ResourceClient: common.ResourceClient{
			AuthClient: authClient, Endpoint: endpoint, APiVersion: "v2",
			BaseHeaders: map[string]string{},
		},
	}, nil
}
