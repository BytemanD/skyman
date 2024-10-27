package test

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/arrayutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/cli/test/checkers"
	"github.com/BytemanD/skyman/cli/test/server_actions"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runActionTest(action server_actions.ServerAction) (bool, error) {
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

func getServerBootOption(testId int) nova.ServerOpt {
	opt := nova.ServerOpt{
		Name:             fmt.Sprintf("skyman-server-%d-%v", testId, time.Now().Format("20060102-150405")),
		Flavor:           server_actions.TEST_FLAVORS[0].Id,
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
func createDefaultServer(client *openstack.Openstack, testId int) (*nova.Server, error) {
	bootOption := getServerBootOption(testId)
	return client.NovaV2().Server().Create(bootOption)
}
func waitServerCreated(client *openstack.Openstack, server *nova.Server) error {
	var err error
	server, err = client.NovaV2().Server().Show(server.Id)
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

func preTest(client *openstack.Openstack) {
	logging.Info("check flavors ...")
	for _, flavorId := range common.TASK_CONF.Flavors {
		flavor, err := client.NovaV2().Flavor().Found(flavorId, false)
		utility.LogError(err, fmt.Sprintf("get flavor %s failed", flavorId), true)
		server_actions.TEST_FLAVORS = append(server_actions.TEST_FLAVORS, *flavor)
	}
	logging.Info("check images ...")
	for _, idOrName := range common.TASK_CONF.Images {
		_, err := client.GlanceV2().Images().Found(idOrName)
		utility.LogError(err, fmt.Sprintf("get image %s failed", idOrName), true)
	}
	logging.Info("check networks ...")
	for _, idOrName := range common.TASK_CONF.Networks {
		_, err := client.NeutronV2().Network().Show(idOrName)
		utility.LogError(err, fmt.Sprintf("get network %s failed", idOrName), true)
	}
}
func runTests(client *openstack.Openstack, serverId string, testId int, actionInterval int, tasks [][]string) error {
	for _, task := range tasks {
		if err := runTest(client, serverId, testId, actionInterval, task); err != nil {
			return err
		}
	}
	return nil
}

func runTest(client *openstack.Openstack, serverId string, testId int, actionInterval int, serverActions []string) error {
	if len(serverActions) == 0 {
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
		TotalActions: serverActions,
	}
	if serverId != "" {
		server, err = client.NovaV2().Server().Found(serverId)
		if err != nil {
			return fmt.Errorf("get server failed: %s", err)
		}
		task.ServerId = serverId
		server_actions.TestTasks = append(server_actions.TestTasks, &task)
		logging.Info("use server: %s(%s)", server.Id, server.Name)
		logging.Info("[%s] ======== %s ========", serverId, strings.Join(serverActions, ","))

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
		logging.Info("[%s] ======== Test actions: %s", server.Id, strings.Join(serverActions, ","))
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
			if !task.HasFailed() || common.TASK_CONF.DeleteIfError {
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
	for _, actionName := range serverActions {
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

var cliActions = []string{}

var serverAction = &cobra.Command{
	Use:   "server-actions <TASK FILE> [server]",
	Short: "Test server actions",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		actions, _ := cmd.Flags().GetString("actions")
		if testActions, err := server_actions.ParseServerActions(actions); err != nil {
			return err
		} else {
			cliActions = testActions
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := common.LoadTaskConfig(args[0]); err != nil {
			logging.Info("load task file %s failed", args[0])
			os.Exit(1)
		}

		total, _ := cmd.Flags().GetInt("total")
		worker, _ := cmd.Flags().GetInt("worker")
		servers, _ := cmd.Flags().GetString("servers")
		actionInterval, _ := cmd.Flags().GetInt("action-interval")
		reportEvents, _ := cmd.Flags().GetBool("report-events")
		web, _ := cmd.Flags().GetBool("web")

		if worker > 0 {
			common.TASK_CONF.Workers = worker
		}
		if total > 0 {
			common.TASK_CONF.Total = total
		}
		if actionInterval >= 0 {
			common.TASK_CONF.ActionInterval = actionInterval
		}
		if servers != "" {
			common.TASK_CONF.UseServers = strings.Split(servers, ",")
		}

		var TestActions = [][]string{}
		// 检查 actions
		if len(cliActions) == 0 {
			for _, actions := range common.TASK_CONF.ActionTasks {
				if testActions, err := server_actions.ParseServerActions(actions); err != nil {
					logging.Fatal("parse action failed: %s", err)
				} else {
					TestActions = append(TestActions, testActions)
				}
			}
		} else {
			TestActions = append(TestActions, cliActions)
		}
		if len(TestActions) == 0 {
			logging.Error("test actions is empty")
			return
		}

		logging.Info("===== sever action tasks =====")
		for i, act := range TestActions {
			logging.Info("%d. %s", i+1, act)
		}
		logging.Info("==============================")

		client := openstack.DefaultClient()
		// 测试前检查
		preTest(client)

		if web {
			go RunSimpleWebServer()
		}
		logging.Info("tasks total: %d, workers: %d, action interval: %d",
			common.TASK_CONF.Total, common.TASK_CONF.Workers, common.TASK_CONF.ActionInterval)
		var taskGroup syncutils.TaskGroup = syncutils.TaskGroup{
			MaxWorker: common.TASK_CONF.Workers,
		}
		if len(common.TASK_CONF.UseServers) > 0 {
			logging.Info("test with servers: %s", strings.Join(common.TASK_CONF.UseServers, ","))
			taskGroup.Items = arrayutils.Range(1, len(common.TASK_CONF.UseServers)+1)
			taskGroup.Func = func(item interface{}) error {
				index := item.(int)
				err := runTests(client, common.TASK_CONF.UseServers[index-1], index,
					common.TASK_CONF.ActionInterval, TestActions)
				if err != nil {
					logging.Error("[%s] test failed: %s", common.TASK_CONF.UseServers[index-1], err)
				}
				return nil
			}
		} else {
			taskGroup.Items = arrayutils.Range(1, common.TASK_CONF.Total+1)
			taskGroup.Func = func(item interface{}) error {
				err := runTests(client, "", item.(int), common.TASK_CONF.ActionInterval, TestActions)
				if err != nil {
					logging.Error("test failed: %s", err)
				}
				return nil
			}
		}
		taskGroup.Start()

		server_actions.PrintTestTasks()
		if reportEvents {
			server_actions.PrintServerEvents(client)
		}
		if web {
			WaitWebServer()
		}
	},
}

func init() {
	supportedActions := []string{}

	for _, actions := range arrayutils.SplitStrings(server_actions.VALID_ACTIONS.Keys(), 5) {
		supportedActions = append(supportedActions, strings.Join(actions, ", "))
	}
	serverAction.Flags().StringP("actions", "A", "", "Test actions\nFormat: <action>[:count], "+
		"multiple actions separate by ','.\nExample: reboot,live_migrate:3\n"+
		"Actions: "+strings.Join(supportedActions, ",\n         "),
	)

	serverAction.Flags().Int("action-interval", 0, "Action interval")
	serverAction.Flags().Int("total", 0, i18n.T("theNumOfTask"))
	serverAction.Flags().Int("worker", 0, i18n.T("theNumOfWorker"))
	serverAction.Flags().String("servers", "", "Use existing servers")
	serverAction.Flags().Bool("report-events", false, i18n.T("reportServerEvents"))
	serverAction.Flags().Bool("web", false, "Start web server")

	viper.BindPFlag("test.total", serverAction.Flags().Lookup("total"))
	viper.BindPFlag("test.workers", serverAction.Flags().Lookup("worker"))
	viper.BindPFlag("test.actionInterval", serverAction.Flags().Lookup("action-interval"))

	TestCmd.AddCommand(serverAction)
}
