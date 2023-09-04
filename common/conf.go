package common

import (
	"github.com/BytemanD/stackcrud/common/i18n"
	"github.com/BytemanD/stackcrud/openstack/identity"
	"github.com/spf13/viper"
)

var CONF_FILES = []string{
	"etc/stackcrud.yaml",
	"/etc/stackcrud/stackcrud.yaml",
}

var (
	CONF      ConfGroup
	CONF_FILE string
)

var (
	FORMAT_TABLE_LIGHT = "table-light"
	FORMAT_TABLE       = "table"
)

type ConfGroup struct {
	Debug    bool   `yaml:"debug"`
	Format   string `yaml:"format"`
	Language string `yaml:"language"`

	Auth   Auth   `yaml:"auth"`
	Server Server `yaml:"server"`
}
type Auth struct {
	Url             string           `yaml:"url"`
	RegionName      string           `yaml:"regionName"`
	User            identity.User    `yaml:"user"`
	Project         identity.Project `yaml:"project"`
	TokenExpireTime int              `yaml:"tokenExpireTime"`
}

type Server struct {
	Flavor           string `yaml:"flavor"`
	Image            string `yaml:"image"`
	Network          string `yaml:"network"`
	VolumeBoot       bool   `yaml:"volumeBoot"`
	VolumeSize       uint16 `yaml:"volumeSize"`
	AvailabilityZone string `yaml:"availabilityZone"`
	NamePrefix       string `yaml:"namePrefix"`
}

func LoadConfig(configFile string) error {
	viper.SetConfigType("yaml")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("stackcrud.yaml")
		viper.AddConfigPath("./etc")
		viper.AddConfigPath("/etc/stackcrud")
	}
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.Unmarshal(&CONF)
	i18n.InitLocalizer(CONF.Language)
	return nil
}
