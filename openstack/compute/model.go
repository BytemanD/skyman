package compute

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/BytemanD/stackcrud/openstack/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Flavor struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	OriginalName string            `json:"original_name"`
	Ram          int               `json:"ram"`
	Vcpus        int               `json:"vcpus"`
	Disk         int               `json:"disk"`
	Swap         int               `json:"swap"`
	RXTXFactor   float32           `json:"rxtx_factor"`
	ExtraSpecs   ExtraSpecs        `json:"extra_specs"`
	IsPublic     bool              `json:"os-flavor-access:is_public"`
	Ephemeral    int               `json:"OS-FLV-DISABLED:ephemeral"`
	Disabled     map[string]string `json:"OS-FLV-DISABLED:disabled"`
}
type ExtraSpecs map[string]string

type ExtraSpecsBody struct {
	ExtraSpecs ExtraSpecs `json:"extra_specs"`
}

func (extraSpecs ExtraSpecs) GetList() []string {
	properties := []string{}
	for k, v := range extraSpecs {
		properties = append(properties, fmt.Sprintf("%s=%s", k, v))
	}
	return properties
}

type Fault struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Details string `json:"details"`
}

type Server struct {
	common.Resource
	Status       string               `json:"status"`
	TaskState    string               `json:"OS-EXT-STS:task_state"`
	PowerState   int                  `json:"OS-EXT-STS:power_state"`
	VmState      string               `json:"OS-EXT-STS:vm_state"`
	Host         string               `json:"OS-EXT-SRV-ATTR:host"`
	AZ           string               `json:"OS-EXT-AZ:availability_zone"`
	Flavor       Flavor               `json:"flavor"`
	Image        Image                `json:"image"`
	Fault        Fault                `json:"fault"`
	Addresses    map[string][]Address `json:"addresses"`
	InstanceName string               `json:"OS-EXT-SRV-ATTR:instance_name"`
	ConfigDriver string               `json:"config_drive"`
	Created      string               `json:"created"`
	Updated      string               `json:"updated"`
	TerminatedAt string               `json:"OS-SRV-USG:terminated_at"`
	LaunchedAt   string               `json:"OS-SRV-USG:launched_at"`
	UserId       string               `json:"user_id"`
	Description  string               `json:"description"`
	RootBdmType  string               `json:"root_bdm_type"`
}
type Image struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type Address struct {
	MacAddr string `json:"OS-EXT-IPS-MAC:mac_addr"`
	Version int    `json:"version"`
	Addr    string `json:"addr"`
	Type    string `json:"OS-EXT-IPS:type"`
}
type ServerBody struct {
	Server *Server `json:"server"`
}
type Servers []Server
type ServersBody struct {
	Servers Servers `json:"servers"`
}

type ServeCreaterBody struct {
	Server ServerOpt `json:"server"`
}

func (server *Server) GetPowerState() string {
	return POWER_STATE[server.PowerState]
}
func (server *Server) GetTaskState() string {
	if server.TaskState == "" {
		return "-"
	}
	return server.TaskState
}
func (server *Server) GetNetworks() []string {
	var networks []string
	for net, addresses := range server.Addresses {
		for _, address := range addresses {
			networks = append(networks, fmt.Sprintf("%s=%s", net, address.Addr))
		}
	}
	return networks
}
func (server Server) GetFlavorExtraSpecsString() string {
	var extraList []string
	for key, value := range server.Flavor.ExtraSpecs {
		extraList = append(extraList, key+"="+value)
	}
	sort.Sort(sort.StringSlice(extraList))
	return strings.Join(extraList, "\n")
}
func (server Server) GetFaultString() string {
	fault, _ := json.Marshal(server.Fault)
	return string(fault)
}

func (server Server) Print() {
	header := table.Row{"Property", "Value"}

	tableWriter := table.NewWriter()
	tableWriter.AppendHeader(header)
	tableWriter.AppendRows([]table.Row{
		{"ID", server.Id}, {"name", server.Name},
		{"description", server.Description},

		{"flavor:original_name", server.Flavor.OriginalName},
		{"flavor:ram", server.Flavor.Ram},
		{"flavor:vcpus", server.Flavor.Vcpus},
		{"flavor:disk", server.Flavor.Disk},
		{"flavor:swap", server.Flavor.Swap},
		{"flavor:extra_specs", server.GetFlavorExtraSpecsString()},

		{"image", server.Image.Id},

		{"availability_zone  ", server.AZ}, {"host", server.Host},

		{"status", server.Status}, {"task_state", server.TaskState},
		{"power_state", server.PowerState}, {"vm_state", server.VmState},

		{"root_bdm_type", server.RootBdmType},

		{"created", server.Created}, {"updated", server.Updated},
		{"terminated_at", server.TerminatedAt}, {"launched_at", server.LaunchedAt},

		{"user_id", server.UserId},
		{"fault:code", server.Fault.Code},
		{"fault:message", server.Fault.Message},
		{"fault:details", server.Fault.Details},
	})
	// tableWriter.SetStyle(table.StyleLight)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()
}

