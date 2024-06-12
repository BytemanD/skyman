package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/checkers"
	"github.com/BytemanD/skyman/common"
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
	if err := t.ServerMustHasInterface(port.Id); err != nil {
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

	err = t.Client.NovaV2().Servers().DeleteInterface(t.Server.Id, portId)
	if err != nil {
		return err
	}
	logging.Info("[%s] detaching interface %s", t.Server.Id, portId)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server status is error")
	}
	interfaces, err := t.Client.NovaV2().Servers().ListInterfaces(t.Server.Id)
	if err != nil {
		return err
	}
	for _, vif := range interfaces {
		if vif.PortId == portId {
			return fmt.Errorf("interface %s is not detached", portId)
		}
	}
	return nil
}

type ServerAttachHotPlug struct {
	ServerActionTest
	attachments []string
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
		t.attachments = append(t.attachments, port.Id)
	}

	for _, portId := range t.attachments {
		err := t.Client.NovaV2().Servers().DeleteInterface(t.Server.Id, portId)
		if err != nil {
			return err
		}
		logging.Info("[%s] detaching interface %s", t.ServerId(), portId)
		if err := t.WaitServerTaskFinished(false); err != nil {
			return err
		}
		if err := t.ServerMustNotError(); err != nil {
			return err
		}
		if err := t.ServerMustHasNotInterface(portId); err != nil {
			return err
		}
	}
	return nil
}
func (t ServerAttachHotPlug) Cleanup() {
	for _, portId := range t.attachments {
		logging.Info("[%s] cleanup %d interfaces", t.ServerId(), len(t.attachments))

		logging.Info("[%s] deleting port %s", t.ServerId(), portId)
		err := t.Client.NeutronV2().Ports().Delete(portId)
		if err != nil {
			logging.Error("[%s] delete port %s failed", t.ServerId(), portId)
		}
	}
}
