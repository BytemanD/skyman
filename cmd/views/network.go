package views

import (
	"strings"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/neutron"
)

func PrintNetwork(network neutron.Network) {
	pt := common.PrettyItemTable{
		Item: network,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "AdminStateUp"},
			{Name: "AvailabilityZoneHints"},
			{Name: "AvailabilityZones"},
			{Name: "Mtu"},
			{Name: "Shared"},
			{Name: "Status"},
			{Name: "Tags"},
			{Name: "QosPolicyId"},
			{Name: "PortSecurityEnabled"},
			{Name: "RouterExternal"},
			{Name: "ProviderPhysicalNetwork"},
			{Name: "ProviderNetworkType"},
			{Name: "ProviderSegmentation"},
			{Name: "ProjectId"},
			{Name: "UpdatedAt"}, {Name: "CreatedAt"},
		},
	}
	common.PrintPrettyItemTable(pt)
}
func PrintSubnet(network neutron.Subnet) {
	pt := common.PrettyItemTable{
		Item: network,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "NetworkId"},
			{Name: "Cidr"},
			{Name: "IpVersion"},
			{Name: "EnableDhcp"},
			{Name: "AllocationPools", Slot: func(item interface{}) interface{} {
				p, _ := item.(neutron.Subnet)
				return strings.Join(p.GetAllocationPoolsList(), ",")
			}},
			{Name: "GatewayIp"},
			{Name: "RevisionNumber"},
			{Name: "HostRouters"},
			{Name: "Tags"},

			{Name: "ProjectId"},
			{Name: "UpdatedAt"}, {Name: "CreatedAt"},
		},
	}
	common.PrintPrettyItemTable(pt)
}
func PrintRouter(router neutron.Router) {
	pt := common.PrettyItemTable{
		Item: router,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "AdminStateUp"},
			{Name: "AvailabilityZoneHints"},
			{Name: "AvailabilityZones"},
			{Name: "Distributed"},
			{Name: "ExternalGatewayInfo"},
			{Name: "HA", Text: "Ha"},
			{Name: "Status"},
			{Name: "Tags"},
			{Name: "ProjectId"},
			{Name: "UpdatedAt"},
			{Name: "CreatedAt"},
		},
	}
	common.PrintPrettyItemTable(pt)
}
