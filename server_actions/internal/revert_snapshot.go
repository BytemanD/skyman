package internal

import "github.com/BytemanD/easygo/pkg/global/logging"

type ServerRevertToSnapshot struct {
	ServerActionTest
	EmptyCleanup
	srcStatus string
}

func (t *ServerRevertToSnapshot) Start() error {
	t.srcStatus = t.Server.Status
	if t.Server.IsActive() {
		logging.Info("[%s] stopping", t.ServerId())
		t.Client.NovaV2().Server().StopAndWait(t.ServerId())
	}
	rootBdm, err := t.GetRootVolume()
	if err != nil {
		return err
	}
	snap, err := t.Client.CinderV2().Snapshot().Create(rootBdm.VolumeId, "skyman-snap", true)
	if err != nil {
		return err
	}
	logging.Info("[%s] creating snapshot, waiting", t.ServerId())
	if err := t.WaitSnapshotCreated(snap.Id); err != nil {
		return err
	}
	logging.Info("[%s] snapshot created", t.ServerId())
	logging.Info("[%s] revert volume to snapshot %s, waiting", t.ServerId(), rootBdm.VolumeId)
	if err := t.Client.CinderV2().Volume().Revert(rootBdm.VolumeId, snap.Id); err != nil {
		return err
	}
	if err := t.WaitVolumeTaskDone(rootBdm.VolumeId); err != nil {
		return err
	}

	logging.Info("[%s] revert to snapshot success", t.ServerId())
	return nil
}

func (t ServerRevertToSnapshot) TearDown() error {
	if t.srcStatus == "ACTIVE" {
		logging.Info("[%s] starting", t.ServerId())
		if err := t.Client.NovaV2().Server().Start(t.ServerId()); err != nil {
			return err
		}
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
	}
	return nil
}
