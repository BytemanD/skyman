package compute

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/BytemanD/stackcrud/openstack/common"
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
		{"Id", server.Id}, {"name", server.Name},
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

type Service struct {
	common.Resource
	Zone           string `json:"zone,omitempty"`
	Host           string `json:"host,omitempty"`
	Binary         string `json:"binary,omitempty"`
	Status         string `json:"status,omitempty"`
	State          string `json:"state,omitempty"`
	DisabledReason string `json:"disabled_reason,omitempty"`
	ForcedDown     bool   `json:"forced_down,omitempty"`
}

type Services []Service
type ServiceBody struct {
	Service Service `json:"service"`
}
type ServicesBody struct {
	Services Services `json:"services"`
}

type Flavors []Flavor
type FlavorBody struct {
	Flavor Flavor `json:"flavor"`
}
type FlavorsBody struct {
	Flavors Flavors `json:"flavors"`
}

type HypervisorServer struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}
type Hypervisor struct {
	common.Resource
	Hostname     string `json:"hypervisor_hostname"`
	HostIp       string `json:"host_ip"`
	Status       string `json:"status"`
	State        string `json:"state"`
	Type         string `json:"hypervisor_type"`
	Version      int    `json:"hypervisor_version"`
	Uptime       string `json:"uptime"`
	Vcpus        int    `json:"vcpus"`
	VcpusUsed    int    `json:"vcpus_used"`
	MemoryMB     int    `json:"memory_mb"`
	MemoryMBUsed int    `json:"memory_mb_used"`
	Servers      []HypervisorServer
}
type Hypervisors []Hypervisor

type keypair struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
}

type Keypair struct {
	Keypair keypair `json:"keypair"`
}
type Keypairs []Keypair

type InstanceAction struct {
	common.Resource
	Action       string                `json:"action"`
	InstanceUUID string                `json:"instance_uuid"`
	Message      string                `json:"message"`
	RequestId    string                `json:"request_id"`
	StartTime    string                `json:"start_time"`
	Events       []InstanceActionEvent `json:"events"`
}
type InstanceActionEvent struct {
	Event      string `json:"event"`
	StartTime  string `json:"start_time"`
	FinishTime string `json:"finish_time"`
	Host       string `json:"host"`
	HostId     string `json:"hostId"`
	Result     string `json:"result"`
	Traceback  string `json:"traceback"`
}
type InstanceActions []InstanceAction
type InstanceActionEvents []InstanceActionEvent

type VolumeAttachment struct {
	Id                  string `json:"id"`
	Device              string `json:"device"`
	ServerId            string `json:"serverId"`
	VolumeId            string `json:"volumeId"`
	DeleteOnTermination bool   `json:"delete_on_termination,omitempty"`
}
type InterfaceAttachment struct {
	MacAddr   string    `json:"mac_addr"`
	NetId     string    `json:"net_id"`
	PortId    string    `json:"port_id"`
	PortState string    `json:"port_state"`
	FixedIps  []FixedIp `json:"fixed_ips"`
}
type FixedIp struct {
	IpAddress string `json:"ip_address"`
	SubnetId  string `json:"subnet_id"`
}

func (attachment InterfaceAttachment) GetIPAddresses() []string {
	addresses := []string{}
	for _, fixedIp := range attachment.FixedIps {
		addresses = append(addresses, fixedIp.IpAddress)
	}
	return addresses
}

type ConsoleLog struct {
	Output string `json:"output"`
}
type Console struct {
	Type     string `json:"type"`
	Url      string `json:"url"`
	Protocol string `json:"protocol"`
}
