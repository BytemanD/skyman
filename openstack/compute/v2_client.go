package compute

import (
	"strconv"

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
	MicroVersion Version
	BaseHeaders  map[string]string
}

func (client *ComputeClientV2) GetMicroVersion() float64 {
	migroVersion, _ := strconv.ParseFloat(client.MicroVersion.Version, 64)
	return migroVersion
}

func (client *ComputeClientV2) MicroVersionLargeThen(version float64) bool {
	microVersion := client.GetMicroVersion()
	return (microVersion) >= version
}

// X-OpenStack-Nova-API-Version
func (client *ComputeClientV2) UpdateVersion() error {
	versionBody := VersionBody{}
	if err := client.Index(&versionBody); err != nil {
		return err
	}

	client.MicroVersion = versionBody.Version
	client.BaseHeaders["OpenStack-API-Versionn"] = client.MicroVersion.Version
	client.BaseHeaders["X-OpenStack-Nova-API-Version"] = client.MicroVersion.Version
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
