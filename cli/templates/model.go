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
type BlockDeviceMappingV2 struct {
	BootIndex           int    `yaml:"bootIndex"`
	UUID                string `yaml:"uuid"`
	VolumeSize          uint16 `yaml:"volumeSize"`
	SourceType          string `yaml:"sourceType"`
	DestinationType     string `yaml:"destinationType"`
	VolumeType          string `yaml:"volumeType"`
	DeleteOnTermination bool   `yaml:"deleteOnTermination"`
}
type SecurityGroup struct {
	Name string `yaml:"name,omitempty"`
}
type Nic struct {
	Name string `yaml:"name"`
	UUID string `yaml:"uuid"`
	Port string `yaml:"port"`
}
type Server struct {
	Name     string `yaml:"name"`
	Flavor   BaseResource
	Image    BaseResource
	Networks []Nic `yaml:"networks"`

	AvailabilityZone     string                 `yaml:"availabilityZone"`
	Min                  uint16                 `yaml:"min"`
	Max                  uint16                 `yaml:"max"`
	BlockDeviceMappingV2 []BlockDeviceMappingV2 `yaml:"blockDeviceMappingV2,omitempty"`
	UserData             string                 `yaml:"userData"`
	SecurityGroups       []string               `yaml:"securityGroups,omitempty"`
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
type Subnet struct {
	Name      string `yaml:"name,omitempty"`
	Cidr      string `yaml:"cidr,omitempty"`
	IpVersion int    `yaml:"ipVersion,omitempty"`
}
type Network struct {
	Name    string   `yaml:"name,omitempty"`
	Subnets []Subnet `yaml:"subnets,omitempty"`
}
type CreateTemplate struct {
	Default  Default   `yaml:"default"`
	Flavors  []Flavor  `yaml:"flavors"`
	Networks []Network `yaml:"networks"`
	Servers  []Server  `yaml:"servers"`
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
	for i, _ := range template.Servers {
		if template.Servers[i].Min == 0 {
			template.Servers[i].Min = 1
		}
		if template.Servers[i].Max == 0 {
			template.Servers[i].Max = template.Servers[i].Min
		}
	}
	return &template, err
}
