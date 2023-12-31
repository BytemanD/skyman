package identity

import (
	"fmt"
	"strings"
	"time"

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

func (v *ApiVersion) VersoinInfo() string {
	version := v.Id
	if v.MinVersion != "" {
		version += fmt.Sprintf(" (%s ~ %s)", v.MinVersion, v.Version)
	}
	return version
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

type Token struct {
	IsDomain  bool      `json:"is_domain"`
	Methods   []string  `json:"methods"`
	ExpiresAt string    `json:"expires_at"`
	Name      bool      `json:"name"`
	Catalogs  []Catalog `json:"catalog"`
	Project   Project
	User      User
}

type TokenCache struct {
	TokenExpireSecond int
	token             Token
	TokenId           string
	expiredAt         time.Time
}

func (tc *TokenCache) GetCatalogByType(serviceType string) *Catalog {
	for _, catalog := range tc.token.Catalogs {
		if catalog.Type == serviceType {
			return &catalog
		}
	}
	return nil
}

func (tc *TokenCache) GetEndpoints(option OptionCatalog) []Endpoint {
	endpoints := []Endpoint{}

	for _, catalog := range tc.token.Catalogs {
		if (option.Type != "" && catalog.Type != option.Type) ||
			(option.Name != "" && catalog.Name != option.Name) {
			continue
		}
		for _, endpoint := range catalog.Endpoints {
			if (option.Region != "" && endpoint.Region != option.Region) ||
				(option.Interface != "" && endpoint.Interface != option.Interface) {
				continue
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

type Region struct {
	Name           string `json:"name,omitempty"`
	Id             string `json:"id,omitempty"`
	Description    string `json:"description,omitempty"`
	Enabled        bool   `json:"enabled,omitempty"`
	ParentRegionId string `json:"parent_region_id,omitempty"`
}
