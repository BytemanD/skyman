package internal

import (
	"fmt"
	"time"

	"github.com/BytemanD/go-console/console"
)

type ServerRename struct {
	ServerActionTest
	EmptyCleanup
}

func (t *ServerRename) Skip() (bool, string) {
	if t.Server.IsShelved() {
		return true, "server is shelved"
	}
	return false, ""
}
func (t ServerRename) Start() error {
	newName := fmt.Sprintf("server-%v", time.Now())
	console.Info("[%s] old name is %s", t.Server.Id, t.Server.Name)
	err := t.Client.NovaV2().Server().SetName(t.Server.Id, newName)
	if err != nil {
		return err
	}
	console.Info("[%s] set name to %s", t.Server.Id, t.Server.Name)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if t.Server.Name == newName {
		console.Info("[%s] name is %s", t.Server.Id, t.Server.Name)
		return nil
	} else {
		return fmt.Errorf("server name is not %s", newName)
	}
}
