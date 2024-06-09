package server_actions

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type ServerLiveMigrate struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerLiveMigrate) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}

	sourceHost := t.Server.Host
	logging.Info("[%s] source host is %s", t.Server.Id, sourceHost)
	startTime := time.Now()

	err := t.Client.NovaV2().Servers().LiveMigrate(t.Server.Id, "auto", "")
	if err != nil {
		return err
	}
	logging.Info("[%s] live migrating", t.Server.Id)

	if err := t.WaitServerTaskFinished(true); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	if t.Server.Host == sourceHost {
		return fmt.Errorf("server host not changed")
	}
	logging.Info("[%s] migrated, %s -> %s, used: %v",
		t.Server.Id, sourceHost, t.Server.Host, time.Since(startTime))
	return nil
}

type ServerMigrate struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerMigrate) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}

	sourceHost := t.Server.Host
	startTime := time.Now()
	logging.Info("[%s] source host is %s", t.Server.Id, sourceHost)

	err := t.Client.NovaV2().Servers().Migrate(t.Server.Id, "")
	if err != nil {
		return err
	}
	logging.Info("[%s] migrating", t.Server.Id)

	if err := t.WaitServerTaskFinished(true); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	if t.Server.Host == sourceHost {
		return fmt.Errorf("server host not changed")
	}
	logging.Info("[%s] migrated, %s -> %s, used: %v",
		t.Server.Id, sourceHost, t.Server.Host, time.Since(startTime))
	return nil
}
