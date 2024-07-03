package server_actions

import (
	"fmt"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/checkers"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/neutron"
)

type ServerAttachInterface struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerAttachInterface) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	nextNetwork, err := t.nextNetwork()
	if err != nil {
		return err
	}
	logging.Info("[%s] creating port", t.Server.Id)
	port, err := t.Client.NeutronV2().Ports().Create(map[string]interface{}{
		"network_id": nextNetwork,
	})
	if err != nil {
		logging.Error("[%s] create port failed: %s", t.ServerId(), err)
		return err
	}

	logging.Info("[%s] attaching interface: %s", t.Server.Id, port.Id)
	attachment, err := t.Client.NovaV2().Servers().AddInterface(t.Server.Id, "", port.Id)
	if err != nil {
		return err
	}
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if err := t.ServerMustNotError(); err != nil {
		return err
	}
	serverCheckers, err := checkers.GetServerCheckers(t.Client, t.Server)
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
	interfaces, err := t.Client.NovaV2().Servers().ListInterfaces(t.Server.Id)
	if err != nil {
		return "", err
	}
	if len(interfaces) == 0 {
		return "", fmt.Errorf("has no interface")
	}
	return interfaces[len(interfaces)-1].PortId, nil
}

func (t ServerDetachInterface) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	portId, err := t.lastInterface()
	if err != nil {
		return err
	}
	port, err := t.Client.NeutronV2().Ports().Show(portId)
	if err != nil {
		return err
	}
	err = t.Client.NovaV2().Servers().DeleteInterfaceAndWait(t.Server.Id, portId, time.Minute*5)
	if err != nil {
		return err
	}
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if err := t.ServerMustNotError(); err != nil {
		return err
	}
	serverCheckers, err := checkers.GetServerCheckers(t.Client, t.Server)
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
	attachedPorts []*neutron.Port
}

func (t *ServerAttachHotPlug) Start() error {
	t.RefreshServer()
	if !t.Server.IsActive() {
		return fmt.Errorf("server is not active")
	}
	serverCheckers, err := checkers.GetServerCheckers(t.Client, t.Server)
	if err != nil {
		return fmt.Errorf("get server checker failed: %s", err)
	}
	for i := 1; i <= common.CONF.Test.InterfaceHotplug.Nums; i++ {
		logging.Info("[%s] attach interface %d", t.ServerId(), i)
		nextNetwork, err := t.nextNetwork()
		if err != nil {
			return err
		}
		logging.Info("[%s] creating port", t.Server.Id)
		port, err := t.Client.NeutronV2().Ports().Create(map[string]interface{}{
			"network_id": nextNetwork,
		})
		if err != nil {
			logging.Error("[%s] create port failed: %s", t.ServerId(), err)
			return err
		}

		logging.Info("[%s] attaching interface %s", t.Server.Id, port.Id)
		attachment, err := t.Client.NovaV2().Servers().AddInterface(t.Server.Id, "", port.Id)
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
		t.attachedPorts = append(t.attachedPorts, port)
	}

	for _, port := range t.attachedPorts {
		err = t.Client.NovaV2().Servers().DeleteInterfaceAndWait(t.Server.Id, port.Id, time.Minute*5)
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
func (t ServerAttachHotPlug) Cleanup() {
	for _, port := range t.attachedPorts {
		logging.Info("[%s] cleanup %d interfaces", t.ServerId(), len(t.attachedPorts))

		logging.Info("[%s] deleting port %s", t.ServerId(), port.Id)
		err := t.Client.NeutronV2().Ports().Delete(port.Id)
		if err != nil {
			logging.Error("[%s] delete port %s failed", t.ServerId(), port.Id)
		}
	}
}
