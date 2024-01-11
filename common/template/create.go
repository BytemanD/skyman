package template

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Server struct {
	NamePrefix       string `yaml:"namePrefix"`
	Name             string `yaml:"name"`
	Flavor           string `yaml:"flavor"`
	Image            string `yaml:"image"`
	Network          string `yaml:"network"`
	VolumeBoot       bool   `yaml:"volumeBoot"`
	VolumeType       string `yaml:"volumeType"`
	VolumeSize       uint16 `yaml:"volumeSize"`
	AvailabilityZone string `yaml:"availabilityZone"`
	Min              uint16 `yaml:"min"`
	Max              uint16 `yaml:"max"`
}

type CreateTemplate struct {
	Server Server `yaml:"server"`
}

func LoadCreateTemplate(file string) (*CreateTemplate, error) {
	template := CreateTemplate{}
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &template)
	if template.Server.NamePrefix == "" {
		template.Server.NamePrefix = "server-"
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
