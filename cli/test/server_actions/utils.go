package server_actions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

var (
	ACTION_REBOOT            = "reboot"
	ACTION_HARD_REBOOT       = "hard_reboot"
	ACTION_STOP              = "stop"
	ACTION_START             = "start"
	ACTION_PAUSE             = "pause"
	ACTION_UNPAUSE           = "unpause"
	ACTION_MIGRATE           = "migrate"
	ACTION_LIVE_MIGRATE      = "live_migrate"
	ACTION_SHELVE            = "shelve"
	ACTION_UNSHELVE          = "unshelve"
	ACTION_TOGGLE_SHELVE     = "toggle_shelve"
	ACTION_REBUILD           = "rebuild"
	ACTION_RESIZE            = "resize"
	ACTION_RENAME            = "rename"
	ACTION_SUSPEND           = "suspend"
	ACTION_RESUME            = "resume"
	ACTION_TOGGLE_SUSPEND    = "toggle_suspend"
	ACTION_ATTACH_NET        = "net_attach"
	ACTION_ATTACH_PORT       = "port_attach"
	ACTION_DETACH_PORT       = "port_detach"
	ACTION_INTERFACE_HOTPLUG = "interface_hotplug"
	ACTION_ATTACH_VOLUME     = "volume_attach"
	ACTION_DETACH_VOLUME     = "volume_detach"
	ACTION_VOLUME_HOTPLUG    = "volume_hotplug"
	ACTION_VOLUME_EXTEND     = "volume_extend"
	ACTION_REVERT_SYSTEM     = "revert_system"
	ACTION_NOP               = "nop"
)

type ActionCreatorFunc func(server *nova.Server, client *openstack.Openstack) ServerAction

type Actions map[string]ActionCreatorFunc

func (a Actions) Keys() []string {
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func (a Actions) Contains(name string) bool {
	for k := range a {
		if k == name {
			return true
		}
	}
	return false
}
func (a Actions) Get(name string, server *nova.Server, client *openstack.Openstack) ServerAction {
	if creator, ok := a[name]; ok {
		return creator(server, client)
	}
	return nil
}

func (a Actions) register(name string, creator ActionCreatorFunc) {
	if _, ok := a[name]; ok {
		logging.Warning("action %s already exists", name)
		return
	}
	a[name] = creator
}

var VALID_ACTIONS = Actions{}

func ParseServerActions(actions string) ([]string, error) {
	serverActions := []string{}
	if actions == "" {
		return serverActions, nil
	}
	for _, act := range strings.Split(actions, ",") {
		action := strings.TrimSpace(act)
		if !strings.Contains(action, ":") {
			if !VALID_ACTIONS.Contains(action) {
				return nil, fmt.Errorf("action '%s' not found", action)
			}
			serverActions = append(serverActions, action)
			continue
		}
		splited := strings.Split(action, ":")
		count, err := strconv.Atoi(splited[1])
		if err != nil {
			return nil, fmt.Errorf("invalid action '%s'", action)
		}
		if !VALID_ACTIONS.Contains(splited[0]) {
			return nil, fmt.Errorf("action '%s' not found", splited[0])
		}
		for i := 0; i < count; i++ {
			serverActions = append(serverActions, splited[0])
		}
	}
	return serverActions, nil
}

func init() {
	VALID_ACTIONS.register(ACTION_REBOOT, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerReboot{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_HARD_REBOOT, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerHardReboot{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_STOP, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerStop{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_START, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerStart{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_PAUSE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerPause{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_UNPAUSE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerUnpause{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_MIGRATE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerMigrate{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_LIVE_MIGRATE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerLiveMigrate{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_SHELVE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerShelve{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_UNSHELVE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerUnshelve{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_TOGGLE_SHELVE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerToggleShelve{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_REBUILD, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerRebuild{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_RESIZE, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerResize{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_RENAME, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerRename{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_SUSPEND, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerSuspend{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_RESUME, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerResume{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_TOGGLE_SUSPEND, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerToggleSuspend{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_ATTACH_PORT, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerAttachPort{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_ATTACH_NET, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerAttachNet{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_DETACH_PORT, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerDetachInterface{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_ATTACH_VOLUME, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerAttachVolume{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_DETACH_VOLUME, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerDetachVolume{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_INTERFACE_HOTPLUG, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerAttachHotPlug{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_VOLUME_HOTPLUG, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerVolumeHotPlug{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_VOLUME_EXTEND, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerExtendVolume{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_REVERT_SYSTEM, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerRevertToSnapshot{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
	VALID_ACTIONS.register(ACTION_NOP, func(s *nova.Server, c *openstack.Openstack) ServerAction {
		return &ServerActionNop{ServerActionTest: ServerActionTest{Server: s, Client: c}}
	})
}
