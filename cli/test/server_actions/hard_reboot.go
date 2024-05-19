package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type ServerHardReboot struct{ ServerActionTest }

func (t ServerHardReboot) Start() error {
	err := t.Client.NovaV2().Servers().Reboot(t.Server.Id, true)
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
