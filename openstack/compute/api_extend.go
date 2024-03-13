package compute

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/utility"
)

func (client ComputeClientV2) ServerPrune(query url.Values, yes bool, waitDeleted bool) {
	if len(query) == 0 {
		query.Set("status", "error")
	}
	logging.Info("查询虚拟机: %v", query.Encode())
	servers, err := client.ServerList(query)
	utility.LogError(err, "query servers failed", true)
	logging.Info("需要清理的虚拟机数量: %d\n", len(servers))
	if len(servers) == 0 {
		return
	}
	if !yes {
		fmt.Printf("即将删除 %d 个虚拟机:\n", len(servers))
		for _, server := range servers {
			fmt.Printf("    %s(%s)\n", server.Id, server.Name)
		}
		yes = utility.ScanfComfirm("是否删除", []string{"yes", "y"}, []string{"no", "n"})
	}
	if !yes {
		return
	}
	logging.Info("开始删除虚拟机")
	tg := utility.TaskGroup{
		Items: servers,
		Func: func(o interface{}) error {
			s := o.(Server)
			fmt.Println("delete", s.Id)
			err := client.ServerDelete(s.Id)
			if err != nil {
				return fmt.Errorf("delete %s failed: %v", s.Id, err)
			}
			if waitDeleted {
				client.WaitServerDeleted(s.Id)
			}
			return nil
		},
		ShowProgress: true,
	}
	tg.Start()
}

func (client ComputeClientV2) FlavorCopy(flavorId string, newName string, newId string,
	newVcpus int, newRam int, newDisk int, newSwap int,
	newEphemeral int, newRxtxFactor float32, setProperties map[string]string,
	unsetProperties []string) (*Flavor, error) {

	logging.Info("Show flavor")
	flavor, err := client.FlavorShow(flavorId)
	if err != nil {
		return nil, err
	}
	flavor.Name = newName
	flavor.Id = newId
	if newVcpus != 0 {
		flavor.Vcpus = newVcpus
	}
	if newRam != 0 {
		flavor.Ram = int(newRam)
	}
	if newDisk != 0 {
		flavor.Disk = newDisk
	}
	if newSwap != 0 {
		flavor.Swap = newSwap
	}
	if newEphemeral != 0 {
		flavor.Ephemeral = newEphemeral
	}
	if newRxtxFactor != 0 {
		flavor.RXTXFactor = newRxtxFactor
	}
	logging.Info("Show flavor extra specs")
	extraSpecs, err := client.FlavorExtraSpecsList(flavorId)
	if err != nil {
		return nil, err
	}
	for k, v := range setProperties {
		extraSpecs[k] = v
	}
	for _, k := range unsetProperties {
		delete(extraSpecs, k)
	}
	logging.Info("Create new flavor")
	newFlavor, err := client.FlavorCreate(*flavor)
	if err != nil {
		return nil, fmt.Errorf("create flavor failed, %v", err)
	}
	if len(extraSpecs) != 0 {
		logging.Info("Set new flavor extra specs")
		_, err = client.FlavorExtraSpecsCreate(newFlavor.Id, extraSpecs)
		if err != nil {
			return nil, fmt.Errorf("set flavor extra specs failed, %v", err)
		}
		newFlavor.ExtraSpecs = extraSpecs
	}

	return newFlavor, nil
}

func (client ComputeClientV2) ServerActionsWithEvents(id string, actionName string, requestId string) (
	[]InstanceActionAndEvents, error,
) {
	serverActions, err := client.ServerActionList(id)
	if err != nil {
		return nil, err
	}
	var actionEvents []InstanceActionAndEvents
	for _, action := range serverActions {
		if requestId != "" && action.RequestId != requestId {
			continue
		}
		if actionName != "" && action.Action != actionName {
			continue
		}
		serverAction, err := client.ServerActionShow(id, action.RequestId)
		if err != nil {
			logging.Error("get server action %s failed, %s", action.RequestId, err)
		}
		actionEvents = append(actionEvents, InstanceActionAndEvents{
			InstanceAction: action,
			RequestId:      action.RequestId,
			Events:         serverAction.Events,
		})
	}
	return actionEvents, nil
}
