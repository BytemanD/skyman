package image

import (
	"github.com/BytemanD/skyman/openstack/identity"
)

type ImageClientV2 struct {
	identity.RestfuleClient
	BaseHeaders map[string]string
}

func GetImageClientV2(authClient identity.V3AuthClient) (*ImageClientV2, error) {
	if authClient.RegionName == "" {
		authClient.RegionName = "RegionOne"
	}
	endpoint, err := authClient.GetEndpointFromCatalog(
		identity.TYPE_IMAGE, identity.INTERFACE_PUBLIC, authClient.RegionName)

	if err != nil {
		return nil, err
	}
	return &ImageClientV2{
		RestfuleClient: identity.RestfuleClient{
			V3AuthClient: authClient,
			Endpoint:     endpoint + "/v2",
		},
		BaseHeaders: map[string]string{},
	}, nil
}
