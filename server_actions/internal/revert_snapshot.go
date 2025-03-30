package internal

import (
	"fmt"

	"github.com/BytemanD/go-console/console"
)

type ServerRevertToSnapshot struct {
	ServerActionTest
	EmptyCleanup
	srcStatus string
}

func (t *ServerRevertToSnapshot) Start() error {
	t.srcStatus = t.Server.Status
	if t.Server.IsStopped() {
		console.Info("[%s] starting", t.ServerId())
		t.Client.NovaV2().StartServer(t.ServerId())
		if _, err := t.Client.NovaV2().WaitServerStatus(t.ServerId(), "ACTIVE", 2); err != nil {
			return fmt.Errorf("start server %s failed", t.ServerId())
		}
	}
	rootBdm, err := t.GetRootVolume()
	if err != nil {
		return fmt.Errorf("get root volume failed: %s", err)
	}
	snap, err := t.Client.CinderV2().CreateSnapshot(rootBdm.VolumeId, "skyman-snap", true)
	if err != nil {
		return err
	}
	console.Info("[%s] creating snapshot %s, waiting", t.ServerId(), snap.Id)
	if err := t.WaitSnapshotCreated(snap.Id); err != nil {
		return err
	}
	console.Info("[%s] snapshot %s created", t.ServerId(), snap.Id)
	t.RefreshServer()
	if t.Server.IsActive() {
		console.Info("[%s] server is active, stop before reversing", t.ServerId())
		if err := t.Client.NovaV2().StopServerAndWait(t.ServerId()); err != nil {
			console.Info("[%s] stopped", t.ServerId())
		}
	}
	for i := range max(t.Config.RevertSystem.RepeatEveryTime, 1) {
		console.Info("[%s] revert volume to snapshot %s (%d), waiting", t.ServerId(), rootBdm.VolumeId, i+1)
		if err := t.Client.CinderV2().RevertVolume(rootBdm.VolumeId, snap.Id); err != nil {
			console.Error("revert volume %s failed: %s", rootBdm.VolumeId, err)
			return err
		}
		if err := t.WaitVolumeTaskDone(rootBdm.VolumeId); err != nil {
			return err
		}
		console.Success("[%s] revert to snapshot %s success", t.ServerId(), snap.Id)
	}
	console.Info("[%s] starting", t.ServerId())
	if err := t.Client.NovaV2().StartServer(t.ServerId()); err != nil {
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
