package neutron

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/samber/lo"
)

type Route struct {
	Nexthop       string `json:"nexthop,omitempty"`
	Destination   string `json:"destination,omitempty"`
	HostRouteType string `json:"host_route_type,omitempty"`
}
type Router struct {
	model.Resource
	AdminStateUp          bool           `json:"admin_state_up,omitempty"`
	Distributed           bool           `json:"distributed,omitempty"`
	HA                    bool           `json:"ha,omitempty"`
	Routes                []Route        `json:"routes,omitempty"`
	RevsionNumber         int            `json:"revision_number,omitempty"`
	ExternalGatewayInfo   map[string]any `json:"external_gateway_info,omitempty"`
	AvailabilityZones     []string       `json:"availability_zones,omitempty"`
	AvailabilityZoneHints []string       `json:"availability_zone_hints,omitempty"`
	Tags                  []string       `json:"tags,omitempty"`
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
type HostRouter struct {
	NextHop     string `json:"nexthop,omitempty"`
	Destination string `json:"destination,omitempty"`
}
type Subnet struct {
	model.Resource
	NetworkId       string           `json:"network_id,omitempty"`
	Cidr            string           `json:"cidr,omitempty"`
	HostRouters     []HostRouter     `json:"host_routes,omitempty"`
	RevisionNumber  int              `json:"revision_number,omitempty"`
	IpVersion       int              `json:"ip_version,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	EnableDhcp      bool             `json:"enable_dhcp,omitempty"`
	GatewayIp       string           `json:"gateway_ip,omitempty"`
	AllocationPools []AllocationPool `json:"allocation_pools,omitempty"`
}

func (subnet Subnet) GetAllocationPoolsList() []string {
	return lo.Map(subnet.AllocationPools, func(pool AllocationPool, _ int) string {
		return fmt.Sprintf("%s-%s", pool.Start, pool.End)
	})
}

type FixedIp struct {
	SubnetId  string `json:"subnet_id,omitempty"`
	IpAddress string `json:"ip_address,omitempty"`
}

func (fixedIp FixedIp) String() string {
	data, _ := json.Marshal(fixedIp)
	return string(data)
}

type Port struct {
	model.Resource
	AdminStateUp    bool           `json:"admin_state_up,omitempty"`
	MACAddress      string         `json:"mac_address"`
	BindingHostId   string         `json:"binding:host_id,omitempty"`
	BindingVnicType string         `json:"binding:vnic_type,omitempty"`
	BindingVifType  string         `json:"binding:vif_type,omitempty"`
	BindingDetails  map[string]any `json:"binding:vif_details,omitempty"`
	BindingProfile  map[string]any `json:"binding:profile,omitempty"`
	QosPolicyId     string         `json:"qos_policy_id,omitempty"`
	FixedIps        []FixedIp      `json:"fixed_ips"`
	DeviceOwner     string         `json:"device_owner"`
	DeviceId        string         `json:"device_id"`
	SecurityGroups  []string       `json:"security_groups"`
	RevsionNumber   int            `json:"revision_number"`
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

func (agent Agent) AliveEmoji() string {
	if agent.Alive {
		return ":-)"
	}
	return "XXX"
}

type SecurityGroup struct {
	model.Resource
	Tags           []string            `json:"tags,omitempty"`
	Default        bool                `json:"default,omitempty"`
	RevisionNumber int                 `json:"revision_number,omitempty"`
	Rules          []SecurityGroupRule `json:"security_group_rules,omitempty"`
}
type SecurityGroupRule struct {
	model.Resource
	SecurityGroupId string `json:"security_group_id,omitempty"`
	Direction       string `json:"direction,omitempty"`
	Ethertype       string `json:"ethertype,omitempty"`
	PortRangeMin    int    `json:"port_range_min,omitempty"`
	PortRangeMax    int    `json:"port_range_max,omitempty"`
	Protocol        string `json:"protocol,omitempty"`
	RemoteGroupId   string `json:"remote_group_id"`
	RemoteIpPrefix  string `json:"remote_ip_prefix,omitempty"`
	RevisionNumber  int    `json:"revision_number,omitempty"`
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
	return lo.MapToSlice(port.BindingDetails, func(k string, v any) string {
		return fmt.Sprintf("%s=%s", k, v)
	})
}

func (port Port) IsActive() bool {
	return port.Status == "ACTIVE"
}

func (port Port) IsUnbound() bool {
	return port.BindingVifType == "unbound"
}
func (port Port) GetFixedIpaddress() []string {
	address := lo.Map(port.FixedIps, func(fixedIp FixedIp, _ int) string {
		return fixedIp.IpAddress
	})
	return address
}
func (rule SecurityGroupRule) String() string {
	values := []string{
		fmt.Sprintf("Direction=%s", rule.Direction),
	}
	if rule.Ethertype != "" {
		values = append(values, fmt.Sprintf("Ethertype=%s", rule.Ethertype))
	}
	if rule.Protocol != "" {
		values = append(values, fmt.Sprintf("Protocol=%s", rule.Protocol))
	}
	if rule.RemoteGroupId != "" {
		values = append(values, fmt.Sprintf("RemoteGroupId=%s", rule.RemoteGroupId))
	}
	if rule.RemoteIpPrefix != "" {
		values = append(values, fmt.Sprintf("RemoteIpPrefix=%s", rule.RemoteIpPrefix))
	}
	if rule.PortRangeMin > 0 {
		values = append(values, fmt.Sprintf("PortRangeMin=%s", rule.RemoteGroupId))
	}
	if rule.PortRangeMax > 0 {
		values = append(values, fmt.Sprintf("PortRangeMax=%s", rule.RemoteGroupId))
	}
	return strings.Join(values, ",")
}
func (rule SecurityGroupRule) PortRange() string {
	portRange := ""
	if rule.PortRangeMin > 0 {
		portRange = fmt.Sprintf("%d", rule.PortRangeMin)
	}
	if rule.PortRangeMax > 0 {
		portRange += fmt.Sprintf("%s:%d", portRange, rule.PortRangeMax)
	}
	return portRange
}

type QosRule struct {
	model.Resource
	QosPolicyId  string `json:"qos_policy_id,omitempty"`
	Type         string `json:"type,omitempty"`
	Direction    string `json:"direction,omitempty"`
	MaxKbps      int    `json:"max_kbps,omitempty"`
	MinKbps      int    `json:"min_kbps,omitempty"`
	MaxBurstKbps int    `json:"max_burst_kbps,omitempty"`
}
type QosPolicy struct {
	model.Resource
	Shared  bool      `json:"shared,omitempty"`
	Default bool      `json:"default,omitempty"`
	Rules   []QosRule `json:"rules"`
}

type Routers []Router
type Networks []Network
type Ports []Port
type SecurityGroups []SecurityGroup
