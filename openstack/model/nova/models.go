package nova

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/dustin/go-humanize"
)

var POWER_STATE = []string{
	"NOSTATE", "Running", "", "Paused", "ShutDown", "", "Crashed", "Suspend",
}

type Flavor struct {
	Id           string      `json:"id,omitempty"`
	Name         string      `json:"name,omitempty"`
	OriginalName string      `json:"original_name,omitempty"`
	Ram          int         `json:"ram,omitempty"`
	Vcpus        int         `json:"vcpus,omitempty"`
	Disk         int         `json:"disk"`
	Swap         interface{} `json:"swap,omitempty"`
	RXTXFactor   float32     `json:"rxtx_factor,omitempty"`
	ExtraSpecs   ExtraSpecs  `json:"extra_specs,omitempty"`
	IsPublic     bool        `json:"os-flavor-access:is_public,omitempty"`
	Ephemeral    int         `json:"OS-FLV-DISABLED:ephemeral,omitempty"`
	Disabled     bool        `json:"OS-FLV-DISABLED:disabled,omitempty"`
}

func (flavor Flavor) Marshal() string {
	flavorMarshal, _ := json.Marshal(flavor)
	return string(flavorMarshal)
}
func (flavor Flavor) BaseInfo() string {
	return fmt.Sprintf("vcpu=%d, ram=%d", flavor.Vcpus, flavor.Ram)
}

func (flavor Flavor) HumanRam() string {
	return humanize.IBytes(uint64(flavor.Ram) * utility.MB)
}

type ExtraSpecs map[string]string

