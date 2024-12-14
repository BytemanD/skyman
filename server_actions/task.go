package server_actions

import (
	"fmt"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/checkers"
	"github.com/BytemanD/skyman/cli/test/server_actions"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/server_actions/internal"
	"github.com/BytemanD/skyman/utility"
)

type ServerActionTest struct {
	TestId       int
	Worker       int
	Server       *nova.Server
	Client       *openstack.Openstack
	Conf         common.ServerActionsTestConf
	Actions      ActionCountList
	networkIndex int
	Report       TestTask
}

func (t ServerActionTest) getServerBootOption() nova.ServerOpt {
	opt := nova.ServerOpt{
		Name:             fmt.Sprintf("skyman-server-%d-%v", t.TestId, time.Now().Format("20060102-150405")),
		Flavor:           t.Conf.Flavors[0],
		Image:            common.TASK_CONF.Images[0],
		AvailabilityZone: common.TASK_CONF.AvailabilityZone,
	}
	if len(common.TASK_CONF.Networks) >= 1 {
		opt.Networks = []nova.ServerOptNetwork{
			{UUID: common.TASK_CONF.Networks[0]},
		}
	} else {
		logging.Warning("boot without network")
	}
	if common.TASK_CONF.BootWithSG != "" {
		opt.SecurityGroups = append(opt.SecurityGroups,
			neutron.SecurityGroup{
				Resource: model.Resource{Name: common.TASK_CONF.BootWithSG},
			})
	}
	if common.TASK_CONF.BootFromVolume {
		opt.BlockDeviceMappingV2 = []nova.BlockDeviceMappingV2{
			{
				UUID:               common.TASK_CONF.Images[0],
				VolumeSize:         common.TASK_CONF.BootVolumeSize,
				SourceType:         "image",
				DestinationType:    "volume",
				VolumeType:         common.TASK_CONF.BootVolumeType,
				DeleteOnTemination: true,
			},
		}
	} else {
		opt.Image = common.TASK_CONF.Images[0]
	}
	return opt
}
func (t ServerActionTest) createDefaultServer() (*nova.Server, error) {
	bootOption := t.getServerBootOption()
	return t.Client.NovaV2().Server().Create(bootOption)
}
func (t ServerActionTest) waitServerCreated() error {
	var err error
	server, err = t.Client.NovaV2().Server().Show(t.Server.Id)
	if err != nil {
		return err
	}
	logging.Info("[%s] creating with name %s", server.Id, server.Resource.Name)
	server, err = client.NovaV2().Server().WaitBooted(server.Id)
	if err != nil {
		return err
	}
	logging.Success("[%s] create success, host is %s", server.Id, server.Host)
	return nil
}
func (t *ServerActionTest) testActions(acl ActionCountList) error {
	if t.Server == nil {
		if len(t.Conf.Flavors) == 0 {
			logging.Error("test flavors is empty")
			return fmt.Errorf("test flavors is empty")
		}
		if len(common.TASK_CONF.Images) == 0 {
			return fmt.Errorf("test images is empty")
		}
		t.Report.SetStage("creating")
		server, err := t.createDefaultServer()
		if err != nil {
			t.Report.MarkFailed(fmt.Sprintf("create server failed: %s", err))
			return t.Report.GetError()
		}
		t.Server = server
		if err := t.waitServerCreated(); err != nil {
			t.Report.MarkFailed(fmt.Sprintf("create failed: %s", err))
			return t.Report.GetError()
		}
	}
	logging.Info("[%s] ======== Test actions: %s", t.Server.Id, strings.Join(acl.FormatActions(), ","))
	for _, actionName := range acl.Actions() {
		logging.Info(utility.BlueString(fmt.Sprintf("[%s] ==== %s", t.Server.Id, actionName)))
		t.Report.SetStage(actionName)
	}

}
func (t *ServerActionTest) Start() error {
	if len(t.Conf.ActionTasks) == 0 {
		logging.Warning("action is empty")
		return nil
	}
	var (
		server *nova.Server
		err    error
	)
	// success, failed := 0, 0
	task := server_actions.TestTask{
		Id:           testId,
		TotalActions: TestServerActions,
	}
	if serverId != "" {
		server, err = client.NovaV2().Server().Found(serverId)
		if err != nil {
			return fmt.Errorf("get server failed: %s", err)
		}
		task.ServerId = serverId
		server_actions.TestTasks = append(server_actions.TestTasks, &task)
		logging.Info("use server: %s(%s)", server.Id, server.Name)
		logging.Info("[%s] ======== %s ========", serverId, strings.Join(TestServerActions, ","))

	} else {
		if len(server_actions.TEST_FLAVORS) == 0 {
			logging.Error("test flavors is empty")
			return fmt.Errorf("test flavors is empty")
		}
		if len(common.TASK_CONF.Images) == 0 {
			return fmt.Errorf("test images is empty")
		}
		server, err = createDefaultServer(client, testId)
		if err != nil {
			task.MarkFailed(fmt.Sprintf("create server failed: %s", err))
			return task.GetError()
		}
		logging.Info("[%s] ======== Test actions: %s", server.Id, strings.Join(TestServerActions, ","))
		task.SetStage("creating")
		task.ServerId = server.Id
		server_actions.TestTasks = append(server_actions.TestTasks, &task)
		if err := waitServerCreated(client, server); err != nil {
			task.MarkFailed(fmt.Sprintf("create failed: %s", err))
			return task.GetError()
		}
		server, err = client.NovaV2().Server().Show(server.Id)
		if err != nil {
			task.MarkFailed(fmt.Sprintf("get server failed: %s", err))
			return task.GetError()
		}
		defer func() {
			if (!task.HasFailed() && common.TASK_CONF.DeleteIfSuccess) || (task.HasFailed() && common.TASK_CONF.DeleteIfError) {
				task.SetStage("deleting")
				logging.Info("[%s] deleting server", server.Id)
				client.NovaV2().Server().Delete(server.Id)
				client.NovaV2().Server().WaitDeleted(server.Id)
			}
			task.SetStage("")
		}()
		serverCheckers, err := checkers.GetServerCheckers(client, server)
		if err != nil {
			task.MarkFailed(fmt.Sprintf("get server checker failed: %s", err))
			return task.GetError()
		}
		if err := serverCheckers.MakesureServerRunning(); err != nil {
			return err
		}
	}
	for _, actionName := range TestServerActions {
		action := server_actions.VALID_ACTIONS.Get(actionName, server, client)
		if action == nil {
			logging.Error("[%s] action '%s' not found", server.Id, action)
			task.SkipActions = append(task.SkipActions, actionName)
			continue
		}
		logging.Info(utility.BlueString(fmt.Sprintf("[%s] ==== %s", server.Id, actionName)))
		task.SetStage(actionName)
		// 更新实例信息
		if err := action.RefreshServer(); err != nil {
			return fmt.Errorf("refresh server failed: %s", err)
		}
		if actionInterval > 0 {
			logging.Info("[%s] sleep %d seconds", server.Id, actionInterval)
			time.Sleep(time.Second * time.Duration(actionInterval))
		}
		// 开始测试
		skip, err := runActionTest(action)
		// 更新测试结果
		task.IncrementCompleted()
		if skip {
			task.SkipActions = append(task.SkipActions, actionName)
		} else if err != nil {
			task.FailedActions = append(task.FailedActions, actionName)
			task.MarkFailed(fmt.Sprintf("test action '%s' failed: %s", actionName, err))
			logging.Error("[%s] %s", server.Id, task.GetError())
		} else {
			task.SuccessActions = append(task.SuccessActions, actionName)
			logging.Success("[%s] test action '%s' success", server.Id, actionName)
		}
	}

	if task.AllSuccess() {
		task.MarkSuccess()
		logging.Success("[%s] %s", server.Id, task.GetResultString())
	} else if task.HasFailed() {
		logging.Error("[%s] %s", server.Id, task.GetResultString())
	} else {
		task.MarkWarning()
		logging.Warning("[%s] %s", server.Id, task.GetResultString())
	}
	return nil
}
func RunActionTest(action internal.ServerAction) (bool, error) {
	skip, reason := false, ""
	if skip, reason = action.Skip(); skip {
		logging.Warning("[%s] skip this task for the reason: %s", action.ServerId(), reason)
		return true, nil
	}
	defer func() {
		logging.Info("[%s] >>>> tear down ...", action.ServerId())
		if err := action.TearDown(); err != nil {
			logging.Error("[%s] tear down failed: %s", action.ServerId(), err)
		}
	}()
	return false, action.Start()
}
