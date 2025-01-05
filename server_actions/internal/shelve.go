package internal

import (
	"fmt"

	"github.com/BytemanD/go-console/console"
)

type ServerShelve struct {
	ServerActionTest
	EmptyCleanup
}

func (t *ServerShelve) Skip() (bool, string) {
	if t.Server.IsShelved() {
		return true, "server is shelved"
	}
	return false, ""
}
func (t ServerShelve) Start() error {
	err := t.Client.NovaV2().Server().Shelve(t.Server.Id)
	if err != nil {
		return err
	}
	console.Info("[%s] shelving", t.Server.Id)

	if err := t.WaitServerTaskFinished(false); err != nil {
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

func (t *ServerUnshelve) Skip() (bool, string) {
	if !t.Server.IsShelved() {
		return true, "server is not shelved"
	}
	return false, ""
}
func (t ServerUnshelve) Start() error {
	err := t.Client.NovaV2().Server().Unshelve(t.Server.Id)
	if err != nil {
		return err
	}
	console.Info("[%s] unshelving", t.Server.Id)

	if err := t.WaitServerTaskFinished(false); err != nil {
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
		err := t.Client.NovaV2().Server().Unshelve(t.Server.Id)
		if err != nil {
			return err
		}
		console.Info("[%s] unshelving", t.Server.Id)
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if !t.Server.IsActive() {
			return fmt.Errorf("server is not active")
		}
		return nil
	} else {
		err := t.Client.NovaV2().Server().Shelve(t.Server.Id)
		if err != nil {
			return err
		}
		console.Info("[%s] shelving", t.Server.Id)
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if !t.Server.IsShelved() {
			return fmt.Errorf("server is not shelved")
		}
		return nil
	}
}
