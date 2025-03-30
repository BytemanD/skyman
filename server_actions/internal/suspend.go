package internal

import (
	"fmt"

	"github.com/BytemanD/go-console/console"
)

type ServerSuspend struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerSuspend) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	err := t.Client.NovaV2().SuspendServer(t.Server.Id)
	if err != nil {
		return err
	}
	console.Info("[%s] suspending", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if !t.Server.IsSuspended() {
		return fmt.Errorf("server is not suspended")
	}
	return nil
}

type ServerResume struct {
	ServerActionTest
	EmptyCleanup
}

func (t *ServerResume) Skip() (bool, string) {
	if !t.Server.IsSuspended() {
		return true, "server is not suspended"
	}
	return false, ""
}
func (t ServerResume) Start() error {
	err := t.Client.NovaV2().ResumeServer(t.Server.Id)
	if err != nil {
		return err
	}
	console.Info("[%s] resuming", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	return nil
}

type ServerToggleSuspend struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerToggleSuspend) Start() error {
	if t.Server.IsSuspended() {
		err := t.Client.NovaV2().ResumeServer(t.Server.Id)
		if err != nil {
			return err
		}
		console.Info("[%s] resuming", t.Server.Id)
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if !t.Server.IsActive() {
			return fmt.Errorf("server is not active")
		}
	} else if t.Server.IsActive() {
		err := t.Client.NovaV2().SuspendServer(t.Server.Id)
		if err != nil {
			return err
		}
		console.Info("[%s] suspending", t.Server.Id)
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
