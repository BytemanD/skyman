package internal

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type ServerRevertToSnapshot struct {
	ServerActionTest
	EmptyCleanup
	srcStatus string
}

func (t *ServerRevertToSnapshot) Start() error {
	t.srcStatus = t.Server.Status
	if t.Server.IsStopped() {
		logging.Info("[%s] starting", t.ServerId())
		t.Client.NovaV2().Server().Start(t.ServerId())
		if _, err := t.Client.NovaV2().Server().WaitStatus(t.ServerId(), "ACTIVE", 2); err != nil {
			return fmt.Errorf("start server %s failed", t.ServerId())
		}
	}
	rootBdm, err := t.GetRootVolume()
	if err != nil {
		return fmt.Errorf("get root volume failed: %s", err)
	}
	snap, err := t.Client.CinderV2().Snapshot().Create(rootBdm.VolumeId, "skyman-snap", true)
	if err != nil {
		return err
	}
	logging.Info("[%s] creating snapshot %s, waiting", t.ServerId(), snap.Id)
	if err := t.WaitSnapshotCreated(snap.Id); err != nil {
		return err
	}
	logging.Info("[%s] snapshot %s created", t.ServerId(), snap.Id)
	t.RefreshServer()
	if t.Server.IsActive() {
		logging.Info("[%s] server is active, stop before reversing", t.ServerId())
		if err := t.Client.NovaV2().Server().StopAndWait(t.ServerId()); err != nil {
			logging.Info("[%s] stopped", t.ServerId())
		}
	}
	for i := 0; i < max(t.Config.RevertSystem.RepeatEveryTime, 1); i++ {
		logging.Info("[%s] revert volume to snapshot %s (%d), waiting", t.ServerId(), rootBdm.VolumeId, i+1)
		if err := t.Client.CinderV2().Volume().Revert(rootBdm.VolumeId, snap.Id); err != nil {
			logging.Error("revert volume %s failed: %s", rootBdm.VolumeId, err)
			return err
		}
		if err := t.WaitVolumeTaskDone(rootBdm.VolumeId); err != nil {
			return err
		}
		logging.Success("[%s] revert to snapshot %s success", t.ServerId(), snap.Id)
	}
	logging.Info("[%s] starting", t.ServerId())
	if err := t.Client.NovaV2().Server().Start(t.ServerId()); err != nil {
		return err
	}
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	serverCheckers, err := t.getCheckers()
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	if err := serverCheckers.MakesureServerRunning(); err != nil {
		return err
	}
	return nil
}

func (t ServerRevertToSnapshot) TearDown() error {
	return nil
}
