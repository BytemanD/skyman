package server_actions

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type ServerRename struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerRename) Start() error {
	t.RefreshServer()
	if t.Server.IsShelved() {
		return fmt.Errorf("server is shelved")
	}
	newName := fmt.Sprintf("server-%v", time.Now())
	logging.Info("[%s] old name is %s", t.Server.Id, t.Server.Name)
	err := t.Client.NovaV2().Server().SetName(t.Server.Id, newName)
	if err != nil {
		return err
	}
	logging.Info("[%s] set name to %s", t.Server.Id, t.Server.Name)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if t.Server.Name == newName {
		logging.Info("[%s] name is %s", t.Server.Id, t.Server.Name)
		return nil
	} else {
		return fmt.Errorf("server name is not %s", newName)
	}
}
