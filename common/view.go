package common

import (
	"encoding/json"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/BytemanD/skyman/common/datatable"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

func PrintNetworks(items []neutron.Network, long bool) {
	PrintItems(
		[]datatable.Column[neutron.Network]{
			{Name: "Id"}, {Name: "Name"},
			{Name: "Status", AutoColor: true},
			{Name: "AdminStateUp", AutoColor: true},
			{Name: "Subnets", RenderFunc: func(item neutron.Network) interface{} {
				return strings.Join(item.Subnets, "\n")
			}},
		},
		[]datatable.Column[neutron.Network]{
			{Name: "Shared"}, {Name: "ProviderNetworkType"},
			{Name: "AvailabilityZones"},
		},
		items, TableOptions{
			SortBy:       []table.SortBy{{Name: "Name"}},
			SeparateRows: long,
			More:         long},
	)
}
func PrintNetwork(item neutron.Network) {
	PrintItem(
		[]datatable.Field[neutron.Network]{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "AdminStateUp"}, {Name: "AvailabilityZoneHints"},
			{Name: "AvailabilityZones"}, {Name: "Mtu"},
			{Name: "Shared"}, {Name: "Status"}, {Name: "Tags"},
			{Name: "QosPolicyId"}, {Name: "PortSecurityEnabled"},
			{Name: "RouterExternal"}, {Name: "ProviderPhysicalNetwork"},
			{Name: "ProviderNetworkType"}, {Name: "ProviderSegmentation"},
			{Name: "ProjectId"},
			{Name: "UpdatedAt"}, {Name: "CreatedAt"},
		},
		[]datatable.Field[neutron.Network]{},
		item, TableOptions{},
	)
}
func PrintSubnets(items []neutron.Subnet, long bool) {
	PrintItems(
		[]datatable.Column[neutron.Subnet]{
			{Name: "Id"}, {Name: "Name"},
			{Name: "NetworkId"}, {Name: "Cidr"},
		},
		[]datatable.Column[neutron.Subnet]{
			{Name: "EnableDhcp", Text: "Dhcp"},
			{Name: "AllocationPools", RenderFunc: func(item neutron.Subnet) interface{} {
				return strings.Join(item.GetAllocationPoolsList(), ",")
			}},
			{Name: "IpVersion"},
			{Name: "GatewayIp"},
		},
		items, TableOptions{
			More:   long,
			SortBy: []table.SortBy{{Name: "Name"}},
		},
	)
}
func PrintSubnet(item neutron.Subnet) {
	PrintItem(
		[]datatable.Field[neutron.Subnet]{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "NetworkId"}, {Name: "Cidr"}, {Name: "IpVersion"},
			{Name: "EnableDhcp"},
			{Name: "AllocationPools", RenderFunc: func(item neutron.Subnet) interface{} {
				return strings.Join(item.GetAllocationPoolsList(), ",")
			}},
			{Name: "GatewayIp"}, {Name: "RevisionNumber"}, {Name: "HostRouters"},
			{Name: "Tags"},
			{Name: "ProjectId"},
			{Name: "UpdatedAt"}, {Name: "CreatedAt"},
		},
		[]datatable.Field[neutron.Subnet]{},
		item, TableOptions{},
	)
}
func PrintRouters(items []neutron.Router, long bool) {
	PrintItems(
		[]datatable.Column[neutron.Router]{
			{Name: "Id"}, {Name: "Name"},
			{Name: "Status", AutoColor: true},
			{Name: "AdminStateUp", AutoColor: true}, {Name: "Distributed"},
			{Name: "HA", Text: "HA"},
		},
		[]datatable.Column[neutron.Router]{
			{Name: "Routes"}, {Name: "ExternalGatewayinfo"},
		},
		items, TableOptions{
			SortBy: []table.SortBy{{Name: "Name"}},
			More:   long,
		},
	)
}
func PrintRouter(item neutron.Router) {
	PrintItem(
		[]datatable.Field[neutron.Router]{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "AdminStateUp"},
			{Name: "AvailabilityZoneHints"},
			{Name: "AvailabilityZones"},
			{Name: "Distributed"},
			{Name: "ExternalGatewayInfo", RenderFunc: func(item neutron.Router) interface{} {
				return item.MarshalExternalGatewayInfo()
			}},
			{Name: "HA", Text: "Ha"},
			{Name: "Status"},
			{Name: "Tags"},
			{Name: "ProjectId"},
			{Name: "UpdatedAt"},
			{Name: "CreatedAt"},
		},
		[]datatable.Field[neutron.Router]{},
		item, TableOptions{},
	)
}
func PrintPorts(items []neutron.Port, long bool) {
	PrintItems(
		[]datatable.Column[neutron.Port]{
			{Name: "Id"}, {Name: "Name"},
			{Name: "Status", AutoColor: true},
			{Name: "BindingVnicType", Text: "VnicType"},
			{Name: "BindingVifType", Text: "VifType"},
			{Name: "MACAddress", Text: "MAC Address"},
			{Name: "FixedIps", RenderFunc: func(item neutron.Port) interface{} {
				ips := []string{}
				if !long {
					for _, fixedIp := range item.FixedIps {
						ips = append(ips, fixedIp.IpAddress)
					}
					return strings.Join(ips, ", ")
				} else {
					data, _ := json.Marshal(item.FixedIps)
					return string(data)
				}
			}},
			{Name: "DeviceOwner"},
			{Name: "BindingHostId"},
		},
		[]datatable.Column[neutron.Port]{
			{Name: "DeviceId"},
			{Name: "TenantId"},
			{Name: "BindingProfile"},
			{Name: "SecurityGroups", RenderFunc: func(item neutron.Port) interface{} {
				return strings.Join(item.SecurityGroups, "\n")
			}},
		},
		items, TableOptions{
			SortBy:       []table.SortBy{{Name: "Name"}},
			SeparateRows: long,
		},
	)
}
func PrintPort(item neutron.Port) {
	PrintItem(
		[]datatable.Field[neutron.Port]{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Status"},
			{Name: "AdminStateUp"},
			{Name: "MACAddress", Text: "MAC Address"},
			{Name: "BindingVnicType"},
			{Name: "BindingVifType"},
			{Name: "BindingProfile", RenderFunc: func(item neutron.Port) interface{} {
				return item.MarshalBindingProfile()
			}},
			{Name: "BindingDetails", RenderFunc: func(item neutron.Port) interface{} {
				return item.MarshalVifDetails()
			}},
			{Name: "BindingHostId"},
			{Name: "FixedIps"},
			{Name: "DeviceOwner"}, {Name: "DeviceId"},
			{Name: "QosPolicyId"}, {Name: "SecurityGroups"},
			{Name: "RevsionNumber"},
			{Name: "ProjectId"},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
		},
		[]datatable.Field[neutron.Port]{},
		item, TableOptions{},
	)
}