func (extraSpecs ExtraSpecs) GetList() []string {
	properties := []string{}
	for k, v := range extraSpecs {
		properties = append(properties, fmt.Sprintf("%s=%s", k, v))
	}
	return properties
}
func (extraSpecs ExtraSpecs) Get(key string) string {
	return extraSpecs[key]
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

type AddressList []Address

func (a AddressList) Addrs() []string {
	addrs := []string{}
	for _, address := range a {
		addrs = append(addrs, address.Addr)
	}
	return addrs
}

type Server struct {
	model.Resource
	TaskState          string                  `json:"OS-EXT-STS:task_state,omitempty"`
	PowerState         int                     `json:"OS-EXT-STS:power_state,omitempty"`
	VmState            string                  `json:"OS-EXT-STS:vm_state,omitempty"`
	Host               string                  `json:"OS-EXT-SRV-ATTR:host,omitempty"`
	HypervisorHostname string                  `json:"OS-EXT-SRV-ATTR:hypervisor_hostname,omitempty"`
	AZ                 string                  `json:"OS-EXT-AZ:availability_zone,omitempty"`
	Flavor             Flavor                  `json:"flavor,omitempty"`
	Image              interface{}             `json:"image,omitempty"`
	Fault              Fault                   `json:"fault,omitempty"`
	Addresses          map[string]AddressList  `json:"addresses,omitempty"`
	InstanceName       string                  `json:"OS-EXT-SRV-ATTR:instance_name,omitempty"`
	ConfigDriver       string                  `json:"config_drive,omitempty"`
	Created            string                  `json:"created,omitempty"`
	Updated            string                  `json:"updated,omitempty"`
	TerminatedAt       string                  `json:"OS-SRV-USG:terminated_at,omitempty"`
	LaunchedAt         string                  `json:"OS-SRV-USG:launched_at,omitempty"`
	UserId             string                  `json:"user_id,omitempty"`
	Description        string                  `json:"description,omitempty"`
	RootBdmType        string                  `json:"root_bdm_type,omitempty"`
	RootDeviceName     string                  `json:"OS-EXT-SRV-ATTR:root_device_name,omitempty"`
	KeyName            string                  `json:"key_name,omitempty"`
	SecurityGroups     []neutron.SecurityGroup `json:"security_groups,omitempty"`
	Progress           float32                 `json:"progress"`
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

type ServersResp struct {
	model.RequestId
	Items []Server
}

func (s Server) ImageId() string {
	if p, ok := s.Image.(map[string]interface{}); ok {
		return p["id"].(string)
	}
	return ""
}

func (s Server) ImageName() string {
	if p, ok := s.Image.(map[string]interface{}); ok {
		if p["name"] == nil {
			return ""
		} else if name, ok := p["name"].(string); ok {
			return name
		}
	}
	return ""
}
func (s *Server) SetImageName(name string) {
	if p, ok := s.Image.(map[string]string); ok {
		p["name"] = name
	}
	s.Image = s
	// return ""
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
	networksMap := map[string]string{}

	for net, addresses := range server.Addresses {
		networksMap[net] = strings.Join(addresses.Addrs(), ", ")
	}
	networks := []string{}
	for k, v := range networksMap {
		networks = append(networks, fmt.Sprintf("%s=%s", k, v))
	}
	return networks
}
func (server Server) GetFlavorExtraSpecsString() string {
	var extraList []string
	for key, value := range server.Flavor.ExtraSpecs {
		extraList = append(extraList, key+"="+value)
	}
	sort.Strings(extraList)
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
func (server Server) StatusIs(status string) bool {
	return strings.EqualFold(server.Status, status)
}
func (server Server) InResize() bool {
	return server.Status == "VERIFY_RESIZE" || server.Status == "RESIZE"
}
func (server Server) IsActive() bool {
	return server.Status == "ACTIVE"
}
func (server Server) IsShelved() bool {
	return server.Status == "SHELVED" || server.Status == "SHELVED_OFFLOADED"
}
func (server Server) IsError() bool {
	return strings.ToUpper(server.Status) == "ERROR"
}
func (server Server) IsMigrating() bool {
	return strings.ToUpper(server.Status) == "MIGRATING"
}
func (server Server) IsStopped() bool {
	return server.StatusIs("SHUTOFF")
}
func (server Server) IsPaused() bool {
	return server.StatusIs("PAUSED")
}
func (server Server) IsSuspended() bool {
	return server.StatusIs("SUSPENDED")
}
func (server Server) GuestHostname() string {
	return strings.Replace(server.Name, ":", "", -1)
}
func (server Server) IsRunning() bool {
	return strings.ToUpper(server.GetPowerState()) == "RUNNING"
}

type Service struct {
	model.Resource
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
	model.Resource
	Host               string                 `json:"host"`
	HypervisorHostname string                 `json:"hypervisor_hostname"`
	HostIp             string                 `json:"host_ip"`
	Status             string                 `json:"status"`
	State              string                 `json:"state"`
	Type               string                 `json:"hypervisor_type"`
	Version            int                    `json:"hypervisor_version"`
	Uptime             string                 `json:"uptime"`
	Vcpus              int                    `json:"vcpus"`
	VcpusUsed          int                    `json:"vcpus_used"`
	MemoryMB           int                    `json:"memory_mb"`
	MemoryMBUsed       int                    `json:"memory_mb_used"`
	ExtraResources     map[string]interface{} `json:"extra_resources"`
	CpuInfo            CpuInfo                `json:"cpu_info"`

	NumaNode0Hugepages map[string]interface{} `json:"numa_node_0_hugepages"`
	NumaNode1Hugepages map[string]interface{} `json:"numa_node_1_hugepages"`
	NumaNode2Hugepages map[string]interface{} `json:"numa_node_2_hugepages"`
	NumaNode3Hugepages map[string]interface{} `json:"numa_node_3_hugepages"`
	NumaNode4Hugepages map[string]interface{} `json:"numa_node_4_hugepages"`
	NumaNode5Hugepages map[string]interface{} `json:"numa_node_5_hugepages"`
	NumaNode6Hugepages map[string]interface{} `json:"numa_node_6_hugepages"`
	NumaNode7Hugepages map[string]interface{} `json:"numa_node_7_hugepages"`

	NumaNode0Cpuset map[string]interface{} `json:"numa_node_0_cpuset"`
	NumaNode1Cpuset map[string]interface{} `json:"numa_node_1_cpuset"`
	NumaNode2Cpuset map[string]interface{} `json:"numa_node_2_cpuset"`
	NumaNode3Cpuset map[string]interface{} `json:"numa_node_3_cpuset"`
	NumaNode4Cpuset map[string]interface{} `json:"numa_node_4_cpuset"`
	NumaNode5Cpuset map[string]interface{} `json:"numa_node_5_cpuset"`
	NumaNode6Cpuset map[string]interface{} `json:"numa_node_6_cpuset"`
	NumaNode7Cpuset map[string]interface{} `json:"numa_node_7_cpuset"`

	Servers []HypervisorServer

	NumaNodes map[string]NumaNode
}

type NumaNode struct {
	HugePages NumaNodeHugePages
	CpuSet    NumaNodeCpuSet
}

type NumaNodeHugePages struct {
	Total    float32
	Free     float32
	Reserved float32
	Used     float32
}
type NumaNodeCpuSet struct {
	Total    int
	Free     int
	Reserved int
	Used     int
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
func (hypervisor *Hypervisor) SetNumaNode(data []byte) error {
	dataMap := struct {
		Hypervisor map[string]interface{}
	}{}

	if err := json.Unmarshal(data, &dataMap); err != nil {
		return err
	}
	regCpuset, _ := regexp.Compile("numa_node_([0-9]+)_cpuset")
	regHugePages, _ := regexp.Compile("numa_node_([0-9]+)_hugepages")

	hypervisor.NumaNodes = map[string]NumaNode{}
	for k, v := range dataMap.Hypervisor {
		matchHygePages := regHugePages.FindStringSubmatch(k)
		if len(matchHygePages) > 0 {
			nodeIndex := (matchHygePages[0])
			hugePage := NumaNodeHugePages{}
			bytes, _ := json.Marshal(v)
			json.Unmarshal(bytes, &hugePage)
			if node, ok := hypervisor.NumaNodes[nodeIndex]; ok {
				node.HugePages = hugePage
				hypervisor.NumaNodes[nodeIndex] = node
			} else {
				hypervisor.NumaNodes[nodeIndex] = NumaNode{HugePages: hugePage}
			}
			continue
		}
		matchCpuSet := regCpuset.FindStringSubmatch(k)
		if len(matchCpuSet) > 0 {
			nodeIndex := (matchCpuSet[0])
			fmt.Println("4444444", nodeIndex)
			cpuset := NumaNodeCpuSet{}
			bytes, _ := json.Marshal(v)
			json.Unmarshal(bytes, &cpuset)
			if node, ok := hypervisor.NumaNodes[nodeIndex]; ok {
				node.CpuSet = cpuset
				hypervisor.NumaNodes[nodeIndex] = node
			} else {
				hypervisor.NumaNodes[nodeIndex] = NumaNode{CpuSet: cpuset}
			}
			continue
		}
	}
	return nil
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
	model.Resource
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

func (actionWithEvents InstanceAction) GetSpendTime() (float64, error) {
	// event.StartTime
	if actionWithEvents.StartTime == "" {
		return 0, fmt.Errorf("no start time")
	}
	var lastEventFinishTime string
	for _, event := range actionWithEvents.Events {
		if event.FinishTime > lastEventFinishTime {
			lastEventFinishTime = event.FinishTime
		}
	}
	if lastEventFinishTime == "" {
		return 0, fmt.Errorf("no finish time")
	}

	logging.Debug("%s spend time: %s ~ %s", actionWithEvents.Action,
		actionWithEvents.StartTime, lastEventFinishTime)
	startTime, _ := time.Parse("2006-01-02T15:04:05", actionWithEvents.StartTime)
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
	model.RequestId
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
	Type     string `json:"type,omitempty"`
	Url      string `json:"url,omitempty"`
	Protocol string `json:"protocol,omitempty"`
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
	Id               int               `json:"id,omitempty""`
	Name             string            `json:"name,omitempty""`
	AvailabilityZone string            `json:"availability_zone,omitempty""`
	Deleted          bool              `json:"deleted,omitempty""`
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

func ParseExtraSpecsMap(extraSpecs []string) ExtraSpecs {
	extraSpecsMap := ExtraSpecs{}
	for _, property := range extraSpecs {
		kv := strings.Split(property, "=")
		extraSpecsMap[kv[0]] = kv[1]
	}
	return extraSpecsMap
}

type PageInfo struct {
	PageSize int `json:"page_size"`
}
type Capacity struct {
	AllowedSoldNum int    `json:"allowed_sold_num"`
	AZ             string `json:"az"`
	FlavorId       string `json:"flavor_id"`
}
type FlavorCapacities struct {
	PageInfo   PageInfo   `json:"page_info"`
	Capacities []Capacity `json:"capacities"`
}

// {"injected_file_content_bytes": -1, "metadata_items": -1, "server_group_members": -1,
//
//	"server_groups": -1, "ram": -1, "floating_ips": 10, "key_pairs": -1,
//
// "id": "4f362d490e5848ab9f1be44b5479b084",
//
// "instances": -1, "security_group_rules": 20, "injected_files": -1, "cores": -1,
//
//	"fixed_ips": -1, "injected_file_path_bytes": -1, "security_groups": 10}}
type QuotaSet struct {
	Instances                int `json:"instances"`
	Cores                    int `json:"cores"`
	Ram                      int `json:"ram"`
	MetadataItems            int `json:"metadata_items"`
	Keypairs                 int `json:"key_pairs"`
	FloatingIps              int `json:"floating_ips"`
	SecurityGroups           int `json:"server_groups"`
	SecurityGroupsMembers    int `json:"server_group_members"`
	FixedIps                 int `json:"fixed_ips"`
	InjectedFiles            int `json:"injected_files"`
	InjectedFileContentBytes int `json:"injected_file_content_bytes"`
	InjectedFilePathBytes    int `json:"injected_file_path_bytes"`
}
