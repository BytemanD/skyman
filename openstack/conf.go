package openstack

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

const ENV_CLOUD_NAME = "OS_CLOUD_NAME"

var cloud Cloud
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
type NeutronConf struct {
	Endpoint string `yaml:"endpoint"`
}

func LoadConfig(file ...string) error {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("OS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(
		"AUTH.USER.NAME", "USERNAME",
		"AUTH.USER.PASSWORD", "PASSWORD",
		"AUTH.USER.DOMAIN", "USER_DOMAIN",
		"AUTH.PROJECT", "PROJECT",
		"AUTH.REGION.ID", "REGION_NAME",
		"NEUTRON.ENDPOINT", "NEUTRON_ENDPOINT",
		".", "_"))

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

func CloudConfig() Cloud {
	return cloud
}

// export OS_REGION_NAME=RegionOne
// export OS_AUTH_URL=http://keystone.region1.dev:35357/v3
// export OS_USERNAME=admin
// export OS_PASSWORD=PASSWORD
// export OS_PROJECT_NAME=admin
// export OS_USER_DOMAIN_NAME=default
// export OS_PROJECT_DOMAIN_NAME=default
// export OS_IDENTITY_API_VERSION=3
// export OS_IMAGE_API_VERSION=2
// export OS_VOLUME_API_VERSION=2
// export OS_BAREMETAL_API_VERSION=latest
// export IRONIC_API_VERSION=latest
func loadFromEnv() {
	cloud.RegionName = lo.CoalesceOrEmpty(os.Getenv("OS_REGION_NAME"), cloud.RegionName)
	cloud.Auth.AuthUrl = lo.CoalesceOrEmpty(os.Getenv("OS_AUTH_URL"), cloud.Auth.AuthUrl)
	cloud.Auth.ProjectDomainId = lo.CoalesceOrEmpty(os.Getenv("OS_PROJECT_DOMAIN_NAME"), cloud.Auth.ProjectDomainId)
	cloud.Auth.UserDomainId = lo.CoalesceOrEmpty(os.Getenv("OS_USER_DOMAIN_NAME"), cloud.Auth.UserDomainId)
	cloud.Auth.ProjectName = lo.CoalesceOrEmpty(os.Getenv("OS_PROJECT_NAME"), cloud.Auth.ProjectName)
	cloud.Auth.Username = lo.CoalesceOrEmpty(os.Getenv("OS_USERNAME"), cloud.Auth.Username)
	cloud.Auth.Password = lo.CoalesceOrEmpty(os.Getenv("OS_PASSWORD"), cloud.Auth.Password)

	cloud.Identity.Api.Version = lo.CoalesceOrEmpty(os.Getenv("OS_IDENTITY_API_VERSION"), cloud.Identity.Api.Version)
	cloud.Neutron.Endpoint = lo.CoalesceOrEmpty(os.Getenv("OS_NEUTRON_ENDPOINT"), cloud.Neutron.Endpoint)

}
func connectCloud() (*Openstack, error) {
	// 更新默认配置
	cloud.TokenExpireTime = lo.CoalesceOrEmpty(cloud.TokenExpireTime, 60*30)

	if cloud.Auth.AuthUrl == "" {
		return nil, fmt.Errorf("auth url is emptu, forget to load env or set cloud name?")
	}
	conn := NewClient(
		cloud.Auth.AuthUrl,
		model.User{
			Name:     cloud.Auth.Username,
			Domain:   model.Domain{Name: cloud.Auth.UserDomainId},
			Password: cloud.Auth.Password,
		},
		model.Project{
			Name:   cloud.Auth.ProjectName,
			Domain: model.Domain{Name: cloud.Auth.ProjectDomainId},
		},
		cloud.Region(),
	)
	conn.cloudConfig = cloud
	conn.AuthPlugin.SetLocalTokenExpire(cloud.TokenExpireTime)
	if CONF.HttpTimeoutSecond > 0 {
		conn.SetHttpTimeout(time.Second * time.Duration(CONF.HttpTimeoutSecond))
	}
	if CONF.RetryWaitTimeSecond > 0 {
		conn.SetRetryWaitTime(time.Second * time.Duration(CONF.RetryWaitTimeSecond))
	}
	if CONF.RetryCount > 0 {
		conn.SetRetryCount(CONF.RetryCount)
	}
	console.Debug("new openstack client, HttpTimeoutSecond=%d RetryWaitTimeSecond=%d RetryCount=%d",
		CONF.HttpTimeoutSecond, CONF.RetryWaitTimeSecond, CONF.RetryCount,
	)
	console.Debug("cloud: %v", cloud.Auth.ProjectDomainId)
	console.Debug("new openstack client, HttpTimeoutSecond=%d RetryWaitTimeSecond=%d RetryCount=%d",
		CONF.HttpTimeoutSecond, CONF.RetryWaitTimeSecond, CONF.RetryCount,
	)
	if _, err := conn.AuthPlugin.GetToken(); err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}
	return conn, nil
}
func GetOne(name string) (*Openstack, error) {
	if c, ok := CONF.Clouds[name]; !ok {
		return nil, fmt.Errorf("cloud %s not found", name)
	} else {
		cloud = c
		return connectCloud()
	}
}

// 如果指定 cloud 名称, 优先使用 cloud对应的配置;
// 否则从环境变量读取 cloud;
// 最后，使用默认的cloud.
func Connect(name ...string) (*Openstack, error) {
	if len(name) > 0 && name[0] != "" {
		return GetOne(name[0])
	}
	if os.Getenv(ENV_CLOUD_NAME) != "" {
		return GetOne(os.Getenv(ENV_CLOUD_NAME))
	}
	loadFromEnv()
	if c, ok := CONF.Clouds["default"]; ok {
		cloud = c
	}
	return connectCloud()
}

func init() {
	CONF = Config{}
}
