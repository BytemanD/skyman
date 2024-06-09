package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type ServerShelve struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerShelve) Start() error {
	t.RefreshServer()
	if t.Server.IsShelved() {
		return fmt.Errorf("server is shelved")
	}
	err := t.Client.NovaV2().Servers().Shelve(t.Server.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] shelving", t.Server.Id)

	if err := t.WaitServerTaskFinished(true); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	if !t.Server.IsShelved() {
		return fmt.Errorf("server is not shelved")
	}
	return nil
}

type ServerUnshelve struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerUnshelve) Start() error {
	t.RefreshServer()
	if !t.Server.IsShelved() {
		return fmt.Errorf("server is not shelved")
	}
	err := t.Client.NovaV2().Servers().Unshelve(t.Server.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] unshelving", t.Server.Id)

	if err := t.WaitServerTaskFinished(true); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	return nil
}

type ServerToggleShelve struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerToggleShelve) Start() error {
	t.RefreshServer()
	if t.Server.IsShelved() {
		err := t.Client.NovaV2().Servers().Unshelve(t.Server.Id)
		if err != nil {
			return err
		}
		logging.Info("[%s] unshelving", t.Server.Id)
		if err := t.WaitServerTaskFinished(true); err != nil {
			return err
		}
		if !t.Server.IsActive() {
			return fmt.Errorf("server is not active")
		}
		return nil
	} else {
		err := t.Client.NovaV2().Servers().Shelve(t.Server.Id)
		if err != nil {
			return err
		}
		logging.Info("[%s] shelving", t.Server.Id)
		if err := t.WaitServerTaskFinished(true); err != nil {
			return err
		}
		if !t.Server.IsShelved() {
			return fmt.Errorf("server is not shelved")
		}
		return nil
	}
}
