package common

import (
	"fmt"
	"os"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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

type RevertSystemConf struct {
	RepeatEveryTime int `yaml:"repeatEveryTime"`
}
type CaseConfig struct {
	Flavors        []string `yaml:"flavors"`
	Images         []string `yaml:"images"`
	Workers        int      `yaml:"workers"`
	ActionInterval int      `yaml:"actionInterval"`

	DeleteIfError    bool   `yaml:"deleteIfError"`
	DeleteIfSuccess  bool   `yaml:"deleteIfSuccess"`
	AvailabilityZone string `yaml:"availabilityZone"`
	BootFromVolume   bool   `yaml:"bootFromVolume"`
	BootVolumeSize   uint16 `yaml:"bootVolumeSize"`
	BootVolumeType   string `yaml:"bootVolumeType"`

	BootWithSG string   `yaml:"bootWithSG"`
	Networks   []string `yaml:"networks"`

	VolumeType string `yaml:"volumeType"`
	VolumeSize int    `yaml:"volumeSize"`

	InterfaceHotplug InterfaceHotplug   `yaml:"interfaceHotplug"`
	VolumeHotplug    VolumeHotplug      `yaml:"volumeHotplug"`
	QGAChecker       QGAChecker         `yaml:"qgaChecker"`
	LiveMigrate      LiveMigrateOptions `yaml:"liveMigrate"`
	RevertSystem     RevertSystemConf   `yaml:"revertSystem"`
}
type Case struct {
	Name    string     `yaml:"name"`
	Actions string     `yaml:"actions"`
	Config  CaseConfig `yaml:"config"`
}

type ServerActionsTestConf struct {
	Web     Web        `yaml:"web"`
	Default CaseConfig `yaml:"default"`
	Cases   []Case     `yaml:"cases"`
}

var (
	TASK_CONF ServerActionsTestConf
)

func DefaultServerActionsTtestConf() ServerActionsTestConf {
	return ServerActionsTestConf{
		Web:     Web{Port: 80},
		Default: CaseConfig{},
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

	bytes, err := os.ReadFile(taskFile)
	if err != nil {
		return fmt.Errorf("read file %s failed: %s", taskFile, err)
	}
	testConf := ServerActionsTestConf{
		Web:     Web{Port: 80},
		Default: CaseConfig{},
	}
	if err = yaml.Unmarshal(bytes, &testConf); err != nil {
		return fmt.Errorf("unmarshal test conf failed: %s", err)
	}
	TASK_CONF = testConf
	return nil
}

func NewActionCaseConfig(config CaseConfig, def CaseConfig) CaseConfig {
	caseConfig := CaseConfig{
		Workers:         OneOfNumber(config.Workers, def.Workers, 1),
		ActionInterval:  OneOfNumber(config.Workers, def.Workers),
		DeleteIfError:   OneOfBoolean(config.DeleteIfError, def.DeleteIfError),
		DeleteIfSuccess: OneOfBoolean(config.DeleteIfSuccess, def.DeleteIfSuccess),

		Flavors:          OneOfStringArrays(config.Flavors, def.Flavors),
		Images:           OneOfStringArrays(config.Images, def.Images),
		AvailabilityZone: OneOfString(config.AvailabilityZone, def.AvailabilityZone),

		BootFromVolume: OneOfBoolean(config.BootFromVolume, def.BootFromVolume),
		BootVolumeSize: OneOfNumber(config.BootVolumeSize, def.BootVolumeSize, 50),
		BootVolumeType: OneOfString(config.BootVolumeType, def.BootVolumeType),

		BootWithSG: OneOfString(config.BootWithSG, def.BootWithSG),
		Networks:   OneOfStringArrays(config.Networks, def.Networks),

		VolumeType: OneOfString(config.VolumeType, def.VolumeType),
		VolumeSize: OneOfNumber(config.VolumeSize, def.VolumeSize, 10),

		InterfaceHotplug: InterfaceHotplug{
			Nums: OneOfNumber(config.InterfaceHotplug.Nums, def.InterfaceHotplug.Nums, 1),
		},
		VolumeHotplug: VolumeHotplug{
			Nums: OneOfNumber(config.VolumeHotplug.Nums, def.VolumeHotplug.Nums, 1),
		},
		QGAChecker: QGAChecker{
			Enabled: OneOfBoolean(config.QGAChecker.Enabled, def.QGAChecker.Enabled),
			GuestConnectTimeout: OneOfNumber(config.QGAChecker.GuestConnectTimeout, def.QGAChecker.GuestConnectTimeout,
				DEFAULT_GUEST_CONNECT_TIMEOUT),
			QgaConnectTimeout: OneOfNumber(config.QGAChecker.QgaConnectTimeout, def.QGAChecker.QgaConnectTimeout,
				DEFAULT_QGA_CONNECT_TIMEOUT),
		},
		LiveMigrate: LiveMigrateOptions{
			PingEnabled: OneOfBoolean(config.LiveMigrate.PingEnabled, def.LiveMigrate.PingEnabled),
			PingInterval: OneOfNumber(config.LiveMigrate.PingInterval, def.LiveMigrate.PingInterval,
				DEFAULT_PING_INTERVAL),
			MaxLoss: OneOfNumber(config.LiveMigrate.MaxLoss, def.LiveMigrate.MaxLoss),
		},
		RevertSystem: RevertSystemConf{
			RepeatEveryTime: OneOfNumber(config.RevertSystem.RepeatEveryTime, def.RevertSystem.RepeatEveryTime, 1),
		},
	}

	return caseConfig
}
