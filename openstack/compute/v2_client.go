package compute

import (
	"encoding/json"

	"github.com/BytemanD/stackcrud/openstack/common"
	"github.com/BytemanD/stackcrud/openstack/identity"
)

type ComputeClientV2 struct {
	common.ResourceClient
}

// X-OpenStack-Nova-API-Version
func (computeClient ComputeClientV2) UpdateVersion() {
	resp, _ := computeClient.AuthClient.Get(computeClient.Endpoint, nil, nil)
	versionBody := common.VersionBody{}
	json.Unmarshal(resp.Body, &versionBody)
	computeClient.Version = versionBody.Version
	computeClient.BaseHeaders["OpenStack-API-Versionn"] = computeClient.Version.Version
	computeClient.BaseHeaders["X-OpenStack-Nova-API-Version"] = computeClient.Version.Version
	computeClient.BaseHeaders["User-Agent"] = "go-stackcurd"
}

func GetComputeClientV2(authClient identity.V3AuthClient) (ComputeClientV2, error) {
	if authClient.RegionName == "" {
		authClient.RegionName = "RegionOne"
	}
	endpoint, err := authClient.GetEndpointFromCatalog(
		identity.TYPE_COMPUTE, identity.INTERFACE_PUBLIC, authClient.RegionName)

	if err != nil {
		return ComputeClientV2{}, err
	}
	return ComputeClientV2{
		common.ResourceClient{
			AuthClient: authClient, Endpoint: endpoint,
			BaseHeaders: map[string]string{},
		},
	}, nil
}
