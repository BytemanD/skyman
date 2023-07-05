package compute

import (
	"os"

	"github.com/BytemanD/stackcrud/openstack/identity"
)

func GetComputeClientV2(authClient identity.V3AuthClient) (ComputeClientV2, error) {
	regionName := os.Getenv("OS_REGION_NAME")
	if regionName == "" {
		regionName = "RegionOne"
	}
	endpoint, err := authClient.GetEndpointFromCatalog(
		identity.TYPE_COMPUTE, identity.INTERFACE_PUBLIC, regionName)

	if err != nil {
		return ComputeClientV2{}, err
	}
	return ComputeClientV2{AuthClient: authClient, Endpoint: endpoint}, nil
}
