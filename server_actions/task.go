package server_actions

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/go-console/console"
	"github.com/samber/lo"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/server_actions/checkers"
	"github.com/BytemanD/skyman/server_actions/internal"
	"github.com/BytemanD/skyman/utility"
)

func startAction(action internal.ServerAction) error {
	defer func() {
		console.Info("[%s] >>>> tear down ...", action.ServerId())
		if err := action.TearDown(); err != nil {
			console.Error("[%s] tear down failed: %s", action.ServerId(), err)
		}
	}()
	return action.Start()
}

type Case struct {
	Name       string
	Actions    ActionCountList
	UseServers []string
	Client     *openstack.Openstack
	Config     common.CaseConfig

	caseReport CaseReport
	reportLock *sync.Mutex
}

func (t Case) getServerBootOption(testId int) nova.ServerOpt {
	opt := nova.ServerOpt{
		Name:             fmt.Sprintf("skyman-server-%d-%v", testId, time.Now().Format("20060102-150405")),
		Flavor:           t.firstFlavor(),
		Image:            t.firstImage(),
		AvailabilityZone: t.Config.AvailabilityZone,
	}
	if len(t.Config.Networks) >= 1 {
		opt.Networks = []nova.ServerOptNetwork{
			{UUID: t.Config.Networks[0]},
		}
	} else {
		console.Warn("boot without network")
	}
	if t.Config.BootWithSG != "" {
		opt.SecurityGroups = append(opt.SecurityGroups,
			neutron.SecurityGroup{
				Resource: model.Resource{Name: t.Config.BootWithSG},
			})
	}
	if t.Config.BootFromVolume {
		opt.BlockDeviceMappingV2 = []nova.BlockDeviceMappingV2{
			{
				UUID:               t.Config.Images[0],
				VolumeSize:         t.Config.BootVolumeSize,
				SourceType:         "image",
				DestinationType:    "volume",
				VolumeType:         t.Config.BootVolumeType,
				DeleteOnTemination: true,
			},
		}
	} else {
		opt.Image = t.Config.Images[0]
	}
	return opt
}
func (t Case) createDefaultServer(testId int) (*nova.Server, error) {
	bootOption := t.getServerBootOption(testId)
	return t.Client.NovaV2().CreateServer(bootOption)
}
func (t Case) waitServerCreated(serverId string) error {
	var err error
	server, err := t.Client.NovaV2().GetServer(serverId)
	if err != nil {
		return err
	}
	console.Info("[%s] creating with name %s", server.Id, server.Resource.Name)
	server, err = t.Client.NovaV2().WaitServerBooted(server.Id)
	if err != nil {
		return err
	}
	console.Success("[%s] create success, host is %s", server.Id, server.Host)
	return nil
}
func (t Case) firstFlavor() string {
	if len(t.Config.Flavors) > 0 {
		return t.Config.Flavors[0]
	}
	return ""
}
func (t Case) firstImage() string {
	if len(t.Config.Images) > 0 {
		return t.Config.Images[0]
	}
	return ""
}
func (t *Case) destroyServer(serverId string) {
	console.Info("[%s] deleting server", serverId)
	if err := t.Client.NovaV2().DeleteServer(serverId); err != nil {
		console.Error("[%s] delete failed: %s", serverId, err)
		return
	}
	console.Info("[%s] wait deleted", serverId)
	if err := t.Client.NovaV2().WaitServerDeleted(serverId); err != nil {
		console.Error("[%s] wait deleted failed: %s", serverId, err)
	}
}
func (t *Case) AddActionsReport(testId int, report WorkerReport) {
	t.reportLock.Lock()
	t.caseReport.WorkerReports = append(t.caseReport.WorkerReports, report)
	t.reportLock.Unlock()
}
func (t *Case) testActions(testId int, serverId string) (actionsReport WorkerReport) {
	// 执行一轮测试
	var (
		server *nova.Server
		err    error
	)
	actionsReport.Init(testId, lo.CoalesceOrEmpty(serverId, "-"))

	console.Info("start worker, id=%d", testId)
	if serverId == "" {
		if t.firstFlavor() == "" {
			actionsReport.Error = fmt.Errorf("flavors is empty")
			return
		}
		if t.firstImage() == "" {
			actionsReport.Error = fmt.Errorf("images is empty")
			return
		}
		server, err = t.createDefaultServer(testId)
		if server != nil {
			actionsReport.Server = server.Id
		}
		if err != nil {
			console.Error("create server failed, %s", err)
			actionsReport.Error = fmt.Errorf("create server failed")
			return
		}
		if err := t.waitServerCreated(server.Id); err != nil {
			console.Error("[%s] create server failed, %s", server.Id, err)
			actionsReport.Error = fmt.Errorf("create server failed")
			return
		}
		server, err = t.Client.NovaV2().GetServer(server.Id)
		if err != nil {
			console.Error("get server failed, %s", err)
			actionsReport.Error = fmt.Errorf("get server failed")
			return
		}
		serverCheckers, err := checkers.GetServerCheckers(t.Client, server, t.Config.QGAChecker)
		if err != nil {
			console.Error("get server checkers failed, %s", err)
			actionsReport.Error = fmt.Errorf("get server checkers failed")
			return
		}
		if err := serverCheckers.MakesureServerRunning(); err != nil {
			console.Error("server is not running, %s", err)
			actionsReport.Error = fmt.Errorf("server is not running")
			return
		}
		defer func() {
			if (!actionsReport.HasError() && t.Config.DeleteIfSuccess) ||
				(actionsReport.HasError() && t.Config.DeleteIfError) {
				t.destroyServer(server.Id)
			}
		}()
	} else {
		var err error
		server, err = t.Client.NovaV2().FindServer(serverId)
		if err != nil {
			console.Error("get server failed, %s", err)
			actionsReport.Error = fmt.Errorf("get server failed")
			return
		}
		actionsReport.Server = server.Id

		console.Info("use server: %s(%s)", server.Id, server.Name)
	}
	console.Info("[%s] ======== Test actions: %s, workers: %d", server.Id,
		strings.Join(t.Actions.FormatActions(), ","), t.Config.Workers,
	)

	title := t.Name
	if title == "" {
		title = strings.Join(t.Actions.FormatActions(), ",")
	}
	pbr := console.NewPbr(t.Actions.Total(), fmt.Sprintf("%s (%d)", title, testId))
	for _, actionName := range t.Actions.Actions() {
		action := internal.VALID_ACTIONS.Get(actionName, server, t.Client)
		action.SetConfig(t.Config)
		if action == nil {
			actionsReport.Error = fmt.Errorf("action '%s' not found", actionName)
			return
		}
		console.Info(utility.BlueString("[%s] ==== %s"), server.Id, actionName)

		// 更新实例信息
		if err := action.RefreshServer(); err != nil {
			actionsReport.Error = fmt.Errorf("action '%s' not found", actionName)
			return
		}
		if t.Config.ActionInterval > 0 {
			console.Info("[%s] sleep %d seconds", server.Id, t.Config.ActionInterval)
			time.Sleep(time.Second * time.Duration(t.Config.ActionInterval))
		}
		// 开始测试
		err := startAction(action)
		pbr.Increment()
		// 更新测试结果
		actionsReport.Results = append(actionsReport.Results, ActionResult{Action: actionName, Error: err})
		if err != nil {
			console.Error("[%s] test '%s' failed: %s", server.Id, actionName, err)
			actionsReport.Error = fmt.Errorf("test '%s' failed", actionName)
			return
		}
	}
	return actionsReport
}

