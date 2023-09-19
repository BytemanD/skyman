package networking

import (
	"github.com/BytemanD/skyman/openstack/identity"
)

type NeutronClientV2 struct {
	identity.RestfuleClient
	BaseHeaders map[string]string
}

func GetNeutronClientV2(authClient identity.V3AuthClient) (*NeutronClientV2, error) {
	if authClient.RegionName == "" {
		authClient.RegionName = "RegionOne"
	}
	endpoint, err := authClient.GetEndpointFromCatalog(
		identity.TYPE_NETWORK, identity.INTERFACE_PUBLIC, authClient.RegionName)

	if err != nil {
		return nil, err
	}
	return &NeutronClientV2{
		RestfuleClient: identity.RestfuleClient{
			V3AuthClient: authClient,
			Endpoint:     endpoint + "/v2.0",
		},
		BaseHeaders: map[string]string{},
	}, nil
}
