package server_actions

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/checkers"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

type ServerAction interface {
	Start() error
	// Rollback()
	Cleanup()
	ServerId() string
}
type ServerActionTest struct {
	Server       *nova.Server
	Client       *openstack.Openstack
	networkIndex int
}

func (t ServerActionTest) ServerId() string {
	return t.Server.Id
}

func (t *ServerActionTest) RefreshServer() error {
	server, err := t.Client.NovaV2().Servers().Show(t.Server.Id)
	if err != nil {
		return nil
	}
	t.Server = server
	return nil
}

func (t *ServerActionTest) WaitServerTaskFinished(showProgress bool) error {
	serverError := false
	utility.Retry(
		utility.RetryCondition{
			Timeout:      time.Second * 60 * 20,
			IntervalMin:  time.Second,
			IntervalMax:  time.Second * 10,
			IntervalStep: time.Second * 2,
		},
		func() bool {
			if err := t.RefreshServer(); err != nil {
				serverError = true
				return false
			}
			progress := ""
			if showProgress {
				progress = fmt.Sprintf(", progress: %d", int(t.Server.Progress))
			}
			logging.Info("[%s] %s%s", t.Server.Id, t.Server.AllStatus(), progress)
			return t.Server.TaskState != ""
		},
	)
	if serverError {
		return fmt.Errorf("server is error")
	} else {
		return nil
	}
}
func (t *ServerActionTest) nextNetwork() (string, error) {
	if len(common.CONF.Test.Networks) == 0 {
		return "", fmt.Errorf("the num of networks == 0")
	}
	if t.networkIndex >= len(common.CONF.Test.Networks)-1 {
		t.networkIndex = 0
	}
	defer func() { t.networkIndex += 1 }()
	return common.CONF.Test.Networks[t.networkIndex], nil
}
func (t *ServerActionTest) lastVolume() (*nova.VolumeAttachment, error) {
	volumes, err := t.Client.NovaV2().Servers().ListVolumes(t.Server.Id)
	if err != nil {
		return nil, err
	}
	if len(volumes) == 0 {
		return nil, fmt.Errorf("has no volume")
	}
	return &volumes[len(volumes)-1], nil
}

func (t *ServerActionTest) ServerMustHasNotInterface(portId string) error {
	interfaces, err := t.Client.NovaV2().Servers().ListInterfaces(t.Server.Id)
	if err != nil {
		return err
	}
	for _, vif := range interfaces {
		if vif.PortId == portId {
			return fmt.Errorf("server has no interface: %s", portId)
		}
	}
	return nil
}

func (t *ServerActionTest) ServerMustHasNotVolume(volumeId string) error {
	volumes, err := t.Client.NovaV2().Servers().ListVolumes(t.Server.Id)
	if err != nil {
		return err
	}
	for _, vol := range volumes {
		if vol.VolumeId == volumeId {
			return fmt.Errorf("server has not volume: %s", volumeId)
		}
	}
	return nil
}
func (t ServerActionTest) ServerMustNotError() error {
	if t.Server.IsError() {
		return fmt.Errorf("server status is error")
	}
	return nil
}

func (t ServerActionTest) CreateBlankVolume() (*cinder.Volume, error) {
	options := map[string]interface{}{
		"size": common.CONF.Test.VolumeSize,
	}
	if common.CONF.Test.VolumeType != "" {
		options["volume_type"] = common.CONF.Test.VolumeType
	}
	volume, err := t.Client.CinderV2().Volumes().Create(options)
	if err != nil {
		return nil, err
	}
	for i := 0; i <= 60; i++ {
		volume, err = t.Client.CinderV2().Volumes().Show(volume.Id)
		if err != nil {
			return nil, err
		}
		logging.Info("[%s] volume status is: %s", t.ServerId(), volume.Status)
		if volume.IsAvailable() {
			return volume, nil
		}
		if volume.IsError() {
			return volume, fmt.Errorf("volume is error")
		}
		time.Sleep(time.Second * 2)
	}
	return volume, fmt.Errorf("create volume timeout")
}

func (t ServerActionTest) getCheckers() (checkers.ServerCheckers, error) {
	serverCheckers, err := checkers.GetServerCheckers(t.Client, t.Server)
	if err != nil {
		return nil, fmt.Errorf("get server checker failed: %s", err)
	}
	return serverCheckers, nil
}
func (t ServerActionTest) MakesureServerRunning() error {
	if serverCheckers, err := t.getCheckers(); err == nil {
		return serverCheckers.MakesureServerRunning()
	}
	return nil
}

func (t ServerActionTest) getServerBootOption(name string) nova.ServerOpt {
	opt := nova.ServerOpt{
		Name:             name,
		Flavor:           TEST_FLAVORS[0].Id,
		Image:            common.CONF.Test.Images[0],
		AvailabilityZone: common.CONF.Test.AvailabilityZone,
	}
	if len(common.CONF.Test.Networks) >= 1 {
		opt.Networks = []nova.ServerOptNetwork{
			{UUID: common.CONF.Test.Networks[0]},
		}
	} else {
		logging.Warning("boot without network")
	}
	if common.CONF.Test.BootWithSG != "" {
		opt.SecurityGroups = append(opt.SecurityGroups,
			neutron.SecurityGroup{
				Resource: model.Resource{Name: common.CONF.Test.BootWithSG},
			})
	}
	if common.CONF.Test.BootFromVolume {
		opt.BlockDeviceMappingV2 = []nova.BlockDeviceMappingV2{
			{
				UUID:               common.CONF.Test.Images[0],
				VolumeSize:         common.CONF.Test.BootVolumeSize,
				SourceType:         "image",
				DestinationType:    "volume",
				VolumeType:         common.CONF.Test.BootVolumeType,
				DeleteOnTemination: true,
			},
		}
	} else {
		opt.Image = common.CONF.Test.Images[0]
	}
	return opt
}

type EmptyCleanup struct {
}

func (t EmptyCleanup) Cleanup() {
	// logging.Info("[%s] clean up", t.Server.Id)
}
