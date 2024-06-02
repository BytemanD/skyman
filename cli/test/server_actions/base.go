package server_actions

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

type ServerAction interface {
	Start() error
	// Rollback()
	Cleanup()
}
type ServerActionTest struct {
	Server *nova.Server
	Client *openstack.Openstack
}

func (t *ServerActionTest) RefreshServer() error {
	server, err := t.Client.NovaV2().Servers().Show(t.Server.Id)
	if err != nil {
		return nil
	}
	t.Server = server
	return nil
}

func (t *ServerActionTest) WaitServerTaskFinished(showProgress bool) error {
	interval, maxInterval := 1, 10

	for i := 0; i <= 60; i++ {
		if err := t.RefreshServer(); err != nil {
			return err
		}
		progress := ""
		if showProgress {
			progress = fmt.Sprintf(", progress: %d", int(t.Server.Progress))
		}
		logging.Info("[%s] vmState=%s, powerState=%s, taskState=%s%s",
			t.Server.Id, t.Server.VmState, t.Server.GetPowerState(), t.Server.TaskState, progress)
		if t.Server.TaskState == "" {
			return nil
		}
		time.Sleep(time.Second * time.Duration(interval))
		if interval < maxInterval {
			interval += 1
		}
	}
	return fmt.Errorf("server task state is %s", t.Server.TaskState)
}

func (t ServerActionTest) Cleanup() {
	logging.Info("[%s] clean up", t.Server.Id)
}