func (servers Servers) Print(long bool, verbose bool) {
	header := table.Row{
		"ID", "Name", "Status", "Task State", "Power State", "Networks",
	}
	var networksJoinSep string
	if long {
		header = append(header, "AZ", "Host", "Instance Name", "Flavor:Name")
		if verbose {
			header = append(header, "Flavor:ram")
			header = append(header, "Flavor:vcpus")
			header = append(header, "Image")
		}
		networksJoinSep = "\n"
	} else {
		networksJoinSep = "; "
	}
	tableWriter := table.NewWriter()

	for _, server := range servers {
		row := table.Row{
			server.Id, server.Name, server.Status,
			server.GetTaskState(), server.GetPowerState(),
			strings.Join(server.GetNetworks(), networksJoinSep),
		}
		if long {
			row = append(row, server.AZ, server.Host, server.InstanceName, server.Flavor.OriginalName)
			if verbose {
				row = append(row, server.Flavor.Ram, server.Flavor.Vcpus)
				if server.Image.Name != "" {
					row = append(row, server.Image.Name)
				} else {
					row = append(row, server.Image.Id)
				}
			}
		}
		tableWriter.SortBy([]table.SortBy{
			{Name: "Name", Mode: table.Asc},
		})
		tableWriter.AppendRow(row)
	}

	// tableWriter.SetStyle(table.StyleLight)
	tableWriter.AppendHeader(header)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()
}

type Service struct {
	common.Resource
	Zone           string `json:"zone"`
	Host           string `json:"host"`
	Binary         string `json:"binary"`
	Status         string `json:"status"`
	State          string `json:"state"`
	DisabledReason string `json:"disabled_reason"`
	ForceDown      bool   `json:"forced_down"`
}

type Services []Service
type ServiceBody struct {
	Service Service `json:"service"`
}
type ServicesBody struct {
	Services Services `json:"services"`
}

func (services Services) PrintTable(long bool) {
	header := table.Row{
		"ID", "Binary", "Host", "Zone", "Status", "State", "Update At",
	}
	if long {
		header = append(header, "Disable Reason", "Forced Down")
	}
	tableWriter := table.NewWriter()

	for _, service := range services {
		row := table.Row{
			service.Id, service.Binary, service.Host, service.Zone,
			service.Status, service.State, service.UpdatedAt,
		}
		if long {
			row = append(row, service.DisabledReason, service.ForceDown)
		}
		tableWriter.AppendRow(row)
	}

	// tableWriter.SetStyle(table.StyleLight)
	tableWriter.AppendHeader(header)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()
}

type Flavors []Flavor
type FlavorBody struct {
	Flavor Flavor `json:"flavor"`
}
type FlavorsBody struct {
	Flavors Flavors `json:"flavors"`
}

func (flavors Flavors) PrintTable(long bool) {
	header := table.Row{
		"ID", "Name", "VCPUs", "RAM", "Disk", "Ephemeral", "Is Public",
	}
	if long {
		header = append(header, "Swap", "RXTX Factor", "Properties")
	}
	tableWriter := table.NewWriter()

	for _, flavor := range flavors {
		row := table.Row{
			flavor.Id, flavor.Name, flavor.Vcpus, flavor.Ram,
			flavor.Disk, flavor.Ephemeral, flavor.IsPublic,
		}
		if long {

			row = append(row, flavor.Swap, flavor.RXTXFactor,
				strings.Join(flavor.ExtraSpecs.GetList(), "\n"))
		}
		tableWriter.AppendRow(row)
	}
	tableWriter.SortBy([]table.SortBy{
		{Name: "Name", Mode: table.Asc},
	})
	// tableWriter.SetStyle(table.StyleLight)
	tableWriter.AppendHeader(header)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()
}
