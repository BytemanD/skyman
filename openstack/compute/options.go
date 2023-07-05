package compute

type BlockDeviceMappingV2 struct {
	BootIndex          int    `json:"boot_index"`
	UUID               string `json:"uuid"`
	VolumeSize         int    `json:"volume_size"`
	SourceType         string `json:"source_type"`
	DestinationType    string `json:"destination_type"`
	DeleteOnTemination bool   `json:"delete_on_termination"`
}

type ServerOpt struct {
	Flavor               string                 `json:"flavorRef"`
	Image                string                 `json:"imageRef,omitempty"`
	Name                 string                 `json:"name"`
	Networks             interface{}            `json:"networks"`
	AvailabilityZone     string                 `json:"availability_zone,omitempty"`
	BlockDeviceMappingV2 []BlockDeviceMappingV2 `json:"block_device_mapping_v2,omitempty"`
}
