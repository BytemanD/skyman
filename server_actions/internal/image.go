package internal

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type ServerImage struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerImage) Start() error {
	// TODO
	err := t.Client.NovaV2().Server().Pause(t.Server.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] pausing", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if !t.Server.IsPaused() {
		return fmt.Errorf("server is not active")
	}
	return nil
}
