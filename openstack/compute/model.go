package compute

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	markdown "github.com/MichaelMure/go-term-markdown"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/networking"
	"github.com/BytemanD/skyman/openstack/storage"
)

type Flavor struct {
	Id           string     `json:"id,omitempty"`
	Name         string     `json:"name,omitempty"`
	OriginalName string     `json:"original_name,omitempty"`
	Ram          int        `json:"ram,omitempty"`
	Vcpus        int        `json:"vcpus,omitempty"`
	Disk         int        `json:"disk"`
	Swap         int        `json:"swap,omitempty"`
	RXTXFactor   float32    `json:"rxtx_factor,omitempty"`
	ExtraSpecs   ExtraSpecs `json:"extra_specs,omitempty"`
	IsPublic     bool       `json:"os-flavor-access:is_public,omitempty"`
	Ephemeral    int        `json:"OS-FLV-DISABLED:ephemeral,omitempty"`
	Disabled     bool       `json:"OS-FLV-DISABLED:disabled,omitempty"`
}

func (flavor Flavor) Marshal() string {
	flavorMarshal, _ := json.Marshal(flavor)
	return string(flavorMarshal)
}

type ExtraSpecs map[string]string

func (extraSpecs ExtraSpecs) GetList() []string {
	properties := []string{}
	for k, v := range extraSpecs {
		properties = append(properties, fmt.Sprintf("%s=%s", k, v))
	}
	return properties
}

type Fault struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

func (fault *Fault) Marshal() string {
	faltMarshal, _ := json.Marshal(fault)
	return string(faltMarshal)
}

type Server struct {
	common.Resource
	TaskState    string               `json:"OS-EXT-STS:task_state,omitempty"`
	PowerState   int                  `json:"OS-EXT-STS:power_state,omitempty"`
	VmState      string               `json:"OS-EXT-STS:vm_state,omitempty"`
	Host         string               `json:"OS-EXT-SRV-ATTR:host,omitempty"`
	AZ           string               `json:"OS-EXT-AZ:availability_zone,omitempty"`
	Flavor       Flavor               `json:"flavor,omitempty"`
	Image        Image                `json:"image,omitempty"`
	Fault        Fault                `json:"fault,omitempty"`
	Addresses    map[string][]Address `json:"addresses,omitempty"`
	InstanceName string               `json:"OS-EXT-SRV-ATTR:instance_name,omitempty"`
	ConfigDriver string               `json:"config_drive,omitempty"`
	Created      string               `json:"created,omitempty"`
	Updated      string               `json:"updated,omitempty"`
	TerminatedAt string               `json:"OS-SRV-USG:terminated_at,omitempty"`
	LaunchedAt   string               `json:"OS-SRV-USG:launched_at,omitempty"`
	UserId       string               `json:"user_id,omitempty"`
	Description  string               `json:"description,omitempty"`
	RootBdmType  string               `json:"root_bdm_type,omitempty"`
}
type Image struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
type Address struct {
	MacAddr string `json:"OS-EXT-IPS-MAC:mac_addr"`
	Version int    `json:"version"`
	Addr    string `json:"addr"`
	Type    string `json:"OS-EXT-IPS:type"`
}

type Servers []Server

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
func (server Server) AllStatus() string {
	return fmt.Sprintf("status=%s, task_state=%s, power_state=%s",
		server.Status, server.TaskState, server.GetPowerState())
}
func (server Server) InResize() bool {
	return server.Status == "VERIFY_RESIZE" || server.Status == "RESIZE"
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

type Flavors []Flavor

type HypervisorServer struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}
type CpuInfo struct {
	Arch     string                 `json:"arch"`
	Model    string                 `json:"model"`
	Vendor   string                 `json:"vendor"`
	Features []string               `json:"Features"`
	Topology map[string]interface{} `json:"topology"`
}
type Hypervisor struct {
	common.Resource
	Hostname       string                 `json:"hypervisor_hostname"`
	HostIp         string                 `json:"host_ip"`
	Status         string                 `json:"status"`
	State          string                 `json:"state"`
	Type           string                 `json:"hypervisor_type"`
	Version        int                    `json:"hypervisor_version"`
	Uptime         string                 `json:"uptime"`
	Vcpus          int                    `json:"vcpus"`
	VcpusUsed      int                    `json:"vcpus_used"`
	MemoryMB       int                    `json:"memory_mb"`
	MemoryMBUsed   int                    `json:"memory_mb_used"`
	ExtraResources map[string]interface{} `json:"extra_resources"`
	CpuInfo        CpuInfo                `json:"cpu_info"`

	NumaNode0Hugepages map[string]interface{} `json:"numa_node_0_hugepages"`
	NumaNode1Hugepages map[string]interface{} `json:"numa_node_1_hugepages"`
	NumaNode2Hugepages map[string]interface{} `json:"numa_node_2_hugepages"`
	NumaNode3Hugepages map[string]interface{} `json:"numa_node_3_hugepages"`

	NumaNode0Cpuset map[string]interface{} `json:"numa_node_0_cpuset"`
	NumaNode1Cpuset map[string]interface{} `json:"numa_node_1_cpuset"`
	NumaNode2Cpuset map[string]interface{} `json:"numa_node_2_cpuset"`
	NumaNode3Cpuset map[string]interface{} `json:"numa_node_3_cpuset"`

	Servers []HypervisorServer
}

