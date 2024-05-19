package server_actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

func GetTestAction(actionName string, server *nova.Server, client *openstack.Openstack) (ServerAction, error) {
	testBase := ServerActionTest{Server: server, Client: client}
	switch actionName {
	case "reboot":
		return ServerReboot{ServerActionTest: testBase}, nil
	case "hard_reboot":
		return ServerHardReboot{ServerActionTest: testBase}, nil
	case "live_migrate":
		return ServerLiveMigrate{ServerActionTest: testBase}, nil
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
