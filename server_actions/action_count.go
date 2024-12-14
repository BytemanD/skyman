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

func (acl *ActionCountList) Add(name string, count ...int) {
	addCount := 0
	if len(count) > 0 {
		addCount = count[0]
	} else {
		addCount = 1
	}
	if acl.Empty() {
		acl.items = append(acl.items, ActionCount{Name: name, Count: addCount})
	} else {
		lastAc := acl.Last()
		if lastAc.Name == name {
			lastAc.Count += 1
		} else {
			acl.items = append(acl.items, ActionCount{Name: name, Count: addCount})
		}
	}
}

func (acl ActionCountList) Actions() []string {
	actions := make([]string, 0, len(acl.items))
	for _, ac := range acl.items {
		actions = append(actions, ac.Name)
	}
	return actions
}
func (acl ActionCountList) FormatActions() []string {
	actions := make([]string, 0, len(acl.items))
	for _, ac := range acl.items {
		actions = append(actions, fmt.Sprintf("%s:%d", ac.Name, ac.Count))
	}
	return actions
}
func NewActionCountList(actions string) (*ActionCountList, error) {
	actionCountList := &ActionCountList{}

	for _, item := range strings.Split(actions, ",") {
		action := strings.TrimSpace(item)
		if !strings.Contains(action, ":") {
			if !internal.VALID_ACTIONS.Contains(action) {
				return nil, fmt.Errorf("action '%s' not found", action)
			}
			actionCountList.Add(action)
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
		actionCountList.Add(splited[0], count)
	}
	return actionCountList, nil
}
