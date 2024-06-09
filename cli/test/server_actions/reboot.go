package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type ServerReboot struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerReboot) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	err := t.Client.NovaV2().Servers().Reboot(t.Server.Id, false)
	if err != nil {
		return err
	}
	logging.Info("[%s] rebooting", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	return nil
}

type ServerHardReboot struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerHardReboot) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	err := t.Client.NovaV2().Servers().Reboot(t.Server.Id, true)
	if err != nil {
		return err
	}
	logging.Info("[%s] hard rebooting", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	return nil
}

type ServerStop struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerStop) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	err := t.Client.NovaV2().Servers().Stop(t.Server.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] stopping", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if !t.Server.IsStopped() {
		return fmt.Errorf("server is not stopped")
	}
	return nil
}

type ServerStart struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerStart) Start() error {
	t.RefreshServer()
	if !t.Server.IsStopped() {
		return fmt.Errorf("server is not stopped")
	}
	err := t.Client.NovaV2().Servers().Start(t.Server.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] starting", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	return nil
}
