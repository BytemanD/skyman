package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model/neutron"
)

type ServerAttachPort struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerAttachPort) Start() error {
	if t.Server.IsError() {
		return fmt.Errorf("server is not active")
	}
	nextNetwork, err := t.nextNetwork()
	if err != nil {
		return err
	}
	console.Info("[%s] creating port", t.Server.Id)
	port, err := t.Client.NeutronV2().CreatePort(map[string]any{
		"network_id": nextNetwork,
	})
	if err != nil {
		console.Error("[%s] create port failed: %s", t.ServerId(), err)
		return err
	}

	console.Info("[%s] attaching port: %s", t.Server.Id, port.Id)
	attachment, err := t.Client.NovaV2().ServerAddInterface(t.Server.Id, "", port.Id)
	if err != nil {
		return err
	}
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if err := t.ServerMustNotError(); err != nil {
		return err
	}
	serverCheckers, err := t.getCheckers()
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	if err := serverCheckers.MakesureInterfaceExist(attachment); err != nil {
		return err
	}
	return nil
}

type ServerAttachNet struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerAttachNet) Start() error {
	if t.Server.IsError() {
		return fmt.Errorf("server is not active")
	}
	nextNetwork, err := t.nextNetwork()
	if err != nil {
		return err
	}

	console.Info("[%s] attaching net: %s", t.Server.Id, nextNetwork)
	attachment, err := t.Client.NovaV2().ServerAddInterface(t.Server.Id, nextNetwork, "")
	if err != nil {
		return err
	}
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if err := t.ServerMustNotError(); err != nil {
		return err
	}
	serverCheckers, err := t.getCheckers()
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	if err := serverCheckers.MakesureInterfaceExist(attachment); err != nil {
		return err
	}
	return nil
}

type ServerDetachInterface struct {
	ServerActionTest
	EmptyCleanup
}

func (t *ServerDetachInterface) lastInterface() (string, error) {
	interfaces, err := t.Client.NovaV2().ListServerInterfaces(t.Server.Id)
	if err != nil {
		return "", err
	}
	if len(interfaces) == 0 {
		return "", fmt.Errorf("has no interface")
	}
	return interfaces[len(interfaces)-1].PortId, nil
}

func (t ServerDetachInterface) Start() error {
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	portId, err := t.lastInterface()
	if err != nil {
		return err
	}
	port, err := t.Client.NeutronV2().GetPort(portId)
	if err != nil {
		return err
	}
	err = t.Client.NovaV2().DeleteServerInterfaceAndWait(t.Server.Id, portId, time.Minute*5)
	if err != nil {
		return err
	}
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if err := t.ServerMustNotError(); err != nil {
		return err
	}
	serverCheckers, err := t.getCheckers()
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}

	if err := serverCheckers.MakesureInterfaceNotExists(port); err != nil {
		return err
	}
	return nil
}

type ServerAttachHotPlug struct {
	ServerActionTest
	EmptyCleanup
	attachedPort []*neutron.Port
}

func (t *ServerAttachHotPlug) Skip() (bool, string) {
	if !t.Server.IsActive() {
		return true, "server is not active"
	}
	return false, ""
}

func (t *ServerAttachHotPlug) Start() error {
	serverCheckers, err := t.getCheckers()
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	for i := 1; i <= t.Config.InterfaceHotplug.Nums; i++ {
		console.Info("[%s] attach interface %d", t.ServerId(), i)
		nextNetwork, err := t.nextNetwork()
		if err != nil {
			return err
		}
		console.Info("[%s] creating port", t.Server.Id)
		port, err := t.Client.NeutronV2().CreatePort(map[string]any{
			"network_id": nextNetwork,
		})
		if err != nil {
			console.Error("[%s] create port failed: %s", t.ServerId(), err)
			return err
		}

		console.Info("[%s] attaching interface %s", t.Server.Id, port.Id)
		attachment, err := t.Client.NovaV2().ServerAddInterface(t.Server.Id, "", port.Id)
		if err != nil {
			return err
		}

		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if err := t.ServerMustNotError(); err != nil {
			return err
		}
		if err := serverCheckers.MakesureInterfaceExist(attachment); err != nil {
			return err
		}
		t.attachedPort = append(t.attachedPort, port)
	}

	for _, port := range t.attachedPort {
		err = t.Client.NovaV2().DeleteServerInterfaceAndWait(t.Server.Id, port.Id, time.Minute*5)
		if err != nil {
			return err
		}
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if err := t.ServerMustNotError(); err != nil {
			return err
		}
		if err := serverCheckers.MakesureInterfaceNotExists(port); err != nil {
			return err
		}
	}
	return nil
}
func (t ServerAttachHotPlug) TearDown() error {
	deleteFailed := []string{}
	console.Info("[%s] cleanup %d interfaces", t.ServerId(), len(t.attachedPort))
	for _, port := range t.attachedPort {
		console.Info("[%s] deleting port %s", t.ServerId(), port.Id)
		err := t.Client.NeutronV2().DeletePort(port.Id)
		if err != nil {
			deleteFailed = append(deleteFailed, port.Id)
			console.Error("[%s] delete port %s failed: %s", t.ServerId(), port.Id, err)
		}
	}
	if len(deleteFailed) > 0 {
		return fmt.Errorf("delete port(s) %s failed", strings.Join(deleteFailed, ","))
	} else {
		return nil
	}
}