func PrintSecurityGroups(items []neutron.SecurityGroup, long bool) {
	PrintItems(
		[]datatable.Column[neutron.SecurityGroup]{
			{Name: "Id"}, {Name: "Name"},
			{Name: "ProjectId"},
			{Name: "RevisionNumber"},
			{Name: "Rules", RenderFunc: func(item neutron.SecurityGroup) interface{} {
				rules := []string{}
				for _, rule := range item.Rules {
					rules = append(rules, rule.String())
				}
				return strings.Join(rules, "\n")
			}},
		},
		[]datatable.Column[neutron.SecurityGroup]{
			// {Name: "Description"},
			{Name: "CreatedAt"},
			{Name: "UpdatedAt"},
		},
		items, TableOptions{More: long},
	)

}
func PrintSecurityGroup(item neutron.SecurityGroup) {
	PrintItem(
		[]datatable.Field[neutron.SecurityGroup]{
			{Name: "Id"}, {Name: "Name"},
			{Name: "Description"},
			{Name: "RevisionNumber"},
			{Name: "CreatedAt"},
			{Name: "UpdatedAt"},
			{Name: "Rules", RenderFunc: func(item neutron.SecurityGroup) interface{} {
				rules := []string{}
				for _, rule := range item.Rules {
					rules = append(rules, rule.String())
				}
				return strings.Join(rules, "\n")
			}},
			{Name: "ProjectId"},
		},
		[]datatable.Field[neutron.SecurityGroup]{},
		item, TableOptions{},
	)
}
func PrintSecurityGroupRules(items []neutron.SecurityGroupRule, long bool) {
	PrintItems(
		[]datatable.Column[neutron.SecurityGroupRule]{
			{Name: "Id"},
			{Name: "Protocol"},
			{Name: "RemoteIpPrefix"},
			{Name: "PortRange", RenderFunc: func(p neutron.SecurityGroupRule) interface{} {
				return p.PortRange()
			}},
			{Name: "RemoteGroupId"},
			{Name: "SecurityGroupId"},
		},
		[]datatable.Column[neutron.SecurityGroupRule]{
			// {Name: "Description"},
			{Name: "Direction"},
			{Name: "Ethertype"},
		},
		items, TableOptions{More: long},
	)

}
func PrintSecurityGroupRule(item neutron.SecurityGroupRule) {
	PrintItem(
		[]datatable.Field[neutron.SecurityGroupRule]{
			{Name: "Id"},
			{Name: "Protocol"},
			{Name: "Direction"},
			{Name: "Ethertype"},
			{Name: "RemoteIpPrefix"},
			{Name: "PortRange", RenderFunc: func(item neutron.SecurityGroupRule) interface{} {
				return item.PortRange()
			}},
			{Name: "RemoteGroupId"},
			{Name: "SecurityGroupId"},
			{Name: "RevisionNumber"},
			{Name: "CreatedAt"},
			{Name: "UpdatedAt"},
			{Name: "ProjectId"},
		},
		[]datatable.Field[neutron.SecurityGroupRule]{},
		item, TableOptions{},
	)
}
func PrintQosPolicys(items []neutron.QosPolicy, long bool) {
	PrintItems(
		[]datatable.Column[neutron.QosPolicy]{
			{Name: "Id"}, {Name: "Name"}, {Name: "Shared"}, {Name: "Default"},
			{Name: "ProjectId", Text: "Project"},
		},
		[]datatable.Column[neutron.QosPolicy]{},
		items, TableOptions{More: long},
	)

}
func PrintQosPolicy(item neutron.QosPolicy) {
	PrintItem(
		[]datatable.Field[neutron.QosPolicy]{
			{Name: "Id"}, {Name: "Name"},
			{Name: "Description"},
			{Name: "Shared"},
			{Name: "ProjectId", Text: "Project"},
			{Name: "Rules", RenderFunc: func(item neutron.QosPolicy) interface{} {
				bytes, _ := json.Marshal(item.Rules)
				return string(bytes)
			}},
		},
		[]datatable.Field[neutron.QosPolicy]{},
		item, TableOptions{},
	)
}
func PrintQosPolicyRules(items []neutron.QosRule, long bool) {
	PrintItems(
		[]datatable.Column[neutron.QosRule]{
			{Name: "Id"}, {Name: "QosPolicyId"},
			{Name: "Type"},
			{Name: "Direction"},
			{Name: "MaxKbps"},
			{Name: "MaxBurstKbps"},
			{Name: "MinKbps"},
		},
		[]datatable.Column[neutron.QosRule]{},
		items, TableOptions{More: long},
	)

}
func PrintQosPolicyRule(item neutron.QosRule) {
	PrintItem(
		[]datatable.Field[neutron.QosRule]{
			{Name: "Id"}, {Name: "Name"},
			{Name: "Description"},
			{Name: "Shared"},
			{Name: "ProjectId", Text: "Project"},
			{Name: "Rules", RenderFunc: func(item neutron.QosRule) interface{} {
				bytes, _ := json.Marshal(item)
				return string(bytes)
			}},
		},
		[]datatable.Field[neutron.QosRule]{},
		item, TableOptions{},
	)
}

