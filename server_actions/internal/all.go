package internal

import (
	"fmt"
	"sort"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/server_actions/checkers"
	"github.com/BytemanD/skyman/utility"
)

var (
	ACTION_REBOOT            = "reboot"
	ACTION_HARD_REBOOT       = "hard_reboot"
	ACTION_STOP              = "stop"
	ACTION_START             = "start"
	ACTION_PAUSE             = "pause"
	ACTION_UNPAUSE           = "unpause"
	ACTION_MIGRATE           = "migrate"
	ACTION_LIVE_MIGRATE      = "live_migrate"
	ACTION_SHELVE            = "shelve"
	ACTION_UNSHELVE          = "unshelve"
	ACTION_TOGGLE_SHELVE     = "toggle_shelve"
	ACTION_REBUILD           = "rebuild"
	ACTION_RESIZE            = "resize"
	ACTION_RENAME            = "rename"
	ACTION_SUSPEND           = "suspend"
	ACTION_RESUME            = "resume"
	ACTION_TOGGLE_SUSPEND    = "toggle_suspend"
	ACTION_ATTACH_NET        = "net_attach"
	ACTION_ATTACH_PORT       = "port_attach"
	ACTION_DETACH_PORT       = "port_detach"
	ACTION_INTERFACE_HOTPLUG = "interface_hotplug"
	ACTION_ATTACH_VOLUME     = "volume_attach"
	ACTION_DETACH_VOLUME     = "volume_detach"
	ACTION_VOLUME_HOTPLUG    = "volume_hotplug"
	ACTION_VOLUME_EXTEND     = "volume_extend"
	ACTION_REVERT_SYSTEM     = "revert_system"
	ACTION_NOP               = "nop"
)

type ServerAction interface {
	RefreshServer() error
	Start() error
	TearDown() error
	ServerId() string
	SetConfig(c common.CaseConfig)
}
type ServerActionTest struct {
	Server       *nova.Server
	Client       *openstack.Openstack
	networkIndex int
	Config       common.CaseConfig
}

func (t *ServerActionTest) SetConfig(c common.CaseConfig) {
	t.Config = c
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
	if len(t.Config.Networks) == 0 {
		return "", fmt.Errorf("the num of networks == 0")
	}
	if t.networkIndex >= len(t.Config.Networks)-1 {
		t.networkIndex = 0
	}
	defer func() { t.networkIndex += 1 }()
	return t.Config.Networks[t.networkIndex], nil
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
		"size": t.Config.VolumeSize,
	}
	if t.Config.VolumeType != "" {
		options["volume_type"] = t.Config.VolumeType
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
	return checkers.GetServerCheckers(t.Client, t.Server, t.Config.QGAChecker)
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
		Image:            t.Config.Images[0],
		AvailabilityZone: t.Config.AvailabilityZone,
	}
	if len(t.Config.Networks) >= 1 {
		opt.Networks = []nova.ServerOptNetwork{
			{UUID: t.Config.Networks[0]},
		}
	} else {
		logging.Warning("boot without network")
	}
	if t.Config.BootWithSG != "" {
		opt.SecurityGroups = append(opt.SecurityGroups,
			neutron.SecurityGroup{
				Resource: model.Resource{Name: t.Config.BootWithSG},
			})
	}
	if t.Config.BootFromVolume {
		opt.BlockDeviceMappingV2 = []nova.BlockDeviceMappingV2{
			{
				UUID:               t.Config.Images[0],
				VolumeSize:         t.Config.BootVolumeSize,
				SourceType:         "image",
				DestinationType:    "volume",
				VolumeType:         t.Config.BootVolumeType,
				DeleteOnTemination: true,
			},
		}
	} else {
		opt.Image = t.Config.Images[0]
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
			logging.Info("[%s] snapshot %s status is %s", t.ServerId(), snapshot.Id, snapshot.Status)
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

type ActionCreatorFunc func(server *nova.Server, client *openstack.Openstack) ServerAction

type Actions map[string]ActionCreatorFunc

func (a Actions) Keys() []string {
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func (a Actions) Contains(name string) bool {
	for k := range a {
		if k == name {
			return true
		}
	}
	return false
}
func (a Actions) Get(name string, server *nova.Server, client *openstack.Openstack) ServerAction {
	if creator, ok := a[name]; ok {
		return creator(server, client)
	}
	return nil
}

func (a Actions) register(name string, creator ActionCreatorFunc) {
	if _, ok := a[name]; ok {
		logging.Warning("action %s already exists", name)
		return
	}
	a[name] = creator
}

var VALID_ACTIONS = Actions{}

func init() {
	VALID_ACTIONS[ACTION_REBOOT] = func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerReboot{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	}
	VALID_ACTIONS.register(ACTION_REBOOT, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerReboot{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_HARD_REBOOT, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerHardReboot{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_STOP, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerStop{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_START, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerStart{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_PAUSE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerPause{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_UNPAUSE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerUnpause{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_MIGRATE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerMigrate{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_LIVE_MIGRATE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerLiveMigrate{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_SHELVE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerShelve{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_UNSHELVE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerUnshelve{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_TOGGLE_SHELVE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerToggleShelve{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_REBUILD, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerRebuild{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_RESIZE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerResize{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_RENAME, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerRename{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_SUSPEND, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerSuspend{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_RESUME, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerResume{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_TOGGLE_SUSPEND, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerToggleSuspend{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_ATTACH_PORT, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerAttachPort{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_ATTACH_NET, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerAttachNet{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_DETACH_PORT, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerDetachInterface{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_ATTACH_VOLUME, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerAttachVolume{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_DETACH_VOLUME, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerDetachVolume{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_INTERFACE_HOTPLUG, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerAttachHotPlug{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_VOLUME_HOTPLUG, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerVolumeHotPlug{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_VOLUME_EXTEND, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerExtendVolume{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_REVERT_SYSTEM, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerRevertToSnapshot{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_NOP, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerActionNop{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
}
