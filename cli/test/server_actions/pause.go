package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type ServerPause struct{ ServerActionTest }

func (t ServerPause) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	err := t.Client.NovaV2().Servers().Pause(t.Server.Id)
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

type ServerUnpause struct{ ServerActionTest }

func (t ServerUnpause) Start() error {
	t.RefreshServer()
	if !t.Server.IsPaused() {
		return fmt.Errorf("server is not paused")
	}
	err := t.Client.NovaV2().Servers().Unpause(t.Server.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] unpausing", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	return nil
}

type ServerTogglePause struct{ ServerActionTest }

func (t ServerTogglePause) Start() error {
	t.RefreshServer()
	if t.Server.IsPaused() {
		err := t.Client.NovaV2().Servers().Unpause(t.Server.Id)
		if err != nil {
			return err
		}
		logging.Info("[%s] unpausing", t.Server.Id)
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if !t.Server.IsActive() {
			return fmt.Errorf("server is not active")
		}
	} else if t.Server.IsActive() {
		err := t.Client.NovaV2().Servers().Pause(t.Server.Id)
		if err != nil {
			return err
		}
		logging.Info("[%s] pausing", t.Server.Id)
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if !t.Server.IsPaused() {
			return fmt.Errorf("server is not paused")
		}
	} else {
		return fmt.Errorf("skip server status is %s", t.Server.Status)
	}
	return nil
}