func (hypervisor Hypervisor) ExtraResourcesMarshal(indent bool) string {
	var m []byte
	if indent {
		m, _ = json.MarshalIndent(hypervisor.ExtraResources, "", "  ")
	} else {
		m, _ = json.Marshal(hypervisor.ExtraResources)
	}
	return string(m)
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

func (event InstanceActionEvent) GetSpendTime() (float64, error) {
	// event.StartTime
	if event.StartTime == "" {
		return 0, fmt.Errorf("No start time")
	}
	if event.FinishTime == "" {
		return 0, fmt.Errorf("No finish time")
	}
	logging.Debug("%s spend time: %s ~ %s", event.Event, event.StartTime, event.FinishTime)
	startTime, _ := time.Parse("2006-01-02T15:04:05", event.StartTime)
	finishTime, _ := time.Parse("2006-01-02T15:04:05", event.FinishTime)
	return float64(finishTime.Sub(startTime).Milliseconds()) / 1000, nil
}

type InstanceActionAndEvents struct {
	InstanceAction InstanceAction        `json:"action"`
	RequestId      string                `json:"request_id"`
	Events         []InstanceActionEvent `json:"events"`
}

func (actionWithEvents InstanceActionAndEvents) GetSpendTime() (float64, error) {
	// event.StartTime
	if actionWithEvents.InstanceAction.StartTime == "" {
		return 0, fmt.Errorf("No start time")
	}
	var lastEventFinishTime string
	for _, event := range actionWithEvents.Events {
		if event.FinishTime > lastEventFinishTime {
			lastEventFinishTime = event.FinishTime
		}
	}
	if lastEventFinishTime == "" {
		return 0, fmt.Errorf("No finish time")
	}

	logging.Debug("%s spend time: %s ~ %s", actionWithEvents.InstanceAction.Action,
		actionWithEvents.InstanceAction.StartTime, lastEventFinishTime)
	startTime, _ := time.Parse("2006-01-02T15:04:05", actionWithEvents.InstanceAction.StartTime)
	finishTime, _ := time.Parse("2006-01-02T15:04:05", lastEventFinishTime)
	return float64(finishTime.Sub(startTime).Milliseconds()) / 1000, nil
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

func (attachment InterfaceAttachment) GetIpAddresses() []string {
	addresses := []string{}
	for _, fixedIp := range attachment.FixedIps {
		addresses = append(addresses, fixedIp.IpAddress)
	}
	return addresses
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

type Migration struct {
	Id                int    `json:"id"`
	OldInstanceTypeId int    `json:"old_instance_type_id"`
	NewInstanceTypeId int    `json:"new_instance_type_id"`
	InstanceUUID      string `json:"instance_uuid"`
	MigrationType     string `json:"migration_type"`
	Status            string `json:"status"`
	DestCompute       string `json:"dest_compute"`
	DestNode          string `json:"dest_node"`
	DestHost          string `json:"dest_host"`
	SourceCompute     string `json:"source_compute"`
	SourceNode        string `json:"source_node"`
	SourceRegion      string `json:"source_region,omitempty"`
	DestRegion        string `json:"dest_regoin,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

type ZoneState struct {
	Available bool `json:"available"`
}
type ServiceState struct {
	Available bool   `json:"available"`
	Active    bool   `json:"active"`
	UpdatedAt string `json:"updated_at"`
}

type AZHost map[string]ServiceState

type AvailabilityZone struct {
	ZoneName  string            `json:"zoneName"`
	ZoneState ZoneState         `json:"zoneState"`
	Hosts     map[string]AZHost `json:"hosts,omitempty"`
}
type Aggregate struct {
	Id               int               `json:"id"`
	Name             string            `json:"name"`
	AvailabilityZone string            `json:"availability_zone"`
	Deleted          bool              `json:"deleted"`
	Hosts            []string          `json:"hosts,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	CreatedAt        string            `json:"created_at,omitempty"`
	UpdatedAt        string            `json:"updated_at,omitempty"`
	DeletedAt        string            `json:"deleted_at,omitempty"`
}

func (agg Aggregate) GetMetadataList() []string {
	metadataList := []string{}
	for k, v := range agg.Metadata {
		metadataList = append(metadataList, fmt.Sprintf("%s=%s", k, v))
	}
	return metadataList
}

func (agg Aggregate) MarshalMetadata() string {
	data, err := json.Marshal(agg.Metadata)
	if err != nil {
		logging.Warning("marshal metadata failed, %s", err)
		return ""
	}
	return string(data)
}

// {"policies": ["soft-anti-affinity"], "name": "soft-anti-affinity", "custom": false, "members": [], "id": "7357b2f7-d004-4c72-b38a-d3a156827c11", "metadata": {}}]
type ServerGroup struct {
	Id        string                 `json:"id"`
	Name      string                 `json:"name"`
	Policies  []string               `json:"policies"`
	Custom    bool                   `json:"custom"`
	Members   []string               `json:"members"`
	Metadata  map[string]interface{} `json:"metadata"`
	ProjectId string                 `json:"project_id"`
	UserId    string                 `json:"user_id"`
}

func (serverGroup ServerGroup) GetMetadataList() []string {
	metadataList := []string{}
	for k, v := range serverGroup.Metadata {
		metadataList = append(metadataList, fmt.Sprintf("%s=%s", k, v))
	}
	return metadataList
}

type RegionMigrateResp struct {
	AllowLiveMigrate bool   `json:"allow_live_migrate"`
	Reason           string `json:"reason"`
}

type ServerInspect struct {
	Server          Server                `json:"server"`
	Interfaces      []InterfaceAttachment `json:"interfaces"`
	Volumes         []VolumeAttachment    `json:"volumes"`
	PowerState      string
	InterfaceDetail map[string]networking.Port `json:"interfaceDetail"`
	VolumeDetail    map[string]storage.Volume  `json:"volumeDetail"`
}

func (serverInspect *ServerInspect) Print() {
	source := `
# Inspect For Instance {{.Server.Name}}

## 基本信息

1. 实例详情
	- **UUID**: {{.Server.Id}}
	- **Name**: {{.Server.Name}}
	- **Root BDM Type**: {{.Server.RootBdmType}}
	- **InstanceName**: {{.Server.InstanceName}}
	- **Description**: {{.Server.Description}}
1. 状态
	- **VmState**: {{.Server.VmState}}
	- **PowerState**: {{.PowerState}}
	- **TaskState**: {{.Server.TaskState}}
1. 节点
	- **Availability Zone**: {{.Server.AZ}}
	- **Host**: {{.Server.Host}}

## 规格

- **Name**: {{.Server.Flavor.OriginalName}}
- **Vcpus**: {{.Server.Flavor.Vcpus}}
- **Ram**: {{.Server.Flavor.Ram}}
- **Extra specs**:
{{ range $key, $value := .Server.Flavor.ExtraSpecs }}
	1. {{$key}} = {{$value}}
{{end}}

## 网卡
{{ range $index, $interface := .Interfaces }}
1. **PortId**: {{$interface.PortId}}
  - **Mac Address**: {{$interface.MacAddr}}
  - **IP Address**: {{ range $index, $ip := $interface.FixedIps }} {{$ip.IpAddress}} {{end}}
  - **Binding Type**: vnic_type:{{ getPortVnicType $interface.PortId }} vif_type:{{ getPortVifType $interface.PortId }}
  - **Binding Detail**: {{ getPortBindingDetail $interface.PortId }}
{{end}}

## 磁盘
{{ range $index, $volume := .Volumes }}
1. **Volume Id**: {{$volume.VolumeId}}
  - **Device**: {{$volume.Device}}
  - **Type**: {{ getVolumeType $volume.VolumeId }}
  - **Size**: {{ getVolumeSize $volume.VolumeId }}
  - **Image**: {{ getVolumeImage $volume.VolumeId }}

{{end}}
`
	templateBuffer := bytes.NewBuffer([]byte{})
	bufferWriter := bufio.NewWriter(templateBuffer)
	serverInspect.PowerState = serverInspect.Server.GetPowerState()

	tmpl := template.New("inspect")
	tmpl = tmpl.Funcs(template.FuncMap{
		"getVolumeType": func(volumeId string) string {
			if vol, ok := serverInspect.VolumeDetail[volumeId]; ok {
				return vol.VolumeType
			} else {
				return "-"
			}
		},
		"getVolumeSize": func(volumeId string) string {
			if vol, ok := serverInspect.VolumeDetail[volumeId]; ok {
				return fmt.Sprintf("%d GB", vol.Size)
			} else {
				return "-"
			}
		},
		"getVolumeImage": func(volumeId string) string {
			if vol, ok := serverInspect.VolumeDetail[volumeId]; ok {
				var image string
				if imageId, ok := vol.VolumeImageMetadata["image_id"]; ok {
					image = imageId
				}
				if imageName, ok := vol.VolumeImageMetadata["image_name"]; ok {
					image += fmt.Sprintf("(%s)", imageName)
				}
				return image
			} else {
				return "-"
			}
		},
		"getPortVnicType": func(portId string) string {
			if port, ok := serverInspect.InterfaceDetail[portId]; ok {
				return port.BindingVnicType
			} else {
				return "-"
			}
		},
		"getPortVifType": func(portId string) string {
			if port, ok := serverInspect.InterfaceDetail[portId]; ok {
				return port.BindingVifType
			} else {
				return "-"
			}
		},
		"getPortBindingDetail": func(portId string) string {
			if port, ok := serverInspect.InterfaceDetail[portId]; ok {
				bytes, _ := json.Marshal(port.BindingDetails)
				return string(bytes)
			} else {
				return "-"
			}
		},
	})
	tmpl, _ = tmpl.Parse(source)
	tmpl.Execute(bufferWriter, serverInspect)
	bufferWriter.Flush()
	result := markdown.Render(templateBuffer.String(), 100, 4)
	fmt.Println(string(result))
}
