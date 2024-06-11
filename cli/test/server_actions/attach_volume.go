package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
)

type ServerAttachVolume struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerAttachVolume) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	logging.Info("[%s] creating volume", t.ServerId())
	volume, err := t.CreateBlankVolume()
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

type ServerDetachVolume struct {
	ServerActionTest
	EmptyCleanup
}

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

type ServerVolumeHotPlug struct {
	ServerActionTest
	attachments []string
}

func (t *ServerVolumeHotPlug) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	for i := 0; i < common.CONF.Test.VolumeHotplug.Nums; i++ {
		logging.Info("[%s] attach volume (%d)", t.ServerId(), i+1)

		logging.Info("[%s] creating volume", t.ServerId())
		volume, err := t.CreateBlankVolume()
		if err != nil {
			return fmt.Errorf("create volume failed: %s", err)
		}

		_, err = t.Client.NovaV2().Servers().AddVolume(t.Server.Id, volume.Id)
		if err != nil {
			return err
		}
		logging.Info("[%s] attaching volume %s", t.Server.Id, volume.Id)
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if err := t.ServerMustNotError(); err != nil {
			return err
		}
		if err := t.ServerMustHasVolume(volume.Id); err != nil {
			return err
		}
		t.attachments = append(t.attachments, volume.Id)
	}

	for _, volId := range t.attachments {
		err := t.Client.NovaV2().Servers().DeleteVolume(t.Server.Id, volId)
		if err != nil {
			return err
		}
		logging.Info("[%s] detaching volume %s", t.ServerId(), volId)
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if err := t.ServerMustNotError(); err != nil {
			return err
		}
		if err := t.ServerMustHasNotVolume(volId); err != nil {
			return err
		}
	}
	return nil
}

func (t ServerVolumeHotPlug) Cleanup() {
	logging.Info("[%s] cleanup %d volumes", t.ServerId(), len(t.attachments))
	for _, volId := range t.attachments {
		logging.Info("[%s] deleting volume %s", t.ServerId(), volId)
		err := t.Client.CinderV2().Volumes().Delete(volId, true, true)
		if err != nil {
			logging.Error("[%s] delete volume %s failed", t.ServerId(), volId)
		}
	}
}
