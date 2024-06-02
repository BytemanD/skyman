package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type ServerRebuild struct{ ServerActionTest }

func (t ServerRebuild) Start() error {
	err := t.Client.NovaV2().Servers().Rebuild(t.Server.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] rebuilding", t.Server.Id)
	if err := t.WaitServerTaskFinished(true); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	return nil
}