func PrintAgents(items []neutron.Agent, long bool) {
	PrintItems(
		[]datatable.Column[neutron.Agent]{
			{Name: "Id"}, {Name: "AgentType"},
			{Name: "Host"},
			{Name: "AvailabilityZone", Align: text.AlignRight},
			{Name: "Alive", AutoColor: true, RenderFunc: func(item neutron.Agent) interface{} {
				return item.AliveEmoji()
			}},
			{Name: "AdminStateUp"},
			{Name: "Binary"},
		},
		[]datatable.Column[neutron.Agent]{},
		items, TableOptions{
			More: long},
	)
}

// glance

func PrintImages(items []glance.Image, long bool) {
	PrintItems(
		[]datatable.Column[glance.Image]{
			{Name: "Id"}, {Name: "Name"},
			{Name: "Status", AutoColor: true},
			{Name: "Size", Align: text.AlignRight,
				RenderFunc: func(item glance.Image) interface{} {
					return item.HumanSize()
				}},
			{Name: "DiskFormat"}, {Name: "ContainerFormat"},
		},
		[]datatable.Column[glance.Image]{
			{Name: "Visibility"}, {Name: "Protected"},
		},
		items, TableOptions{
			SortBy: []table.SortBy{{Name: "Name"}},
			More:   long,
		},
	)
}
func PrintImage(item glance.Image, human bool) {
	PrintItem(
		[]datatable.Field[glance.Image]{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Checksum"}, {Name: "Schema"},
			{Name: "DirectUrl"}, {Name: "Status"},
			{Name: "ContainerFormat"}, {Name: "DiskFormat"},
			{Name: "File"},
			{Name: "Size", RenderFunc: func(item glance.Image) interface{} {
				if human {
					return item.HumanSize()
				} else {
					return item.Size
				}
			}},
			{Name: "Properties", RenderFunc: func(item glance.Image) interface{} {
				return strings.Join(item.GetPropertyList(), "\n")
			}},
			{Name: "VirtualSize"}, {Name: "ProcessInfo"}, {Name: "Protected"},
			{Name: "Visibility"},
			{Name: "OSHashAlgo", Text: "OS Hash Algo"},
			{Name: "OSHashValue", Text: "OS Hash Value"},
			{Name: "Tags"}, {Name: "Owner"},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
		},
		[]datatable.Field[glance.Image]{},
		item, TableOptions{},
	)
}

// quota

func PrintQuotaSet(item nova.QuotaSet, more bool) {
	PrintItem(
		[]datatable.Field[nova.QuotaSet]{
			{Name: "Instances"},
			{Name: "Cores"}, {Name: "Ram"},
			{Name: "MetadataItems"},
			{Name: "SecurityGroups"},
			{Name: "SecurityGroupsMembers"},
			{Name: "InjectedFiles"},
			{Name: "InjectedFileContentBytes"},
			{Name: "InjectedFilePathBytes"},
		},
		[]datatable.Field[nova.QuotaSet]{
			{Name: "FloatingIps"},
			{Name: "FixedIps"},
		},
		item, TableOptions{More: more},
	)
}
