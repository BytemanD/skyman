package identity

import "github.com/BytemanD/skyman/openstack/common"

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
