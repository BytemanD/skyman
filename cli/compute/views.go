package compute

import (
	"fmt"
	"os"
	"strings"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/openstack/compute"
	"github.com/jedib0t/go-pretty/v6/list"
)

func printServer(server compute.Server) {
	dataTable := cli.DataTable{
		Item: server,
		ShortFields: []cli.Field{
			{Name: "Id"}, {Name: "Id"}, {Name: "Description"},
			{Name: "Flavor:original_name"}, {Name: "Flavor:ram"},
			{Name: "Flavor:vcpus"}, {Name: "Flavor:disk"},
			{Name: "Flavor:swap"}, {Name: "Flavor:extra_specs"},
			{Name: "Image"},
			{Name: "AZ"}, {Name: "Host"},
			{Name: "Status"}, {Name: "TaskState"}, {Name: "PowerState"},
			{Name: "RootBdmType"},
			{Name: "Created"}, {Name: "Updated"},
			{Name: "UserId"}, {Name: "LaunchedAt"},

			{Name: "Fault:code"}, {Name: "Fault:message"},
			{Name: "Fault:details"},
		},
		Slots: map[string]func(item interface{}) interface{}{
			"Flavor:original_name": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.Flavor.OriginalName
			},
			"Flavor:ram": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.Flavor.Ram
			},
			"Flavor:vcpus": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.Flavor.Vcpus
			},
			"Flavor:disk": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.Flavor.Disk
			},
			"Flavor:swap": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.Flavor.Swap
			},
			"Flavor:extra_specs": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.GetFlavorExtraSpecsString()
			},
			"Image": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.Image.Id
			},
			"Fault:code": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.Fault.Code
			},
			"Fault:message": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.Fault.Message
			},
			"Fault:details": func(item interface{}) interface{} {
				p, _ := item.(compute.Server)
				return p.Fault.Details
			},
		},
	}
	dataTable.Print(false)
}
func printFlavor(server compute.Flavor) {
	dataTable := cli.DataTable{
		Item: server,
		ShortFields: []cli.Field{
			{Name: "Id"}, {Name: "Name"},
			{Name: "Vcpus"}, {Name: "Ram"}, {Name: "Disk"}, {Name: "Swap"},
			{Name: "RXTXFactor"},
			{Name: "OS-FLV-EXT-DATA:ephemeral"},
			{Name: "os-flavor-access:is_public"},
			{Name: "OS-FLV-DISABLED:disabled"},
			{Name: "ExtraSpecs"},
		},
		Slots: map[string]func(item interface{}) interface{}{
			"os-flavor-access:is_public": func(item interface{}) interface{} {
				p, _ := item.(compute.Flavor)
				return p.IsPublic
			},
			"OS-FLV-EXT-DATA:ephemeral": func(item interface{}) interface{} {
				p, _ := item.(compute.Flavor)
				return p.Ephemeral
			},
			"OS-FLV-DISABLED:disabled": func(item interface{}) interface{} {
				p, _ := item.(compute.Flavor)
				return p.Disabled
			},
			"ExtraSpecs": func(item interface{}) interface{} {
				p, _ := item.(compute.Flavor)
				return strings.Join(p.ExtraSpecs.GetList(), "\n")
			},
		},
	}
	dataTable.Print(false)
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

func printAZInfo(azList []compute.AvailabilityZone) {
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
					azHost.ZoneState = cli.BaseColorFormatter.Format("available")
				} else {
					azHost.ZoneState = cli.BaseColorFormatter.Format("disabled")
				}
				if service.Active {
					azHost.ServiceStatus = cli.BaseColorFormatter.Format("enabled")
				} else {
					azHost.ServiceStatus = cli.BaseColorFormatter.Format("disabled")
				}
				if service.Available {
					azHost.ServiceAvailable = cli.BaseColorFormatter.Format(":)")
				} else {
					azHost.ServiceAvailable = cli.BaseColorFormatter.Format("XXX")
				}
				azHostList = append(azHostList, azHost)
			}
		}
	}

	table := cli.DataListTable{
		ShortHeaders: []string{"ZoneName", "ZoneState", "HostName", "ServiceName",
			"ServiceStatus", "ServiceAvailable", "ServiceUpdatedAt"},
		HeaderLabel: map[string]string{"ServiceUpdatedAt": "Updated At"},
		Slots:       map[string]func(item interface{}) interface{}{},
	}
	table.AddItems(azHostList)
	table.Print(false)
}
func printAZInfoTree(azList []compute.AvailabilityZone) {
	tw := list.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(list.StyleConnectedRounded)

	for _, az := range azList {
		var zoneState string
		if az.ZoneState.Available {
			zoneState = cli.BaseColorFormatter.Format("available")
		} else {
			zoneState = cli.BaseColorFormatter.Format("disabled")
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
					serviceStatus = cli.BaseColorFormatter.Format("enabled")
				} else {
					serviceStatus = cli.BaseColorFormatter.Format("disabled")
				}
				if service.Available {
					serviceAvailable = cli.BaseColorFormatter.Format(":)")
				} else {
					serviceAvailable = cli.BaseColorFormatter.Format("XXX")
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
