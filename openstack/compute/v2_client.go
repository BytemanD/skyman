package compute

import (
	"github.com/BytemanD/stackcrud/openstack/identity"
)

type VersionBody struct {
	Version Version `json:"version"`
}
type Version struct {
	MinVersion string `json:"min_version"`
	Version    string `json:"version"`
}

type ComputeClientV2 struct {
	identity.RestfuleClient
	Version     Version
	BaseHeaders map[string]string
}

// X-OpenStack-Nova-API-Version
func (client *ComputeClientV2) UpdateVersion() error {
	versionBody := VersionBody{}
	if err := client.Index(&versionBody); err != nil {
		return err
	}

	client.Version = versionBody.Version
	client.BaseHeaders["OpenStack-API-Versionn"] = client.Version.Version
	client.BaseHeaders["X-OpenStack-Nova-API-Version"] = client.Version.Version
	return nil
}

func GetComputeClientV2(authClient identity.V3AuthClient) (*ComputeClientV2, error) {
	if authClient.RegionName == "" {
		authClient.RegionName = "RegionOne"
	}
	endpoint, err := authClient.GetEndpointFromCatalog(
		identity.TYPE_COMPUTE, identity.INTERFACE_PUBLIC, authClient.RegionName)
	if err != nil {
		return nil, err
	}
	computeClient := ComputeClientV2{
		RestfuleClient: identity.RestfuleClient{
			V3AuthClient: authClient,
			Endpoint:     endpoint,
		},
		BaseHeaders: map[string]string{},
	}
	return &computeClient, nil
}
