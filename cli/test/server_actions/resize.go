package server_actions

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

var TEST_FLAVORS = []nova.Flavor{}

type ServerResize struct {
	ServerActionTest
	EmptyCleanup
}

func (t ServerResize) nextFlavor() (*nova.Flavor, error) {
	if len(TEST_FLAVORS) <= 1 {
		return nil, fmt.Errorf("the num of flavors <= 1")
	}
	flavorIndex := -1
	for i, flavor := range TEST_FLAVORS {
		if flavor.Name == t.Server.Flavor.OriginalName {
			flavorIndex = i
			break
		}
	}
	if flavorIndex == -1 || flavorIndex+1 >= len(TEST_FLAVORS) {
		flavorIndex = 0
	} else {
		flavorIndex++
	}
	return &TEST_FLAVORS[flavorIndex], nil
}

func (t ServerResize) Start() error {
	nextFlavor, err := t.nextFlavor()
	if err != nil {
		return err
	}

	logging.Info("[%s] resize %s -> %s", t.Server.Id, t.Server.Flavor.OriginalName, nextFlavor.Name)

	err = t.Client.NovaV2().Server().Resize(t.Server.Id, nextFlavor.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] resizing", t.Server.Id)
	if err := t.WaitServerTaskFinished(false); err != nil {
		return err
	}
	if t.Server.IsError() {
		return fmt.Errorf("server is error")
	}
	if t.Server.Flavor.OriginalName != nextFlavor.Name {
		return fmt.Errorf("sever flavor is not %s", nextFlavor.Name)
	}
	return nil
}
