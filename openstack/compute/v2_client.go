package compute

import (
	"encoding/json"
	"fmt"

	"github.com/BytemanD/stackcrud/openstack/identity"
)

type VersionBody struct {
	Version ServerVersion `json:"version"`
}
type ServerVersion struct {
	MinVersion string `json:"min_version"`
	Version    string `json:"version"`
}

type ComputeClientV2 struct {
	AuthClient    identity.V3AuthClient
	Endpoint      string
	ServerVersion ServerVersion
	BaseHeaders   map[string]string
}

func (computeClient *ComputeClientV2) getUrl(resource string, id string) string {
	url := fmt.Sprintf("%s/%s", computeClient.Endpoint, resource)
	if id != "" {
		url += "/" + id
	}
	return url
}

// X-OpenStack-Nova-API-Version
func (computeClient *ComputeClientV2) UpdateVersion() {
	resp, _ := computeClient.AuthClient.Get(computeClient.Endpoint, nil, nil)
	versionBody := VersionBody{}
	json.Unmarshal(resp.Body, &versionBody)
	computeClient.BaseHeaders = map[string]string{}
	computeClient.ServerVersion = versionBody.Version
	computeClient.BaseHeaders["OpenStack-API-Versionn"] = computeClient.ServerVersion.Version
	computeClient.BaseHeaders["X-OpenStack-Nova-API-Version"] = computeClient.ServerVersion.Version
}
