package common

import (
	"fmt"
	"os"

	"github.com/BytemanD/skyman/utility"
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
	Skip    bool       `yaml:"skip"`
	Config  CaseConfig `yaml:"config"`
}

type ServerActionsTestConf struct {
	Web     Web        `yaml:"web"`
	Default CaseConfig `yaml:"default"`
	Cases   []Case     `yaml:"cases"`
}

func LoadTaskConfig(taskFile string) (*ServerActionsTestConf, error) {
	bytes, err := os.ReadFile(taskFile)
	if err != nil {
		return nil, fmt.Errorf("read file %s failed: %s", taskFile, err)
	}
	testConf := ServerActionsTestConf{
		Web:     Web{Port: 80},
		Default: CaseConfig{},
	}
	if err = yaml.Unmarshal(bytes, &testConf); err != nil {
		return nil, fmt.Errorf("unmarshal test conf failed: %s", err)
	}
	return &testConf, nil
}

func NewActionCaseConfig(config CaseConfig, def CaseConfig) CaseConfig {
	caseConfig := CaseConfig{
		Workers:         utility.OneOfNumber(config.Workers, def.Workers, 1),
		ActionInterval:  utility.OneOfNumber(config.ActionInterval, def.ActionInterval),
		DeleteIfError:   utility.OneOfBoolean(config.DeleteIfError, def.DeleteIfError),
		DeleteIfSuccess: utility.OneOfBoolean(config.DeleteIfSuccess, def.DeleteIfSuccess),

		Flavors:          utility.OneOfStringArrays(config.Flavors, def.Flavors),
		Images:           utility.OneOfStringArrays(config.Images, def.Images),
		AvailabilityZone: utility.OneOfString(config.AvailabilityZone, def.AvailabilityZone),

		BootFromVolume: utility.OneOfBoolean(config.BootFromVolume, def.BootFromVolume),
		BootVolumeSize: utility.OneOfNumber(config.BootVolumeSize, def.BootVolumeSize, 50),
		BootVolumeType: utility.OneOfString(config.BootVolumeType, def.BootVolumeType),

		BootWithSG: utility.OneOfString(config.BootWithSG, def.BootWithSG),
		Networks:   utility.OneOfStringArrays(config.Networks, def.Networks),

		VolumeType: utility.OneOfString(config.VolumeType, def.VolumeType),
		VolumeSize: utility.OneOfNumber(config.VolumeSize, def.VolumeSize, 10),

		InterfaceHotplug: InterfaceHotplug{
			Nums: utility.OneOfNumber(config.InterfaceHotplug.Nums, def.InterfaceHotplug.Nums, 1),
		},
		VolumeHotplug: VolumeHotplug{
			Nums: utility.OneOfNumber(config.VolumeHotplug.Nums, def.VolumeHotplug.Nums, 1),
		},
		QGAChecker: QGAChecker{
			Enabled: utility.OneOfBoolean(config.QGAChecker.Enabled, def.QGAChecker.Enabled),
			GuestConnectTimeout: utility.OneOfNumber(config.QGAChecker.GuestConnectTimeout, def.QGAChecker.GuestConnectTimeout,
				DEFAULT_GUEST_CONNECT_TIMEOUT),
			QgaConnectTimeout: utility.OneOfNumber(config.QGAChecker.QgaConnectTimeout, def.QGAChecker.QgaConnectTimeout,
				DEFAULT_QGA_CONNECT_TIMEOUT),
		},
		LiveMigrate: LiveMigrateOptions{
			PingEnabled: utility.OneOfBoolean(config.LiveMigrate.PingEnabled, def.LiveMigrate.PingEnabled),
			PingInterval: utility.OneOfNumber(config.LiveMigrate.PingInterval, def.LiveMigrate.PingInterval,
				DEFAULT_PING_INTERVAL),
			MaxLoss: utility.OneOfNumber(config.LiveMigrate.MaxLoss, def.LiveMigrate.MaxLoss),
		},
		RevertSystem: RevertSystemConf{
			RepeatEveryTime: utility.OneOfNumber(config.RevertSystem.RepeatEveryTime, def.RevertSystem.RepeatEveryTime, 1),
		},
	}

	return caseConfig
}
