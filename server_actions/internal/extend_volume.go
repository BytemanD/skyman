package internal

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/utility"
)

type ServerExtendVolume struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerExtendVolume) Start() error {
	attachment, err := t.lastVolume()
	if err != nil {
		return err
	}

	volume, err := t.Client.CinderV2().Volume().Show(attachment.VolumeId)
	if err != nil {
		return fmt.Errorf("get volume failed: %s", err)
	}
	newSize := volume.Size + 10
	err = t.Client.CinderV2().Volume().Extend(attachment.VolumeId, int(newSize))
	logging.Info("[%s] extending volume size %s to %dG", t.Server.Id, attachment.VolumeId, newSize)
	if err != nil {
		return err
	}
	utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:      60 * 2,
			IntervalMin:  time.Second,
			IntervalStep: time.Second,
			IntervalMax:  time.Second * 5},
		[]string{"VolumeHasTaskError"},
		func() error {
			vol, err := t.Client.CinderV2().Volume().Show(attachment.VolumeId)
			if err != nil {
				return err
			}
			logging.Info("[%s] volume %s task state is %s",
				t.ServerId(), attachment.VolumeId, vol.TaskStatus)
			if vol.TaskStatus == "" {
				return nil
			}
			return utility.NewVolumeHasTaskError(attachment.VolumeId)
		},
	)

	serverCheckers, err := t.getCheckers()
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	if err := serverCheckers.MakesureVolumeSizeIs(attachment, newSize); err != nil {
		return err
	}
	return nil
}
