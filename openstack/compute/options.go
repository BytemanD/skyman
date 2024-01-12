package compute

import "strings"

type BlockDeviceMappingV2 struct {
	BootIndex          int    `json:"boot_index"`
	UUID               string `json:"uuid"`
	VolumeSize         uint16 `json:"volume_size"`
	SourceType         string `json:"source_type"`
	DestinationType    string `json:"destination_type"`
	VolumeType         string `json:"volume_type"`
	DeleteOnTemination bool   `json:"delete_on_termination"`
}
type ServerOptNetwork struct {
	UUID   string `json:"uuid,omitempty"`
	PortId string `json:"port_id,omitempty"`
}
type ServerOpt struct {
	Flavor               string                 `json:"flavorRef,omitempty"`
	Image                string                 `json:"imageRef,omitempty"`
	Name                 string                 `json:"name,omitempty"`
	Networks             interface{}            `json:"networks,omitempty"`
	AvailabilityZone     string                 `json:"availability_zone,omitempty"`
	BlockDeviceMappingV2 []BlockDeviceMappingV2 `json:"block_device_mapping_v2,omitempty"`
	MinCount             uint16                 `json:"min_count"`
	MaxCount             uint16                 `json:"max_count"`
	UserData             string                 `json:"user_data,omitempty"`
	KeyName              string                 `json:"key_name,omitempty"`
	AdminPass            string                 `json:"adminPass,omitempty"`
}

func ParseServerOptyNetworks(nics []string) []ServerOptNetwork {
	networks := []ServerOptNetwork{}
	for _, nic := range nics {
		values := strings.Split(nic, "=")
		if values[0] == "net-id" {
			networks = append(networks, ServerOptNetwork{UUID: values[1]})
		} else if values[0] == "port-id" {
			networks = append(networks, ServerOptNetwork{PortId: values[1]})
		}
	}
	return networks
}
