package internal

import (
	"fmt"
	"time"

	"github.com/BytemanD/go-console/console"
)

type ServerSetName struct {
	ServerActionTest
	EmptyCleanup
}

func (t *ServerSetName) Skip() (bool, string) {
	if t.Server.IsShelved() {
		return true, "server is shelved"
	}
	return false, ""
}
func (t ServerSetName) Start() error {
	console.Info("[%s] old name is %s", t.Server.Id, t.Server.Name)
	newName := time.Now().Format(time.DateTime)
	console.Info("[%s] set name to %s", t.Server.Id, newName)
	err := t.Client.NovaV2().ServerSetName(t.Server.Id, newName)
	if err != nil {
		return err
	}
	if t.Server.Name == newName {
		console.Info("[%s] name is %s", t.Server.Id, t.Server.Name)
		return nil
	} else {
		return fmt.Errorf("server name is not %s", newName)
	}
}
