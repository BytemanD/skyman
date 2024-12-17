package common

import (
	"os"
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
	Debug               bool   `yaml:"debug"`
	Format              string `yaml:"format"`
	Language            string `yaml:"language"`
	HttpTimeoutSecond   int    `yaml:"httpTimeoutSecond"`
	RetryWaitTimeSecond int    `yaml:"retryWaitTimeSecond"`
	RetryCount          int    `yaml:"retryCount"`
	LogFile             string `yaml:"logFile"`
	EnableLogColor      bool   `yaml:"enableLogColor"`

	Auth     Auth        `yaml:"auth"`
	Identity Identity    `yaml:"identity"`
	Neutron  NeutronConf `yaml:"neutron"`
}
type Auth struct {
	Url             string          `yaml:"url"`
	Region          keystone.Region `yaml:"region"`
	User            auth.User       `yaml:"user"`
	Project         auth.Project    `yaml:"project"`
	TokenExpireTime int             `yaml:"tokenExpireTime"`
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

func DefaultConfGroup() ConfGroup {
	return ConfGroup{
		Identity: Identity{
			Api: Api{Version: "3"},
		},
		Auth: Auth{
			TokenExpireTime: DEFAULT_TOKEN_EXPIRE_TIME,
		},
	}
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
		"NEUTRON.ENDPOINT", "NEUTRON_ENDPOINT",
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
	CONF = DefaultConfGroup()
	viper.Unmarshal(&CONF)
	i18n.InitLocalizer(CONF.Language)

	// 环境变量
	if os.Getenv("OS_IDENTITY_API_VERSION") != "" {
		CONF.Identity.Api.Version = os.Getenv("OS_IDENTITY_API_VERSION")
	}

	return nil
}
