package identity

import (
	"strings"

	"github.com/BytemanD/skyman/openstack/common"
)

type Service struct {
	common.Resource
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
}

type Endpoint struct {
	Id        string `json:"id"`
	Region    string `json:"region"`
	Url       string `json:"url"`
	Interface string `json:"interface"`
	RegionId  string `json:"region_id"`
	ServiceId string `json:"service_id"`
}

type ApiVersion struct {
	Id         string `json:"id"`
	MinVersion string `json:"min_version"`
	Status     string `json:"status"`
	Updated    string `json:"updated"`
	Version    string `json:"version"`
}

type ApiVersions []ApiVersion

func (client ApiVersions) Current() *ApiVersion {
	for _, version := range client {
		if strings.ToUpper(version.Status) == "CURRENT" {
			return &version
		}
	}
	return nil
}

func (client ApiVersions) Stable() *ApiVersion {
	for _, version := range client {
		if strings.ToUpper(version.Status) == "STABLE" {
			return &version
		}
	}
	return nil
}
