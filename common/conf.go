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
	Debug          bool   `yaml:"debug"`
	Format         string `yaml:"format"`
	Language       string `yaml:"language"`
	HttpTimeout    int    `yaml:"httpTimeout"`
	LogFile        string `yaml:"logFile"`
	EnableLogColor bool   `yaml:"enableLogColor"`

	Auth  Auth  `yaml:"auth"`
	Iperf Iperf `yaml:"iperf"`
	Test  Test  `yaml:"test"`
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
type Iperf struct {
	GuestPath     string `yaml:"guestPath"`
	LocalPath     string `yaml:"guestPath"`
	ClientOptions string `yaml:"clientOptions"`
	ServerOptions string `yaml:"serverOptions"`
}
type InterfaceHotplug struct {
	Nums int `yaml:"nums"`
}
type VolumeHotplug struct {
	Nums int `yaml:"nums"`
}

var (
	DEFAULT_GUEST_CONNECT_TIMEOUT = 60 * 5
	DEFAULT_QGA_CONNECT_TIMEOUT   = 60 * 10
)

type QGAChecker struct {
	Enabled             bool `yaml:"enabled"`
	GuestConnectTimeout int  `yaml:"guestConnectTimeout"`
	QgaConnectTimeout   int  `yaml:"qgaConnectTimeout"`
}

type LiveMigrateOptions struct {
	PingEnabled bool `yaml:"pingEnabled"`
}

type Test struct {
	Tasks            int                `yaml:"tasks"`
	Workers          int                `yaml:"workers"`
	DeleteIfError    bool               `yaml:"deleteIfError"`
	AvailabilityZone string             `yaml:"availabilityZone"`
	BootFromVolume   bool               `yaml:"bootFromVolume"`
	BootVolumeSize   uint16             `yaml:"bootVolumeSize"`
	BootVolumeType   string             `yaml:"bootVolumeType"`
	Flavors          []string           `yaml:"flavors"`
	Images           []string           `yaml:"images"`
	Networks         []string           `yaml:"networks"`
	VolumeType       string             `yaml:"volumeType"`
	VolumeSize       int                `yaml:"volumeSize"`
	Actions          []string           `yaml:"volumeType"`
	InterfaceHotplug InterfaceHotplug   `yaml:"interfaceHotplug"`
	VolumeHotplug    VolumeHotplug      `yaml:"volumeHotplug"`
	UseServers       []string           `yaml:"userServers"`
	ActionInterval   int                `yaml:"actionInterval"`
	QGAChecker       QGAChecker         `yaml:"qgaChecker"`
	LiveMigrate      LiveMigrateOptions `yaml:"liveMigrate"`
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
	if CONF.Test.VolumeSize <= 0 {
		CONF.Test.VolumeSize = 10
	}
	if CONF.Test.BootVolumeSize <= 0 {
		CONF.Test.BootVolumeSize = 50
	}
	if CONF.Test.BootVolumeSize <= 0 {
		CONF.Test.BootVolumeSize = 50
	}
	if CONF.Test.Tasks <= 0 {
		CONF.Test.Tasks = 1
	}
	if CONF.Test.InterfaceHotplug.Nums == 0 {
		CONF.Test.InterfaceHotplug.Nums = 1
	}
	if CONF.Test.VolumeHotplug.Nums == 0 {
		CONF.Test.VolumeHotplug.Nums = 1
	}
	if CONF.Test.QGAChecker.GuestConnectTimeout == 0 {
		CONF.Test.QGAChecker.GuestConnectTimeout = DEFAULT_GUEST_CONNECT_TIMEOUT
	}
	if CONF.Test.QGAChecker.QgaConnectTimeout == 0 {
		CONF.Test.QGAChecker.QgaConnectTimeout = DEFAULT_QGA_CONNECT_TIMEOUT
	}

	return nil
}
