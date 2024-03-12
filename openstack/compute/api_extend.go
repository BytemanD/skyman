package compute

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/jedib0t/go-pretty/v6/progress"
)

func (client ComputeClientV2) ServerPrune(query url.Values, yes bool, waitDeleted bool) {
	if len(query) == 0 {
		query.Set("status", "error")
	}
	logging.Info("查询虚拟机: %v", query.Encode())
	servers, err := client.ServerList(query)
	common.LogError(err, "query servers failed", true)
	logging.Info("需要清理的虚拟机数量: %d\n", len(servers))
	if len(servers) == 0 {
		return
	}
	if !yes {
		var confirm string
		fmt.Println("即将删除虚拟机:")
		for _, server := range servers {
			fmt.Printf("%s (%s)\n", server.Id, server.Name)
		}
		for {
			fmt.Printf("是否删除[y(es)/n(o)]: ")
			fmt.Scanln(&confirm)
			if confirm == "yes" || confirm == "y" {
				break
			} else if confirm == "no" || confirm == "n" {
				return
			} else {
				fmt.Print("输入错误, 请重新输入!")
			}
		}
	}
	wg := sync.WaitGroup{}
	workers := make(chan struct{}, 1)
	wg.Add(len(servers))
	pw := progress.NewWriter()
	tracker := progress.Tracker{Total: int64(len(servers))}
	pw.AppendTracker(&tracker)
	pw.SetAutoStop(true)

	go pw.Render()

	deleteFunc := func(s Server) {
		tracker.UpdateMessage(fmt.Sprintf("Increment %d", tracker.Value()))
		client.ServerDelete(s.Id)
		if waitDeleted {
			client.WaitServerDeleted(s.Id)
		}
		<-workers
		tracker.Increment(1)
		if tracker.Value() >= tracker.Total {
			tracker.MarkAsDone()
		}
	}

	logging.Info("开始删除虚拟机")
	for _, server := range servers {
		workers <- struct{}{}
		go deleteFunc(server)
	}
	for !tracker.IsDone() {

	}
	if tracker.IsDone() {
		tracker.Increment(1)
		logging.Info("")
		// pw.Stop()
	}
	wg.Wait()
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
