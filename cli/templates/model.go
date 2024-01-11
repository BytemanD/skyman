package templates

import (
	"os"

	"gopkg.in/yaml.v3"
)

type BaseResource struct {
	Name string
	Id   string
}
type Default struct {
	ServerNamePrefix string `yaml:"serverNamePrefix"`
}

type Server struct {
	Name             string `yaml:"name"`
	Flavor           BaseResource
	Image            BaseResource
	Network          string `yaml:"network"`
	VolumeBoot       bool   `yaml:"volumeBoot"`
	VolumeType       string `yaml:"volumeType"`
	VolumeSize       uint16 `yaml:"volumeSize"`
	AvailabilityZone string `yaml:"availabilityZone"`
	Min              uint16 `yaml:"min"`
	Max              uint16 `yaml:"max"`
}

type Flavor struct {
	Id         string            `yaml:"id,omitempty"`
	Name       string            `yaml:"name,omitempty"`
	Vcpus      int               `yaml:"vcpus,omitempty"`
	Ram        int               `yaml:"ram,omitempty"`
	Disk       int               `yaml:"disk,omitempty"`
	Swap       int               `yaml:"swap,omitempty"`
	RXTXFactor float32           `yaml:"rxtx_factor,omitempty"`
	ExtraSpecs map[string]string `yaml:"extra_specs,omitempty"`
}

type CreateTemplate struct {
	Default Default `yaml:"default"`
	Flavor  Flavor  `yaml:"flavor"`
	Server  Server  `yaml:"server"`
}

func LoadCreateTemplate(file string) (*CreateTemplate, error) {
	template := CreateTemplate{}
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &template)
	if template.Default.ServerNamePrefix == "" {
		template.Default.ServerNamePrefix = "server-"
	}
	if template.Server.Min == 0 {
		template.Server.Min = 1
	}
	if template.Server.Max == 0 {
		template.Server.Max = template.Server.Min
	}
	if template.Server.VolumeSize == 0 {
		template.Server.VolumeSize = 10
	}
	return &template, err
}
