package openstack

import (
	"os"
	"path"

	"github.com/samber/lo"
	"github.com/spf13/viper"
)

var cloud Cloud
var cloudName string
var CONF Config

type Config struct {
	HttpTimeoutSecond   int `yaml:"httpTimeoutSecond"`
	RetryWaitTimeSecond int `yaml:"retryWaitTimeSecond"`
	RetryCount          int `yaml:"retryCount"`

	Clouds map[string]Cloud `yaml:"clouds"`
}

type Cloud struct {
	TokenExpireTime int         `yaml:"tokenExpireTime"`
	Identity        Identity    `yaml:"identity"`
	Neutron         NeutronConf `yaml:"neutron"`
	Compute         Compute     `yaml:"compute"`
	RegionName      string      `yaml:"region_name" mapstructure:"region_name"`
	Auth            Auth        `yaml:"auth"`
}

func (c Cloud) Region() string {
	return lo.CoalesceOrEmpty(c.RegionName, "RegionOne")
}

type Auth struct {
	AuthUrl         string `yaml:"auth_url" mapstructure:"auth_url"`
	ProjectDomainId string `yaml:"project_domain_id" mapstructure:"project_domain_id"`
	UserDomainId    string `yaml:"user_domain_id" mapstructure:"user_domain_id"`
	ProjectName     string `yaml:"project_name" mapstructure:"project_name"`
	Username        string `yaml:"username" mapstructure:"username"`
	Password        string `yaml:"password" mapstructure:"password"`
}

type Api struct {
	Version string `yaml:"version"`
}
type Identity struct {
	Api Api `yaml:"api"`
}
type Compute struct {
	Api Api `yaml:"api"`
}
type NeutronConf struct {
	Endpoint string `yaml:"endpoint"`
}

func LoadConfig(file ...string) error {
	viper.SetConfigType("yaml")
	if len(file) > 0 && file[0] != "" {
		viper.SetConfigFile(file[0])
	} else {
		viper.SetConfigName("clouds.yaml")
		userConfDir, _ := os.UserConfigDir()
		if userConfDir != "" {
			userConfDir = path.Join(userConfDir, "skyman")
		}
		confDirs := []string{".", userConfDir, path.Join("/etc", "skyman")}

		for _, p := range confDirs {
			if p != "" {
				viper.AddConfigPath(p)
			}
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	viper.Unmarshal(&CONF)
	return nil
}

func SetOpenstackConfig(c Config) {
	CONF = c
}
func SetCloudName(name string) {
	cloudName = name
}
func CloudName() string {
	return cloudName
}
func CloudConfig() Cloud {
	return cloud
}

func init() {
	CONF = Config{}
}
