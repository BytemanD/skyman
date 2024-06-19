package neutron

import (
	"encoding/json"
	"fmt"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/openstack/model"
)

type Router struct {
	model.Resource
	AdminStateUp          bool                   `json:"admin_state_up,omitempty"`
	Distributed           bool                   `json:"distributed,omitempty"`
	HA                    bool                   `json:"ha,omitempty"`
	Routes                []string               `json:"routes,omitempty"`
	RevsionNumber         int                    `json:"revision_number,omitempty"`
	ExternalGatewayInfo   map[string]interface{} `json:"external_gateway_info,omitempty"`
	AvailabilityZones     []string               `json:"availability_zones,omitempty"`
	AvailabilityZoneHints []string               `json:"availability_zone_hints,omitempty"`
	Tags                  []string               `json:"tags,omitempty"`
}

func (router Router) MarshalExternalGatewayInfo() string {
	jsonString, _ := stringutils.JsonDumpsIndent(router.ExternalGatewayInfo)
	return jsonString

}

type Network struct {
	model.Resource
	AdminStateUp            bool     `json:"admin_state_up"`
	Shared                  bool     `json:"shared"`
	Subnets                 []string `json:"subnets"`
	AvailabilityZones       []string `json:"availability_zones"`
	AvailabilityZoneHints   []string `json:"availability_zone_hints,omitempty"`
	Mtu                     int      `json:"mtu"`
	Tags                    []string `json:"tags,omitempty"`
	RouterExternal          bool     `json:"router:external,omitempty"`
	ProviderSegmentation    int      `json:"provider:segmentation_id,omitempty"`
	ProviderNetworkType     string   `json:"provider:network_type,omitempty"`
	ProviderPhysicalNetwork string   `json:"provider:physical_network,omitempty"`
	QosPolicyId             string   `json:"qos_policy_id,omitempty"`
	IsDefault               bool     `json:"is_default,omitempty"`
	PortSecurityEnabled     bool     `json:"port_security_enabled,omitempty"`
}
type AllocationPool struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}
type Subnet struct {
	model.Resource
	NetworkId       string           `json:"network_id,omitempty"`
	Cidr            string           `json:"cidr,omitempty"`
	HostRouters     []string         `json:"host_routes,omitempty"`
	RevisionNumber  int              `json:"revision_number,omitempty"`
	IpVersion       int              `json:"ip_version,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	EnableDhcp      bool             `json:"enable_dhcp,omitempty"`
	GatewayIp       string           `json:"gateway_ip,omitempty"`
	AllocationPools []AllocationPool `json:"allocation_pools,omitempty"`
}

func (subnet Subnet) GetAllocationPoolsList() []string {
	pools := []string{}
	for _, pool := range subnet.AllocationPools {
		pools = append(pools, fmt.Sprintf("%s-%s", pool.Start, pool.End))
	}
	return pools
}

type FixedIp struct {
	SubnetId  string `json:"subnet_id"`
	IpAddress string `json:"ip_address"`
}
type Port struct {
	model.Resource
	MACAddress      string                 `json:"mac_address"`
	BindingHostId   string                 `json:"binding:host_id,omitempty"`
	BindingVnicType string                 `json:"binding:vnic_type,omitempty"`
	BindingVifType  string                 `json:"binding:vif_type,omitempty"`
	BindingDetails  map[string]interface{} `json:"binding:vif_details,omitempty"`
	BindingProfile  map[string]interface{} `json:"binding:profile,omitempty"`
	QosPolicyId     string                 `json:"qos_policy_id:host_id,omitempty"`
	FixedIps        []FixedIp              `json:"fixed_ips"`
	DeviceOwner     string                 `json:"device_owner"`
	DeviceId        string                 `json:"device_id"`
	SecurityGroups  []string               `json:"security_groups"`
	RevsionNumber   int                    `json:"revision_number"`
	TenantId        string                 `json:"tenant_id,omitempty"`
}
type Agent struct {
	model.Resource
	Binary           string `json:"binary"`
	Host             string `json:"host,omitempty"`
	Topic            string `json:"topic,omitempty"`
	AgentType        string `json:"agent_type,omitempty"`
	AvailabilityZone string `json:"availability_zone"`
	Alive            bool   `json:"alive,omitempty"`
	AdminStateUp     bool   `json:"admin_state_up,omitempty"`
}
type SecurityGroup struct {
	model.Resource
}

func (port Port) MarshalVifDetails() string {
	bytes, _ := json.Marshal(port.BindingDetails)
	return string(bytes)
}
func (port Port) MarshalBindingProfile() string {
	bytes, _ := json.Marshal(port.BindingProfile)
	return string(bytes)
}
func (port Port) VifDetailList() []string {
	details := []string{}
	for k, v := range port.BindingDetails {
		details = append(details, fmt.Sprintf("%s=%v", k, v))
	}
	return details
}

func (port Port) IsActive() bool {
	return port.Status == "ACTIVE"
}

func (port Port) IsUnbound() bool {
	return port.BindingVifType == "unbound"
}
func (port Port) GetFixedIpaddress() []string {
	address := []string{}
	for _, fixedIp := range port.FixedIps {
		address = append(address, fixedIp.IpAddress)
	}
	return address
}

type Routers []Router
type Networks []Network
type Ports []Port
type SecurityGroups []SecurityGroup
