package compute

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/identity"
	"github.com/BytemanD/skyman/openstack/keystoneauth"
)

type Version struct {
	MinVersion string `json:"min_version"`
	Version    string `json:"version"`
}

type ComputeClientV2 struct {
	identity.IdentityClientV3

	MicroVersion identity.ApiVersion
	BaseHeaders  map[string]string
	endpoint     string
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

func (client *ComputeClientV2) String() string {
	return fmt.Sprintf("<%s %s>", client.endpoint, client.Auth.RegionName)
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

func (client *ComputeClientV2) Index() (*common.Response, error) {
	return client.Request(common.NewIndexRequest(client.endpoint, nil, client.BaseHeaders))
}
func (client *ComputeClientV2) GetCurrentVersion() (*identity.ApiVersion, error) {
	resp, err := client.Index()
	if err != nil {
		return nil, err
	}
	versions := map[string]identity.ApiVersions{"versions": {}}
	resp.BodyUnmarshal(&versions)
	return versions["versions"].Current(), nil
}

// X-OpenStack-Nova-API-Version
func (client *ComputeClientV2) UpdateVersion() error {
	version, err := client.GetCurrentVersion()
	if err != nil {
		return err
	}
	client.MicroVersion = *version
	client.BaseHeaders["OpenStack-API-Versionn"] = client.MicroVersion.Version
	client.BaseHeaders["X-OpenStack-Nova-API-Version"] = client.MicroVersion.Version
	return nil
}

func GetComputeClientV2(idendityClient identity.IdentityClientV3) (*ComputeClientV2, error) {
	endpoint, err := idendityClient.GetServiceEndpoint(keystoneauth.TYPE_COMPUTE, "", keystoneauth.INTERFACE_PUBLIC)
	if err != nil {
		return nil, err
	}
	computeClient := ComputeClientV2{
		IdentityClientV3: idendityClient,
		endpoint:         endpoint,
		BaseHeaders: map[string]string{
			"Content-Type": "application/json",
		},
	}
	return &computeClient, nil
}
