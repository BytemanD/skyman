package server_actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

var (
	ACTION_REBOOT          = "reboot"
	ACTION_HARD_REBOOT     = "hard_reboot"
	ACTION_LIVE_MIGRATE    = "live_migrate"
	ACTION_SHELVE          = "shelve"
	ACTION_UNSHELVE        = "unshelve"
	ACTION_TOGGLE_UNSHELVE = "toggle_shelve"
)

func GetActions() []string {
	return []string{
		ACTION_REBOOT, ACTION_HARD_REBOOT,
		ACTION_LIVE_MIGRATE,
		ACTION_SHELVE, ACTION_UNSHELVE, ACTION_TOGGLE_UNSHELVE,
	}
}
func GetTestAction(actionName string, server *nova.Server, client *openstack.Openstack) (ServerAction, error) {
	testBase := ServerActionTest{Server: server, Client: client}
	switch actionName {
	case ACTION_REBOOT:
		return ServerReboot{ServerActionTest: testBase}, nil
	case ACTION_HARD_REBOOT:
		return ServerHardReboot{ServerActionTest: testBase}, nil
	case ACTION_LIVE_MIGRATE:
		return ServerLiveMigrate{ServerActionTest: testBase}, nil
	case ACTION_SHELVE:
		return ServerShelve{ServerActionTest: testBase}, nil
	case ACTION_UNSHELVE:
		return ServerUnshelve{ServerActionTest: testBase}, nil
	case ACTION_TOGGLE_UNSHELVE:
		return ServerToggleShelve{ServerActionTest: testBase}, nil
	}
	return nil, fmt.Errorf("action '%s' not found", actionName)
}

func ParseServerActions(actions string) ([]string, error) {
	serverActions := []string{}
	if actions == "" {
		return serverActions, nil
	}
	for _, action := range strings.Split(actions, ",") {
		if !strings.Contains(action, ":") {
			serverActions = append(serverActions, action)
			continue
		}
		splited := strings.Split(action, ":")
		count, err := strconv.Atoi(splited[1])
		if err != nil {
			return nil, fmt.Errorf("invalid action '%s'", action)
		}
		for i := 0; i < count; i++ {
			serverActions = append(serverActions, splited[0])
		}
	}
	return serverActions, nil
}
