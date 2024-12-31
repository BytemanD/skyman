package keystone

import (
	"github.com/BytemanD/skyman/openstack/model"
)

type Service struct {
	model.Resource
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
}

type Endpoint struct {
	Id        string `json:"id,omitempty"`
	Region    string `json:"region,omitempty"`
	Url       string `json:"url,omitempty"`
	Interface string `json:"interface,omitempty"`
	RegionId  string `json:"region_id,omitempty"`
	Enabled   bool   `json:"enabled"`
	ServiceId string `json:"service_id,omitempty"`
}

type Region struct {
	Id             string      `json:"id,omitempty"`
	ParentRegionId string      `json:"parent_region_id,omitempty"`
	Description    string      `json:"description,omitempty"`
	Links          interface{} `json:"links,omitempty"`
}
type Scope struct {
	Project model.Project `json:"project,omitempty"`
}
type RoleAssigment struct {
	Scope Scope      `json:"scope,omitempty"`
	User  model.User `json:"user,omitempty"`
}

func (service Service) NameOrId() string {
	if service.Name != "" {
		return service.Name
	}
	if service.Type != "" {
		return service.Type
	}
	return service.Id
}
