package keystone

import (
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/openstack/model"
)

type Service struct {
	model.Resource
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

type Region struct {
	Id             string      `json:"id,omitempty"`
	ParentRegionId string      `json:"parent_region_id,omitempty"`
	Description    string      `json:"description,omitempty"`
	Links          interface{} `json:"links,omitempty"`
}
type Scope struct {
	Project auth.Project `json:"project,omitempty"`
}
type RoleAssigment struct {
	Scope Scope     `json:"scope,omitempty"`
	User  auth.User `json:"user,omitempty"`
}
