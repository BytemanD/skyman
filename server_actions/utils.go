package server_actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BytemanD/skyman/server_actions/internal"
)

func ParseServerActions(actions string) ([]string, error) {
	serverActions := []string{}
	if actions == "" {
		return serverActions, nil
	}
	for _, act := range strings.Split(actions, ",") {
		action := strings.TrimSpace(act)
		if !strings.Contains(action, ":") {
			if !internal.VALID_ACTIONS.Contains(action) {
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
		if !internal.VALID_ACTIONS.Contains(splited[0]) {
			return nil, fmt.Errorf("action '%s' not found", splited[0])
		}
		for i := 0; i < count; i++ {
			serverActions = append(serverActions, splited[0])
		}
	}
	return serverActions, nil
}

func ValidActions() internal.Actions {
	return internal.VALID_ACTIONS
}
