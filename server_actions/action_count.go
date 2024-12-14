package server_actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BytemanD/skyman/server_actions/internal"
)

type ActionCount struct {
	Name  string
	Count int
}

type ActionCountList struct {
	items []ActionCount
}

func (acl *ActionCountList) Items() []ActionCount {
	return acl.items
}
func (acl *ActionCountList) Empty() bool {
	return len(acl.items) <= 0
}

func (acl *ActionCountList) Last() *ActionCount {
	if acl.Empty() {
		return nil
	}
	return &acl.items[len(acl.items)-1]
}

func (acl *ActionCountList) Add(name string) {
	if acl.Empty() {
		acl.items = append(acl.items, ActionCount{Name: name, Count: 1})
	} else {
		lastAc := acl.Last()
		if lastAc.Name == name {
			lastAc.Count += 1
		} else {
			acl.items = append(acl.items, ActionCount{Name: name, Count: 1})
		}
	}
}
func (acl ActionCountList) Actions() []string {
	actions := []string{}
	for _, ac := range acl.items {
		actions = append(actions, ac.Name)
	}
	return actions
}
func (acl ActionCountList) FormatActions() []string {
	actions := []string{}
	for _, ac := range acl.items {
		actions = append(actions, fmt.Sprintf("%s:%d", ac.Name, ac.Count))
	}
	return actions
}
func NewActionCountList(actions string) (*ActionCountList, error) {
	actionCountList := &ActionCountList{}
	serverActions := []string{}
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
	return actionCountList, nil
}
