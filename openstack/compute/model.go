package compute

import (
	"encoding/json"
	"fmt"
)

type Flavor struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	OriginalName string            `json:"original_name"`
	Ram          int               `json:"ram"`
	Vcpus        int               `json:"vcpus"`
	Disk         int               `json:"disk"`
	Swap         int               `json:"swap"`
	ExtraSpecs   map[string]string `json:"extra_specs"`
}

type Fault struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Details string `json:"details"`
}

type Server struct {
	Id           string               `json:"id"`
	Name         string               `json:"name"`
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
	Id string `json:"id"`
}
type Address struct {
	MacAddr string `json:"OS-EXT-IPS-MAC:mac_addr"`
	Version int    `json:"version"`
	Addr    string `json:"addr"`
	Type    string `json:"OS-EXT-IPS:type"`
}
type ServerBody struct {
	Server Server `json:"server"`
}

type ServersBody struct {
	Servers []Server `json:"servers"`
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
	extraSpecs, _ := json.Marshal(server.Flavor.ExtraSpecs)
	return string(extraSpecs)
}
func (server Server) GetFaultString() string {
	fault, _ := json.Marshal(server.Fault)
	return string(fault)
}
