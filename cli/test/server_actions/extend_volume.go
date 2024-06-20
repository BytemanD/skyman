package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/checkers"
)

type ServerExtendVolume struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerExtendVolume) Start() error {
	t.RefreshServer()
	attachment, err := t.lastVolume()
	if err != nil {
		return err
	}

	volume, err := t.Client.CinderV2().Volumes().Show(attachment.VolumeId)
	if err != nil {
		return fmt.Errorf("get volume failed: %s", err)
	}
	newSize := volume.Size + 10
	logging.Info("[%s] extending volume size %s to %dG", t.Server.Id, attachment.VolumeId, newSize)
	err = t.Client.CinderV2().Volumes().Extend(attachment.VolumeId, int(newSize))
	if err != nil {
		return err
	}

	serverCheckers, err := checkers.GetServerCheckers(t.Client, t.Server)
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	if err := serverCheckers.MakesureVolumeSizeIs(attachment, newSize); err != nil {
		return err
	}
	return nil
}
