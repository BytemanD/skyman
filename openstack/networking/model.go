package networking

import (
	"encoding/json"

	"github.com/BytemanD/stackcrud/openstack/common"
)

type Router struct {
	common.Resource
	AdminStateUp        bool                   `json:"admin_state_up"`
	Distributed         bool                   `json:"distributed"`
	HA                  bool                   `json:"ha"`
	Routes              []string               `json:"routes"`
	RevsionNumber       int                    `json:"revision_number"`
	ExternalGatewayinfo map[string]interface{} `json:"external_gateway_info"`
	AvailabilityZones   string                 `json:"availability_zones"`
}

func (router Router) MarshalExternalGatewayInfo() string {
	jsonString, _ := common.GetIndentJson(router.ExternalGatewayinfo)
	return jsonString

}

type Network struct {
	common.Resource
	AdminStateUp      bool     `json:"admin_state_up"`
	Shared            bool     `json:"shared"`
	Subnets           []string `json:"subnets"`
	NetworkType       string   `json:"provider:network_type"`
	PhysicalNetwork   string   `json:"provider:physical_network"`
	AvailabilityZones string   `json:"availability_zones"`
	Mtu               int      `json:"mtu"`
}
type FixedIp struct {
	SubnetId  string `json:"subnet_id"`
	IpAddress string `json:"ip_address"`
}
type Port struct {
	common.Resource
	MACAddress      string                 `json:"mac_address"`
	BindingHostId   string                 `json:"binding:host_id,omitempty"`
	BindingVnicType string                 `json:"binding:vnic_type,omitempty"`
	BindingVifType  string                 `json:"binding:vif_type,omitempty"`
	BindingDetails  map[string]interface{} `json:"binding:vif_details,omitempty"`
	QosPolicyId     string                 `json:"qos_policy_id:host_id,omitempty"`
	FixedIps        []FixedIp              `json:"fixed_ips"`
	DeviceOwner     string                 `json:"device_owner"`
	DeviceId        string                 `json:"device_id"`
	SecurityGroups  []string               `json:"security_groups"`
	RevsionNumber   int                    `json:"revision_number"`
}

func (port Port) MarshalVifDetails() string {
	bytes, _ := json.Marshal(port.BindingDetails)
	return string(bytes)

}

type Routers []Router
type Networks []Network
type Ports []Port
