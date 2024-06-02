package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
)

type ServerAttachVolume struct {
	ServerActionTest
	networkIndex int
}

func (t ServerAttachVolume) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	options := map[string]interface{}{
		"size": common.CONF.Test.VolumeSize,
	}
	if common.CONF.Test.VolumeType != "" {
		options["volume_type"] = common.CONF.Test.VolumeType
	}
	volume, err := t.Client.CinderV2().Volumes().Create(options)
	if err != nil {
		return fmt.Errorf("create volume failed: %s", err)
	}
	_, err = t.Client.NovaV2().Servers().AddVolume(t.Server.Id, volume.Id)
	logging.Info("[%s] attaching volume", t.Server.Id)
	if err != nil {
		return err
	}
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server status is error")
	}
	volumes, err := t.Client.NovaV2().Servers().ListVolumes(t.Server.Id)
	if err != nil {
		return err
	}
	for _, vol := range volumes {
		if vol.VolumeId == volume.Id {
			return nil
		}
	}
	return fmt.Errorf("server has no volume %s", volume.Id)
}

type ServerDetachVolume struct{ ServerActionTest }

func (t *ServerDetachVolume) lastVolume() (string, error) {
	volumes, err := t.Client.NovaV2().Servers().ListVolumes(t.Server.Id)
	if err != nil {
		return "", err
	}
	if len(volumes) == 0 {
		return "", fmt.Errorf("has no volume")
	}
	return volumes[len(volumes)-1].VolumeId, nil
}

func (t ServerDetachVolume) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	volumeId, err := t.lastVolume()
	if err != nil {
		return err
	}
	err = t.Client.NovaV2().Servers().DeleteVolume(t.Server.Id, volumeId)
	if err != nil {
		return err
	}
	logging.Info("[%s] detaching volume %s", t.Server.Id, volumeId)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server status is error")
	}
	volumes, err := t.Client.NovaV2().Servers().ListVolumes(t.Server.Id)
	if err != nil {
		return err
	}
	for _, vol := range volumes {
		if vol.VolumeId == volumeId {
			return fmt.Errorf("volume %s is not detached", volumeId)
		}
	}
	return nil
}
