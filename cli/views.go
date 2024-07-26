package cli

import (
	"fmt"
	"strings"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

func PrintServer(server nova.Server) {
	pt := common.PrettyItemTable{
		Item: server,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Flavor:original_name", Text: "Flavor:original_name",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.OriginalName
				}},
			{Name: "Flavor:ram", Text: "Flavor:ram",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.Ram
				}},
			{Name: "Flavor:vcpus", Text: "Flavor:vcpus",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.Vcpus
				}},
			{Name: "Flavor:disk", Text: "Flavor:disk",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.Disk
				}},
			{Name: "Flavor:swap", Text: "Flavor:swap",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.Swap
				}},
			{Name: "Flavor:extra_specs", Text: "Flavor:extra_specs",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.GetFlavorExtraSpecsString()

				}},
			{Name: "Image",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					if p.ImageName() != "" {
						return fmt.Sprintf("%s (%s)", p.ImageName(), p.ImageId())
					} else {
						return p.ImageId()
					}
				}},
			{Name: "KeyName"},
			{Name: "SecurityGroups", Slot: func(item interface{}) interface{} {
				p, _ := item.(nova.Server)
				sgNames := []string{}
				for _, sg := range p.SecurityGroups {
					if stringutils.ContainsString(sgNames, sg.Name) {
						continue
					}
					sgNames = append(sgNames, sg.Name)
				}
				return strings.Join(sgNames, ", ")
			}},
			{Name: "AZ", Text: "AZ"}, {Name: "Host"},
			{Name: "Status"}, {Name: "TaskState"}, {Name: "PowerState"},
			{Name: "RootBdmType"},
			{Name: "Created"}, {Name: "LaunchedAt"}, {Name: "Updated"}, {Name: "TerminatedAt"},

			{Name: "Fault:code", Text: "Fault:code",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Fault.Code
				}},
			{Name: "Fault:message", Text: "Fault:message",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Fault.Message
				}},
			{Name: "Fault:details", Text: "Fault:details",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Fault.Details
				}},
			{Name: "UserId"},
		},
	}
	common.PrintPrettyItemTable(pt)
}

type AZHost struct {
	ZoneName         string
	ZoneState        string
	HostName         string
	ServiceName      string
	ServiceAvailable string
	ServiceStatus    string
	ServiceUpdatedAt string
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
