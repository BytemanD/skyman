package compute

import (
	"strconv"
	"strings"

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

type microVersion struct {
	Version      int
	MicroVersion int
}

func getMicroVersion(vertionStr string) microVersion {
	versionList := strings.Split(vertionStr, ".")
	v, _ := strconv.Atoi(versionList[0])
	micro, _ := strconv.Atoi(versionList[1])
	return microVersion{Version: v, MicroVersion: micro}
}

func (client *ComputeClientV2) MicroVersionLargeEqual(version string) bool {
	clientVersion := getMicroVersion(client.MicroVersion.Version)
	otherVersion := getMicroVersion(version)
	if clientVersion.Version > otherVersion.Version {
		return true
	} else if clientVersion.Version == otherVersion.Version {
		return clientVersion.MicroVersion >= otherVersion.MicroVersion
	} else {
		return false
	}
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
