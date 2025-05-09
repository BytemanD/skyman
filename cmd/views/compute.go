package views

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/jedib0t/go-pretty/v6/list"
)

func PrintServer(server nova.Server, client *openstack.Openstack) {
	pt := common.PrettyItemTable{
		Item: server,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Id"}, {Name: "InstanceName"},
			{Name: "Flavor:original_name", Text: "Flavor:original_name",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					return p.Flavor.OriginalName
				}},
			{Name: "Flavor:ram", Text: "Flavor:ram",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					return p.Flavor.Ram
				}},
			{Name: "Flavor:vcpus", Text: "Flavor:vcpus",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					return p.Flavor.Vcpus
				}},
			{Name: "Flavor:disk", Text: "Flavor:disk",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					return p.Flavor.Disk
				}},
			{Name: "Flavor:swap", Text: "Flavor:swap",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					return p.Flavor.Swap
				}},
			{Name: "Flavor:extra_specs", Text: "Flavor:extra_specs",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					return p.GetFlavorExtraSpecsString()

				}},
			{Name: "Image",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					if p.ImageName() != "" {
						return fmt.Sprintf("%s (%s)", p.ImageName(), p.ImageId())
					} else {
						return p.ImageId()
					}
				}},
			{Name: "AZ", Text: "AZ"}, {Name: "Host"}, {Name: "HypervisorHostname"},
			{Name: "Status"}, {Name: "TaskState"}, {Name: "PowerState"},
			{Name: "RootBdmType"}, {Name: "RootDeviceName"},
			{Name: "SecurityGroups", Slot: func(item any) any {
				p := item.(nova.Server)
				bytes, _ := json.Marshal(p.SecurityGroups)
				return string(bytes)
			}},
			{Name: "Progress"},
			{Name: "Created"}, {Name: "LaunchedAt"}, {Name: "Updated"}, {Name: "TerminatedAt"},

			{Name: "Fault:code", Text: "Fault:code",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					return p.Fault.Code
				}},
			{Name: "Fault:message", Text: "Fault:message",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					return p.Fault.Message
				}},
			{Name: "Fault:details", Text: "Fault:details",
				Slot: func(item any) any {
					p, _ := item.(nova.Server)
					return p.Fault.Details
				}},
			{Name: "UserId"},
			{Name: "TenantId", Text: "ProjectId", Slot: func(item any) any {
				p, _ := item.(nova.Server)
				if client != nil {
					project, err := client.KeystoneV3().GetProject(p.TenantId)
					if err != nil {
						console.Warn("get project %s failed: %s", p.TenantId, err)
					} else {
						return fmt.Sprintf("%s (%s)", project.Id, project.Name)
					}
				}
				return p.TenantId
			}},
		},
	}
	common.PrintPrettyItemTable(pt)
}

func PrintFlavor(server nova.Flavor) {
	pt := common.PrettyItemTable{
		Item: server,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"},
			{Name: "Vcpus"}, {Name: "Ram"}, {Name: "Disk"}, {Name: "Swap"},
			{Name: "RXTXFactor", Text: "RXTXFactor"},
			{Name: "OS-FLV-EXT-DATA:ephemeral", Text: "OS-FLV-EXT-DATA:ephemeral",
				Slot: func(item any) any {
					p, _ := item.(nova.Flavor)
					return p.Ephemeral
				}},
			{Name: "os-flavor-access:is_public", Text: "os-flavor-access:is_public",
				Slot: func(item any) any {
					p, _ := item.(nova.Flavor)
					return p.IsPublic
				}},
			{Name: "OS-FLV-DISABLED:disabled", Text: "OS-FLV-DISABLED:disabled",
				Slot: func(item any) any {
					p, _ := item.(nova.Flavor)
					return p.Disabled
				}},
			{Name: "ExtraSpecs", Slot: func(item any) any {
				p, _ := item.(nova.Flavor)
				extraSpecs := p.ExtraSpecs.GetList()
				sort.Strings(extraSpecs)
				return strings.Join(extraSpecs, "\n")
			}},
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

func PrintAZInfo(azList []nova.AvailabilityZone) {
	azHostList := []AZHost{}
	for _, az := range azList {
		for hostName, services := range az.Hosts {
			for serviceName, service := range services {
				azHost := AZHost{
					ZoneName:         az.ZoneName,
					HostName:         hostName,
					ServiceName:      serviceName,
					ServiceUpdatedAt: service.UpdatedAt,
				}
				if az.ZoneState.Available {
					azHost.ZoneState = "available"
				} else {
					azHost.ZoneState = "disabled"
				}
				if service.Active {
					azHost.ServiceStatus = "enabled"
				} else {
					azHost.ServiceStatus = "disabled"
				}
				if service.Available {
					azHost.ServiceAvailable = ":)"
				} else {
					azHost.ServiceAvailable = "XXX"
				}
				azHostList = append(azHostList, azHost)
			}
		}
	}

	pt := common.PrettyTable{
		ShortColumns: []common.Column{
			{Name: "ZoneName"}, {Name: "ZoneState", AutoColor: true}, {Name: "HostName"},
			{Name: "ServiceName"}, {Name: "ServiceStatus", AutoColor: true},
			{Name: "ServiceAvailable", AutoColor: true},
			{Name: "ServiceUpdatedAt", Text: "Updated At"},
		},
	}
	pt.AddItems(azHostList)
	common.PrintPrettyTable(pt, false)
}
func PrintAZInfoTree(azList []nova.AvailabilityZone) {
	tw := list.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(list.StyleConnectedRounded)

	for _, az := range azList {
		var zoneState string
		if az.ZoneState.Available {
			zoneState = utility.NewColorStatus("available").String()
		} else {
			zoneState = utility.NewColorStatus("disabled").String()
		}
		tw.AppendItem(fmt.Sprintf("%s %v", az.ZoneName, zoneState))
		tw.Indent()
		for hostName, services := range az.Hosts {
			tw.AppendItem(hostName)
			tw.Indent()
			for serviceName, service := range services {
				var (
					serviceStatus    string
					serviceAvailable string
				)
				if service.Active {
					serviceStatus = utility.NewColorStatus("enabled").String()
				} else {
					serviceStatus = utility.NewColorStatus("disabled").String()
				}
				if service.Available {
					serviceAvailable = utility.NewColorStatus(":)").String()
				} else {
					serviceAvailable = utility.NewColorStatus("XXX").String()
				}
				tw.AppendItem(
					fmt.Sprintf("%-20s %-10s %s", serviceName, serviceStatus, serviceAvailable),
				)
			}
			tw.UnIndent()
		}
		tw.UnIndent()
	}

	tw.Render()
}

func PrintAzInfoJson(azInfo []nova.AvailabilityZone) {
	jsonString, err := stringutils.JsonDumpsIndent(azInfo)
	if err != nil {
		console.Fatal("get json string failed, %v", err)
	}
	fmt.Println(jsonString)
}

func PrintAzInfoYaml(azInfo []nova.AvailabilityZone) {
	yamlString, err := common.GetYaml(azInfo)
	if err != nil {
		console.Fatal("get yaml string failed, %v", err)
	}
	fmt.Println(yamlString)
}

func PrintServiceTable(item any) {
	pt := common.PrettyItemTable{
		Item: item,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Binary"}, {Name: "Host"},
			{Name: "Status", AutoColor: true},
			{Name: "State", AutoColor: true},
			{Name: "ForcedDown", Text: "Forced Down"},
			{Name: "DisabledReason", Text: "Disabled Reason"},
		},
	}
	common.PrintPrettyItemTable(pt)
}
