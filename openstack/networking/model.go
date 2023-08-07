package networking

import (
	"github.com/BytemanD/stackcrud/openstack/common"
)

type Router struct {
	common.Resource
	AdminStateUp        bool     `json:"admin_state_up"`
	Distributed         bool     `json:"distributed"`
	HA                  bool     `json:"ha"`
	Routes              []string `json:"routes"`
	ExternalGatewayinfo string   `json:"external_gateway_info"`
}
type Network struct {
	common.Resource
	AdminStateUp      bool     `json:"admin_state_up"`
	Shared            bool     `json:"shared"`
	Subnets           []string `json:"subnets"`
	NetworkType       string   `json:"provider:network_type"`
	AvailabilityZones string   `json:"availability_zones"`
}
type FixedIp struct {
	SubnetId  string `json:"subnet_id"`
	IpAddress string `json:"ip_address"`
}
type Port struct {
	common.Resource
	MACAddress     string    `json:"mac_address"`
	FixedIps       []FixedIp `json:"fixed_ips"`
	DeviceOwner    string    `json:"device_owner"`
	SecurityGroups []string  `json:"security_groups"`
}

type Routers []Router
type Networks []Network
type Ports []Port