func (t *Case) PrintServerEvents() error {
	return nil
}
func (t *Case) init() {
	if t.reportLock == nil {
		t.reportLock = &sync.Mutex{}
	}
	t.caseReport.Name = t.Name
	t.caseReport.Workers = t.Config.Workers
	t.caseReport.Actions = strings.Join(t.Actions.FormatActions(), ",")
}

func (t *Case) Report() CaseReport {
	return t.caseReport
}

func (t *Case) Start() {
	t.init()
	if t.Actions.Empty() {
		console.Warn("action is empty")
		return
	}
	console.Info("run case, worker=%d actions=%s", t.Config.Workers, t.Actions.FormatActions())
	go console.WaitPbrs()
	if len(t.UseServers) > 0 {
		console.Warn("use exits servers: %s", strings.Join(t.UseServers, ","))
		syncutils.StartTasks(
			syncutils.TaskOption{MaxWorker: len(t.UseServers)},
			lo.RangeFrom(1, t.Config.Workers),
			func(testId int) error {
				report := t.testActions(testId, t.UseServers[testId-1])
				t.AddActionsReport(testId, report)
				return nil
			},
		)
	} else {
		syncutils.StartTasks(
			syncutils.TaskOption{
				MaxWorker: max(t.Config.Workers, 1),
			},
			lo.RangeFrom(1, t.Config.Workers),
			func(testId int) error {
				report := t.testActions(testId, "")
				t.AddActionsReport(testId, report)
				return nil
			},
		)
	}
}
