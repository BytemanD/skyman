package image

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

type ImageClientV2 struct {
	AuthClient    identity.V3AuthClient
	Endpoint      string
	ServerVersion ServerVersion
	BaseHeaders   map[string]string
}

func (client ImageClientV2) getUrl(resource string, id string) string {
	url := fmt.Sprintf("%s/v2/%s", client.Endpoint, resource)
	if id != "" {
		url += "/" + id
	}
	return url
}

// X-OpenStack-Nova-API-Version
func (client ImageClientV2) UpdateVersion() {
	resp, _ := client.AuthClient.Get(client.Endpoint, nil, nil)
	versionBody := VersionBody{}
	json.Unmarshal(resp.Body, &versionBody)
	client.BaseHeaders = map[string]string{}
	client.ServerVersion = versionBody.Version
	client.BaseHeaders["OpenStack-API-Versionn"] = client.ServerVersion.Version
	client.BaseHeaders["X-OpenStack-GLANCE-API-Version"] = client.ServerVersion.Version
}
