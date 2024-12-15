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

type CaseConfig struct {
	Flavors        []string `yaml:"flavors"`
	Images         []string `yaml:"images"`
	Workers        int      `yaml:"workers"`
	ActionInterval int      `yaml:"actionInterval"`

	Web              Web    `yaml:"web"`
	DeleteIfError    bool   `yaml:"deleteIfError"`
	DeleteIfSuccess  bool   `yaml:"deleteIfSuccess"`
	AvailabilityZone string `yaml:"availabilityZone"`
	BootFromVolume   bool   `yaml:"bootFromVolume"`
	BootVolumeSize   uint16 `yaml:"bootVolumeSize"`
	BootVolumeType   string `yaml:"bootVolumeType"`
	BootWithSG       string `yaml:"bootWithSG"`

	Networks   []string `yaml:"networks"`
	VolumeType string   `yaml:"volumeType"`
	VolumeSize int      `yaml:"volumeSize"`

	InterfaceHotplug InterfaceHotplug   `yaml:"interfaceHotplug"`
	VolumeHotplug    VolumeHotplug      `yaml:"volumeHotplug"`
	QGAChecker       QGAChecker         `yaml:"qgaChecker"`
	LiveMigrate      LiveMigrateOptions `yaml:"liveMigrate"`
	RevertTimes      int                `yaml:"revertTimes"`
}
type Case struct {
	Actions string     `yaml:"actions"`
	Config  CaseConfig `yaml:"config"`
}

type ServerActionsTestConf struct {
	Default CaseConfig `yaml:"default"`
	Cases   []Case     `yaml:"cases"`
}

var (
	TASK_CONF ServerActionsTestConf
)

func DefaultServerActionsTtestConf() ServerActionsTestConf {
	return ServerActionsTestConf{
		Default: CaseConfig{
			Workers:          1,
			BootVolumeSize:   50,
			VolumeSize:       10,
			Web:              Web{Port: 80},
			InterfaceHotplug: InterfaceHotplug{Nums: 1},
			VolumeHotplug:    VolumeHotplug{Nums: 1},
			QGAChecker: QGAChecker{
				GuestConnectTimeout: DEFAULT_GUEST_CONNECT_TIMEOUT,
				QgaConnectTimeout:   DEFAULT_QGA_CONNECT_TIMEOUT,
			},
			LiveMigrate: LiveMigrateOptions{
				PingInterval: DEFAULT_PING_INTERVAL,
			},
		},
	}
}

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
	TASK_CONF = DefaultServerActionsTtestConf()
	viper.Unmarshal(&TASK_CONF)

	return nil
}
