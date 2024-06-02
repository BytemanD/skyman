package server_actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var (
	ACTION_REBOOT           = "reboot"
	ACTION_HARD_REBOOT      = "hard_reboot"
	ACTION_STOP             = "stop"
	ACTION_START            = "start"
	ACTION_PAUSE            = "pause"
	ACTION_UNPAUSE          = "unpause"
	ACTION_MIGRATE          = "migrate"
	ACTION_LIVE_MIGRATE     = "live_migrate"
	ACTION_SHELVE           = "shelve"
	ACTION_UNSHELVE         = "unshelve"
	ACTION_TOGGLE_SHELVE    = "toggle_shelve"
	ACTION_REBUILD          = "rebuild"
	ACTION_RESIZE           = "resize"
	ACTION_RENAME           = "rename"
	ACTION_SUSPEND          = "suspend"
	ACTION_RESUME           = "resume"
	ACTION_TOGGLE_SUSPEND   = "toggle_suspend"
	ACTION_ATTACH_INTERFACE = "attach_interface"
	ACTION_DETACH_INTERFACE = "detach_interface"
	ACTION_ATTACH_VOLUME    = "attach_volume"
	ACTION_DETACH_VOLUME    = "detach_volume"
)

func GetActions() []string {
	return []string{
		ACTION_REBOOT, ACTION_HARD_REBOOT, ACTION_START, ACTION_STOP,
		ACTION_LIVE_MIGRATE, ACTION_MIGRATE,
		ACTION_RESIZE, ACTION_REBUILD,
		ACTION_SHELVE, ACTION_UNSHELVE, ACTION_TOGGLE_SHELVE,
		ACTION_RENAME,
		ACTION_PAUSE, ACTION_UNPAUSE,
		ACTION_SUSPEND, ACTION_RESUME, ACTION_TOGGLE_SUSPEND,
		ACTION_ATTACH_INTERFACE, ACTION_DETACH_INTERFACE,
		ACTION_ATTACH_VOLUME, ACTION_DETACH_VOLUME,
	}
}
func GetTestAction(actionName string, server *nova.Server, client *openstack.Openstack) (ServerAction, error) {
	testBase := ServerActionTest{Server: server, Client: client}
	switch actionName {
	case ACTION_REBOOT:
		return ServerReboot{ServerActionTest: testBase}, nil
	case ACTION_HARD_REBOOT:
		return ServerHardReboot{ServerActionTest: testBase}, nil
	case ACTION_START:
		return ServerStart{ServerActionTest: testBase}, nil
	case ACTION_STOP:
		return ServerStop{ServerActionTest: testBase}, nil
	case ACTION_MIGRATE:
		return ServerMigrate{ServerActionTest: testBase}, nil
	case ACTION_LIVE_MIGRATE:
		return ServerLiveMigrate{ServerActionTest: testBase}, nil
	case ACTION_SHELVE:
		return ServerShelve{ServerActionTest: testBase}, nil
	case ACTION_UNSHELVE:
		return ServerUnshelve{ServerActionTest: testBase}, nil
	case ACTION_TOGGLE_SHELVE:
		return ServerToggleShelve{ServerActionTest: testBase}, nil
	case ACTION_REBUILD:
		return ServerRebuild{ServerActionTest: testBase}, nil
	case ACTION_RESIZE:
		return ServerResize{ServerActionTest: testBase}, nil
	case ACTION_RENAME:
		return ServerRename{ServerActionTest: testBase}, nil
	case ACTION_SUSPEND:
		return ServerSuspend{ServerActionTest: testBase}, nil
	case ACTION_RESUME:
		return ServerResume{ServerActionTest: testBase}, nil
	case ACTION_TOGGLE_SUSPEND:
		return ServerToggleSuspend{ServerActionTest: testBase}, nil
	case ACTION_ATTACH_INTERFACE:
		return ServerAttachInterface{ServerActionTest: testBase}, nil
	case ACTION_DETACH_INTERFACE:
		return ServerDetachInterface{ServerActionTest: testBase}, nil
	case ACTION_ATTACH_VOLUME:
		return ServerAttachVolume{ServerActionTest: testBase}, nil
	case ACTION_DETACH_VOLUME:
		return ServerDetachVolume{ServerActionTest: testBase}, nil
	default:
		return nil, fmt.Errorf("action '%s' not found", actionName)
	}
}

func ParseServerActions(actions string) ([]string, error) {
	serverActions := []string{}
	if actions == "" {
		return serverActions, nil
	}
	for _, action := range strings.Split(actions, ",") {
		if !strings.Contains(action, ":") {
			if !utility.StringsContains(action, GetActions()) {
				return nil, fmt.Errorf("invalid action '%s'", action)
			}
			serverActions = append(serverActions, action)
			continue
		}
		splited := strings.Split(action, ":")
		count, err := strconv.Atoi(splited[1])
		if err != nil {
			return nil, fmt.Errorf("invalid action '%s'", action)
		}
		if !utility.StringsContains(splited[0], GetActions()) {
			return nil, fmt.Errorf("invalid action '%s'", splited[0])
		}
		for i := 0; i < count; i++ {
			serverActions = append(serverActions, splited[0])
		}
	}
	return serverActions, nil
}
