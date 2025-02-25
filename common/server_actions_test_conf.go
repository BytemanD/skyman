package common

import (
	"fmt"
	"os"

	"github.com/samber/lo"
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
		Workers:         lo.CoalesceOrEmpty(config.Workers, def.Workers, 1),
		ActionInterval:  lo.CoalesceOrEmpty(config.ActionInterval, def.ActionInterval),
		DeleteIfError:   lo.CoalesceOrEmpty(config.DeleteIfError, def.DeleteIfError),
		DeleteIfSuccess: lo.CoalesceOrEmpty(config.DeleteIfSuccess, def.DeleteIfSuccess),

		Flavors:          lo.CoalesceSliceOrEmpty(config.Flavors, def.Flavors),
		Images:           lo.CoalesceSliceOrEmpty(config.Images, def.Images),
		AvailabilityZone: lo.CoalesceOrEmpty(config.AvailabilityZone, def.AvailabilityZone),

		BootFromVolume: lo.CoalesceOrEmpty(config.BootFromVolume, def.BootFromVolume),
		BootVolumeSize: lo.CoalesceOrEmpty(config.BootVolumeSize, def.BootVolumeSize, 50),
		BootVolumeType: lo.CoalesceOrEmpty(config.BootVolumeType, def.BootVolumeType),

		BootWithSG: lo.CoalesceOrEmpty(config.BootWithSG, def.BootWithSG),
		Networks:   lo.CoalesceSliceOrEmpty(config.Networks, def.Networks),

		VolumeType: lo.CoalesceOrEmpty(config.VolumeType, def.VolumeType),
		VolumeSize: lo.CoalesceOrEmpty(config.VolumeSize, def.VolumeSize, 10),

		InterfaceHotplug: InterfaceHotplug{
			Nums: lo.CoalesceOrEmpty(config.InterfaceHotplug.Nums, def.InterfaceHotplug.Nums, 1),
		},
		VolumeHotplug: VolumeHotplug{
			Nums: lo.CoalesceOrEmpty(config.VolumeHotplug.Nums, def.VolumeHotplug.Nums, 1),
		},
		QGAChecker: QGAChecker{
			Enabled: lo.CoalesceOrEmpty(config.QGAChecker.Enabled, def.QGAChecker.Enabled),
			GuestConnectTimeout: lo.CoalesceOrEmpty(config.QGAChecker.GuestConnectTimeout, def.QGAChecker.GuestConnectTimeout,
				DEFAULT_GUEST_CONNECT_TIMEOUT),
			QgaConnectTimeout: lo.CoalesceOrEmpty(config.QGAChecker.QgaConnectTimeout, def.QGAChecker.QgaConnectTimeout,
				DEFAULT_QGA_CONNECT_TIMEOUT),
		},
		LiveMigrate: LiveMigrateOptions{
			PingEnabled: lo.CoalesceOrEmpty(config.LiveMigrate.PingEnabled, def.LiveMigrate.PingEnabled),
			PingInterval: lo.CoalesceOrEmpty(config.LiveMigrate.PingInterval, def.LiveMigrate.PingInterval,
				DEFAULT_PING_INTERVAL),
			MaxLoss: lo.CoalesceOrEmpty(config.LiveMigrate.MaxLoss, def.LiveMigrate.MaxLoss),
		},
		RevertSystem: RevertSystemConf{
			RepeatEveryTime: lo.CoalesceOrEmpty(config.RevertSystem.RepeatEveryTime, def.RevertSystem.RepeatEveryTime, 1),
		},
	}

	return caseConfig
}
