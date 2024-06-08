package server_actions

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

type ServerAction interface {
	Start() error
	// Rollback()
	Cleanup()
}
type ServerActionTest struct {
	Server       *nova.Server
	Client       *openstack.Openstack
	networkIndex int
}

func (t *ServerActionTest) ServerId() string {
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
	interval, maxInterval := 1, 10

	for i := 0; i <= 60; i++ {
		if err := t.RefreshServer(); err != nil {
			return err
		}
		progress := ""
		if showProgress {
			progress = fmt.Sprintf(", progress: %d", int(t.Server.Progress))
		}
		logging.Info("[%s] vmState=%s, powerState=%s, taskState=%s%s",
			t.Server.Id, t.Server.VmState, t.Server.GetPowerState(), t.Server.TaskState, progress)
		if t.Server.TaskState == "" {
			return nil
		}
		time.Sleep(time.Second * time.Duration(interval))
		if interval < maxInterval {
			interval += 1
		}
	}
	return fmt.Errorf("server task state is %s", t.Server.TaskState)
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
func (t *ServerActionTest) ServerMustHasInterface(portId string) error {
	interfaces, err := t.Client.NovaV2().Servers().ListInterfaces(t.Server.Id)
	if err != nil {
		return err
	}
	for _, vif := range interfaces {
		if vif.PortId == portId {
			return nil
		}
	}
	return fmt.Errorf("server has not interface: %s", portId)
}

func (t *ServerActionTest) ServerMustHasNotInterface(portId string) error {
	interfaces, err := t.Client.NovaV2().Servers().ListInterfaces(t.Server.Id)
	if err != nil {
		return err
	}
	for _, vif := range interfaces {
		if vif.PortId == portId {
			return fmt.Errorf("server has not interface: %s", portId)
		}
	}
	return nil
}
func (t *ServerActionTest) ServerMustHasVolume(volumeId string) error {
	volumes, err := t.Client.NovaV2().Servers().ListVolumes(t.Server.Id)
	if err != nil {
		return err
	}
	for _, vol := range volumes {
		if vol.VolumeId == volumeId {
			return nil
		}
	}
	return fmt.Errorf("server has not volume: %s", volumeId)
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

func (t ServerActionTest) Cleanup() {
	logging.Info("[%s] clean up", t.Server.Id)
}
