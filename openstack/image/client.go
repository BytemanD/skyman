package image

import (
	"os"

	"github.com/BytemanD/stackcrud/openstack/identity"
)

func GetImageClientV2(authClient identity.V3AuthClient) (ImageClientV2, error) {
	regionName := os.Getenv("OS_REGION_NAME")
	if regionName == "" {
		regionName = "RegionOne"
	}
	endpoint, err := authClient.GetEndpointFromCatalog(
		identity.TYPE_IMAGE, identity.INTERFACE_PUBLIC, regionName)

	if err != nil {
		return ImageClientV2{}, err
	}
	return ImageClientV2{AuthClient: authClient, Endpoint: endpoint}, nil
}
