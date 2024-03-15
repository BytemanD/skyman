package compute

import (
	"strings"

	"github.com/BytemanD/skyman/openstack/networking"
)

type BlockDeviceMappingV2 struct {
	BootIndex          int    `json:"boot_index"`
	UUID               string `json:"uuid,omitempty"`
	VolumeSize         uint16 `json:"volume_size,omitempty"`
	SourceType         string `json:"source_type,omitempty"`
	DestinationType    string `json:"destination_type,omitempty"`
	VolumeType         string `json:"volume_type,omitempty"`
	DeleteOnTemination bool   `json:"delete_on_termination,omitempty"`
}
type ServerOptNetwork struct {
	UUID string `json:"uuid,omitempty"`
	Port string `json:"port,omitempty"`
}
type ServerOpt struct {
	Flavor               string                     `json:"flavorRef,omitempty"`
	Image                string                     `json:"imageRef,omitempty"`
	Name                 string                     `json:"name,omitempty"`
	Networks             interface{}                `json:"networks,omitempty"`
	AvailabilityZone     string                     `json:"availability_zone,omitempty"`
	BlockDeviceMappingV2 []BlockDeviceMappingV2     `json:"block_device_mapping_v2,omitempty"`
	MinCount             uint16                     `json:"min_count"`
	MaxCount             uint16                     `json:"max_count"`
	UserData             string                     `json:"user_data,omitempty"`
	KeyName              string                     `json:"key_name,omitempty"`
	AdminPass            string                     `json:"adminPass,omitempty"`
	SecurityGroups       []networking.SecurityGroup `json:"security_groups,omitempty"`
}

func ParseServerOptyNetworks(nics []string) []ServerOptNetwork {
	networks := []ServerOptNetwork{}
	for _, nic := range nics {
		values := strings.Split(nic, "=")
		if values[0] == "net-id" {
			networks = append(networks, ServerOptNetwork{UUID: values[1]})
		} else if values[0] == "port-id" {
			networks = append(networks, ServerOptNetwork{Port: values[1]})
		}
	}
	return networks
}
