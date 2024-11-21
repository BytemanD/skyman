package common

import (
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/spf13/viper"
)

var (
	DEFAULT_GUEST_CONNECT_TIMEOUT = 60 * 5
	DEFAULT_QGA_CONNECT_TIMEOUT   = 60 * 10
	DEFAULT_PING_INTERVAL         = float32(1.0)
)

type Web struct {
	Port int `yaml:"port"`
}

type InterfaceHotplug struct {
	Nums int `yaml:"nums"`
}
type VolumeHotplug struct {
	Nums int `yaml:"nums"`
}
type QGAChecker struct {
	Enabled             bool `yaml:"enabled"`
	GuestConnectTimeout int  `yaml:"guestConnectTimeout"`
	QgaConnectTimeout   int  `yaml:"qgaConnectTimeout"`
}

type LiveMigrateOptions struct {
	PingEnabled  bool    `yaml:"pingEnabled"`
	PingInterval float32 `yaml:"pingInterval"`
	MaxLoss      int     `yaml:"maxLoss"`
}

type ServerActionsTestConf struct {
	Total            int                `yaml:"total"`
	Workers          int                `yaml:"workers"`
	Web              Web                `yaml:"web"`
	ActionTasks      []string           `yaml:"actionTasks"`
	DeleteIfError    bool               `yaml:"deleteIfError"`
	DeleteIfSuccess  bool               `yaml:"deleteIfSuccess"`
	AvailabilityZone string             `yaml:"availabilityZone"`
	BootFromVolume   bool               `yaml:"bootFromVolume"`
	BootVolumeSize   uint16             `yaml:"bootVolumeSize"`
	BootVolumeType   string             `yaml:"bootVolumeType"`
	BootWithSG       string             `yaml:"bootWithSG"`
	Flavors          []string           `yaml:"flavors"`
	Images           []string           `yaml:"images"`
	Networks         []string           `yaml:"networks"`
	VolumeType       string             `yaml:"volumeType"`
	VolumeSize       int                `yaml:"volumeSize"`
	InterfaceHotplug InterfaceHotplug   `yaml:"interfaceHotplug"`
	VolumeHotplug    VolumeHotplug      `yaml:"volumeHotplug"`
	UseServers       []string           `yaml:"userServers"`
	ActionInterval   int                `yaml:"actionInterval"`
	QGAChecker       QGAChecker         `yaml:"qgaChecker"`
	LiveMigrate      LiveMigrateOptions `yaml:"liveMigrate"`
}

var (
	TASK_CONF ServerActionsTestConf
)

func LoadTaskConfig(taskFile string) error {
	viper.SetConfigType("yaml")
	if taskFile != "" {
		viper.SetConfigFile(taskFile)
	}
	logging.Info("load server-actions test from file %s", taskFile)
	viper.GetViper().ConfigFileUsed()
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	viper.Unmarshal(&TASK_CONF)
	if TASK_CONF.VolumeSize <= 0 {
		TASK_CONF.VolumeSize = 10
	}
	if TASK_CONF.BootVolumeSize <= 0 {
		TASK_CONF.BootVolumeSize = 50
	}
	if TASK_CONF.BootVolumeSize <= 0 {
		TASK_CONF.BootVolumeSize = 50
	}
	if TASK_CONF.Total <= 0 {
		TASK_CONF.Total = 1
	}
	if TASK_CONF.InterfaceHotplug.Nums == 0 {
		TASK_CONF.InterfaceHotplug.Nums = 1
	}
	if TASK_CONF.VolumeHotplug.Nums == 0 {
		TASK_CONF.VolumeHotplug.Nums = 1
	}
	if TASK_CONF.QGAChecker.GuestConnectTimeout == 0 {
		TASK_CONF.QGAChecker.GuestConnectTimeout = DEFAULT_GUEST_CONNECT_TIMEOUT
	}
	if TASK_CONF.QGAChecker.QgaConnectTimeout == 0 {
		TASK_CONF.QGAChecker.QgaConnectTimeout = DEFAULT_QGA_CONNECT_TIMEOUT
	}
	if TASK_CONF.LiveMigrate.PingInterval <= 0 {
		TASK_CONF.LiveMigrate.PingInterval = DEFAULT_PING_INTERVAL
	}
	if TASK_CONF.Web.Port <= 0 {
		TASK_CONF.Web.Port = 80
	}

	return nil
}
