package server_actions

import (
	"fmt"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/arrayutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"

	"github.com/BytemanD/skyman/cli/test/checkers"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/server_actions/internal"
	"github.com/BytemanD/skyman/utility"
)

func startAction(action internal.ServerAction) (bool, error) {
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

type Case struct {
	Actions        ActionCountList
	Workers        int
	ActionInterval int
	UseServers     []string
	Client         *openstack.Openstack
	Conf           common.Case
	reports        []TestTask
}

func (t Case) getServerBootOption(testId int) nova.ServerOpt {
	opt := nova.ServerOpt{
		Name:             fmt.Sprintf("skyman-server-%d-%v", testId, time.Now().Format("20060102-150405")),
		Flavor:           t.firstFlavor(),
		Image:            t.firstImage(),
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
func (t Case) createDefaultServer(testId int) (*nova.Server, error) {
	bootOption := t.getServerBootOption(testId)
	return t.Client.NovaV2().Server().Create(bootOption)
}
func (t Case) waitServerCreated(serverId string) error {
	var err error
	server, err := t.Client.NovaV2().Server().Show(serverId)
	if err != nil {
		return err
	}
	logging.Info("[%s] creating with name %s", server.Id, server.Resource.Name)
	server, err = t.Client.NovaV2().Server().WaitBooted(server.Id)
	if err != nil {
		return err
	}
	logging.Success("[%s] create success, host is %s", server.Id, server.Host)
	return nil
}
func (t Case) firstFlavor() string {
	if len(t.Conf.Flavors) > 0 {
		return t.Conf.Flavors[0]
	}
	if len(common.TASK_CONF.Flavors[0]) > 0 {
		return common.TASK_CONF.Flavors[0]
	}
	return ""
}
func (t Case) firstImage() string {
	if len(t.Conf.Images) > 0 {
		return t.Conf.Images[0]
	}
	if len(common.TASK_CONF.Images[0]) > 0 {
		return common.TASK_CONF.Images[0]
	}
	return ""
}
func (t *Case) destroyServer(serverId string) {
	logging.Info("[%s] deleting server", serverId)
	t.Client.NovaV2().Server().Delete(serverId)
	t.Client.NovaV2().Server().WaitDeleted(serverId)
}
func (t *Case) testActions(testId int, serverId string) (report *TestTask) {
	// 执行一轮测试
	report = &TestTask{
		TotalActions: t.Actions.FormatActions(),
	}
	var server *nova.Server
	// defer func() {
	// 	if report.AllSuccess() {
	// 		report.MarkSuccess()
	// 		logging.Success("[%s] %s", report.ServerId, report.GetResultString())
	// 	} else if report.HasFailed() {
	// 		logging.Error("[%s] %s", report.ServerId, report.GetResultString())
	// 	} else {
	// 		report.MarkWarning()
	// 		logging.Warning("[%s] %s", report.ServerId, report.GetResultString())
	// 	}
	// }()
	logging.Info("run case with worker: %d, actions: %s", t.Workers, t.Actions.FormatActions())
	if serverId == "" {
		if t.firstFlavor() == "" {
			report.MarkFailed("test flavors is empty", nil)
			return
		}
		if t.firstImage() == "" {
			report.MarkFailed("test images is empty", nil)
			return
		}
		report.SetStage("creating")
		server, err := t.createDefaultServer(testId)
		if err != nil {
			report.MarkFailed("create server failed", err)
			return
		}
		if err := t.waitServerCreated(server.Id); err != nil {
			report.MarkFailed("create server failed", err)
			return
		}
		server, err = t.Client.NovaV2().Server().Show(server.Id)
		if err != nil {
			report.MarkFailed("get server failed", err)
			return
		}
		serverCheckers, err := checkers.GetServerCheckers(t.Client, server)
		if err != nil {
			report.MarkFailed("get server checker failed", err)
			return
		}
		if err := serverCheckers.MakesureServerRunning(); err != nil {
			report.MarkFailed("get server checker failed", err)
			return
		}
		report.ServerId = server.Id
		defer func() {
			if (!report.HasFailed() && common.TASK_CONF.DeleteIfSuccess) ||
				(report.HasFailed() && common.TASK_CONF.DeleteIfError) {
				report.SetStage("deleting")
				t.destroyServer(server.Id)
			}
			report.SetStage("")
		}()
	} else {
		var err error
		server, err = t.Client.NovaV2().Server().Found(serverId)
		if err != nil {
			report.MarkFailed("get server failed", err)
			return
		}

		report.ServerId = server.Id
		logging.Info("use server: %s(%s)", server.Id, server.Name)
	}
	logging.Info("[%s] ======== Test actions: %s", report.ServerId, strings.Join(t.Actions.FormatActions(), ","))
	for _, actionName := range t.Actions.Actions() {
		action := internal.VALID_ACTIONS.Get(actionName, server, t.Client)
		if action == nil {
			logging.Error("[%s] action '%s' not found", server.Id, actionName)
			report.SkipActions = append(report.SkipActions, actionName)
			continue
		}
		logging.Info(utility.BlueString(fmt.Sprintf("[%s] ==== %s", server.Id, actionName)))
		report.SetStage(actionName)
		// 更新实例信息
		if err := action.RefreshServer(); err != nil {
			report.MarkFailed("refresh server failed", err)
			return
		}
		if t.ActionInterval > 0 {
			logging.Info("[%s] sleep %d seconds", server.Id, t.ActionInterval)
			time.Sleep(time.Second * time.Duration(t.ActionInterval))
		}
		// 开始测试
		skip, err := startAction(action)
		// 更新测试结果
		report.IncrementCompleted()
		if skip {
			report.SkipActions = append(report.SkipActions, actionName)
		} else if err != nil {
			report.FailedActions = append(report.FailedActions, actionName)
			report.MarkFailed(fmt.Sprintf("test action '%s' failed", actionName), err)
			logging.Error("[%s] %s", server.Id, report.GetError())
		} else {
			report.SuccessActions = append(report.SuccessActions, actionName)
			logging.Success("[%s] test action '%s' success", server.Id, actionName)
		}
	}
	return
}
func (t *Case) PrintReport() {
	PrintTestTasks(t.reports)
}
func (t *Case) PrintServerEvents() error {
	return nil
}
func (t *Case) Start() error {
	if t.Actions.Empty() {
		logging.Warning("action is empty")
		return nil
	}
	t.Workers = max(t.Workers, 1)
	// success, failed := 0, 0
	taskGroup := syncutils.TaskGroup{
		MaxWorker: t.Workers,
	}
	t.reports = []TestTask{}
	if len(t.UseServers) > 0 {
		logging.Warning("use servers: %s", t.UseServers)
		taskGroup.MaxWorker = len(t.UseServers)
		taskGroup.Items = arrayutils.Range(1, len(t.UseServers)+1)
		taskGroup.Func = func(item interface{}) error {
			testId := item.(int)
			report := t.testActions(testId, t.UseServers[testId-1])
			t.reports = append(t.reports, *report)
			return nil
		}
	} else {
		taskGroup.MaxWorker = max(t.Workers, 1)
		taskGroup.Items = arrayutils.Range(1, t.Workers+1)
		taskGroup.Func = func(item interface{}) error {
			testId := item.(int)
			report := t.testActions(testId, "")
			t.reports = append(t.reports, *report)
			return nil
		}
	}
	taskGroup.Start()
	return nil
}
