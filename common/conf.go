package common

import (
	"strings"

	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/openstack/model/keystone"
	"github.com/spf13/viper"
)

var CONF_FILES = []string{
	"etc/skyman.yaml",
	"/etc/skyman/skyman.yaml",
}

var (
	CONF      ConfGroup
	CONF_FILE string
)

var (
	FORMAT_TABLE_LIGHT        = "table-light"
	FORMAT_TABLE              = "table"
	DEFAULT_TOKEN_EXPIRE_TIME = 60 * 30
)

type ConfGroup struct {
	Debug       bool   `yaml:"debug"`
	Format      string `yaml:"format"`
	Language    string `yaml:"language"`
	HttpTimeout int    `yaml:"httpTimeout"`

	Auth Auth `yaml:"auth"`
}
type Auth struct {
	Url             string          `yaml:"url"`
	Region          keystone.Region `yaml:"region"`
	User            auth.User       `yaml:"user"`
	Project         auth.Project    `yaml:"project"`
	TokenExpireTime int             `yaml:"tokenExpireTime"`
}

type Server struct {
	Flavor           string `yaml:"flavor"`
	Image            string `yaml:"image"`
	Network          string `yaml:"network"`
	VolumeBoot       bool   `yaml:"volumeBoot"`
	VolumeType       string `yaml:"volumeType"`
	VolumeSize       uint16 `yaml:"volumeSize"`
	AvailabilityZone string `yaml:"availabilityZone"`
	NamePrefix       string `yaml:"namePrefix"`
}

func LoadConfig(configFile string) error {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("OS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(
		"AUTH.USER.NAME", "USERNAME",
		"AUTH.USER.PASSWORD", "PASSWORD",
		"AUTH.USER.DOMAIN", "USER_DOMAIN",
		"AUTH.PROJECT", "PROJECT",
		"AUTH.REGION", "REGION",
		".", "_"))

	viper.SetConfigType("yaml")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("skyman.yaml")
		viper.AddConfigPath("./etc")
		viper.AddConfigPath("/etc/skyman")
	}
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	viper.Unmarshal(&CONF)
	i18n.InitLocalizer(CONF.Language)
	if CONF.Auth.TokenExpireTime <= 0 {
		CONF.Auth.TokenExpireTime = DEFAULT_TOKEN_EXPIRE_TIME
	}
	return nil
}
