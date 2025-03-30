package internal

import (
	"fmt"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

type ServerRebuild struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerRebuild) Start() error {
	err := t.Client.NovaV2().RebuildServer(t.Server.Id, nova.RebuilOpt{})
	if err != nil {
		return err
	}
	console.Info("[%s] rebuilding", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	return nil
}
