package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
)

type ServerAttachInterface struct {
	ServerActionTest
	networkIndex int
}

func (t *ServerAttachInterface) nextNetwork() (string, error) {
	if len(common.CONF.Test.Networks) == 0 {
		return "", fmt.Errorf("the num of networks == 0")
	}
	if t.networkIndex >= len(common.CONF.Test.Networks)-1 {
		t.networkIndex = 0
	}
	defer func() { t.networkIndex += 1 }()
	return common.CONF.Test.Networks[t.networkIndex], nil
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
	logging.Info("[%s] attaching interface", t.Server.Id)
	attachment, err := t.Client.NovaV2().Servers().AddInterface(t.Server.Id, nextNetwork, "")
	if err != nil {
		return err
	}
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
		if vif.PortId == attachment.PortId {
			return nil
		}
	}
	return fmt.Errorf("server has no interface %s", attachment.PortId)
}

type ServerDetachInterface struct {
	ServerActionTest
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
