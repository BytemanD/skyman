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
	RefreshServer() error
	Start() error
	Skip() (bool, string)
	TearDown() error
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
	server, err := t.Client.NovaV2().Server().Show(t.Server.Id)
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
	volumes, err := t.Client.NovaV2().Server().ListVolumes(t.Server.Id)
	if err != nil {
		return nil, err
	}
	if len(volumes) == 0 {
		return nil, fmt.Errorf("has no volume")
	}
	return &volumes[len(volumes)-1], nil
}

func (t *ServerActionTest) ServerMustHasNotInterface(portId string) error {
	interfaces, err := t.Client.NovaV2().Server().ListInterfaces(t.Server.Id)
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
	volumes, err := t.Client.NovaV2().Server().ListVolumes(t.Server.Id)
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
	volume, err := t.Client.CinderV2().Volume().Create(options)
	if err != nil {
		return nil, err
	}
	for i := 0; i <= 60; i++ {
		volume, err = t.Client.CinderV2().Volume().Show(volume.Id)
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
func (t *ServerActionTest) GetRootVolume() (*nova.VolumeAttachment, error) {
	if t.Server.RootBdmType != "volume" {
		return nil, fmt.Errorf("root bdm is not volume")
	}
	volumes, err := t.Client.NovaV2().Server().ListVolumes(t.Server.Id)
	if err != nil {
		return nil, err
	}
	for _, volume := range volumes {
		if volume.Device == t.Server.RootDeviceName {
			return &volume, nil
		}
	}
	return nil, fmt.Errorf("root volume not found")
}
func (t *ServerActionTest) WaitSnapshotCreated(snapshotId string) error {
	return utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:     time.Minute * 5,
			IntervalMin: time.Second * 2},
		[]string{"SnapshotIsNotAvailable"},
		func() error {
			snapshot, err := t.Client.CinderV2().Snapshot().Show(snapshotId)
			if err != nil {
				return err
			}
			logging.Info("[%s] snapshot status is %s", t.ServerId(), snapshot.Status)
			switch snapshot.Status {
			case "error":
				return fmt.Errorf("snapshot is error")
			case "available":
				return nil
			default:
				return utility.NewSnapshotIsNotAvailable(snapshotId)
			}
		},
	)
}
func (t *ServerActionTest) WaitVolumeTaskDone(volumeId string) error {
	return utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:      time.Minute * 10,
			IntervalMin:  time.Second,
			IntervalStep: time.Second,
			IntervalMax:  time.Second * 10},
		[]string{"VolumeHasTaskError"},
		func() error {
			vol, err := t.Client.CinderV2().Volume().Show(volumeId)
			if err != nil {
				return err
			}
			logging.Info("[%s] volume %s state=%s, staskState=%s", t.ServerId(),
				volumeId, vol.Status, vol.TaskStatus)
			if vol.IsError() {
				return fmt.Errorf("volume %s is error", volumeId)
			}
			if (vol.IsAvailable() || vol.IsInuse()) && vol.TaskStatus == "" {
				return nil
			}
			return utility.NewVolumeHasTaskError(volumeId)
		},
	)
}

type EmptyCleanup struct {
}

func (t EmptyCleanup) TearDown() error {
	return nil
}
func (t EmptyCleanup) Skip() (bool, string) {
	return false, ""
}
