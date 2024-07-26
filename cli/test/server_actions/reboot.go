package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/checkers"
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
	err := t.Client.NovaV2().Server().Reboot(t.Server.Id, false)
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
	err := t.Client.NovaV2().Server().Reboot(t.Server.Id, true)
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
	return t.MakesureServerRunning()
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
	err := t.Client.NovaV2().Server().Stop(t.Server.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] stopping", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}

	serverCheckers, err := checkers.GetServerCheckers(t.Client, t.Server)
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	if err := serverCheckers.MakesureServerStopped(); err != nil {
		return err
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
	err := t.Client.NovaV2().Server().Start(t.Server.Id)
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
