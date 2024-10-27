package server_actions

import (
	"fmt"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/checkers"
	"github.com/BytemanD/skyman/common"
)

type ServerAttachVolume struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerAttachVolume) Start() error {
	logging.Info("[%s] creating volume", t.ServerId())
	volume, err := t.CreateBlankVolume()
	if err != nil {
		return fmt.Errorf("create volume failed: %s", err)
	}
	attachment, err := t.Client.NovaV2().Server().AddVolume(t.Server.Id, volume.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] attaching volume on %s", t.Server.Id, attachment.Device)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if err := t.ServerMustNotError(); err != nil {
		return err
	}
	serverCheckers, err := checkers.GetServerCheckers(t.Client, t.Server)
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	if err := serverCheckers.MakesureVolumeExist(attachment); err != nil {
		return err
	}
	return nil
}

type ServerDetachVolume struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerDetachVolume) Start() error {
	t.RefreshServer()
	attachment, err := t.lastVolume()
	if err != nil {
		return err
	}
	err = t.Client.NovaV2().Server().DeleteVolume(t.Server.Id, attachment.VolumeId)
	if err != nil {
		return err
	}
	logging.Info("[%s] detaching volume %s", t.Server.Id, attachment.VolumeId)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if err := t.ServerMustNotError(); err != nil {
		return err
	}
	serverCheckers, err := checkers.GetServerCheckers(t.Client, t.Server)
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	if err := serverCheckers.MakesureVolumeNotExists(attachment); err != nil {
		return err
	}
	return nil
}

type ServerVolumeHotPlug struct {
	ServerActionTest
	EmptyCleanup
	attachments []string
}

func (t *ServerVolumeHotPlug) Start() error {
	t.RefreshServer()
	serverCheckers, err := checkers.GetServerCheckers(t.Client, t.Server)
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	for i := 0; i < common.TASK_CONF.VolumeHotplug.Nums; i++ {
		logging.Info("[%s] attach volume (%d)", t.ServerId(), i+1)

		logging.Info("[%s] creating volume", t.ServerId())
		volume, err := t.CreateBlankVolume()
		if err != nil {
			return fmt.Errorf("create volume failed: %s", err)
		}

		attachment, err := t.Client.NovaV2().Server().AddVolume(t.Server.Id, volume.Id)
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
		if err := serverCheckers.MakesureVolumeExist(attachment); err != nil {
			return err
		}
		t.attachments = append(t.attachments, volume.Id)
	}

	for _, volId := range t.attachments {
		err := t.Client.NovaV2().Server().DeleteVolume(t.Server.Id, volId)
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

func (t ServerVolumeHotPlug) TearDown() error {
	deleteFailed := []string{}
	logging.Info("[%s] cleanup %d volumes", t.ServerId(), len(t.attachments))
	for _, volId := range t.attachments {
		logging.Info("[%s] deleting volume %s", t.ServerId(), volId)
		err := t.Client.CinderV2().Volume().Delete(volId, true, true)
		if err != nil {
			logging.Error("[%s] delete volume %s failed: %s", t.ServerId(), volId, err)
			deleteFailed = append(deleteFailed, volId)
		}
	}
	if len(deleteFailed) > 0 {
		return fmt.Errorf("delete volume(s) %s failed", strings.Join(deleteFailed, ","))
	} else {
		return nil
	}
}
