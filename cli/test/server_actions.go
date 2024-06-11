package test

import (
	"fmt"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/arrayutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/cli/test/server_actions"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runActionTest(action server_actions.ServerAction) error {
	defer func() {
		logging.Info("[%s] cleanup ...", action.ServerId())
		action.Cleanup()
	}()

	if err := action.Start(); err != nil {
		return err
	}

	return nil
}

func getServerBootOption(testId int) nova.ServerOpt {
	opt := nova.ServerOpt{
		Name:             fmt.Sprintf("skyman-server-%d-%v", testId, time.Now().Format("20060102-150405")),
		Flavor:           server_actions.TEST_FLAVORS[0].Id,
		Image:            common.CONF.Test.Images[0],
		AvailabilityZone: common.CONF.Test.AvailabilityZone,
	}
	if len(common.CONF.Test.Networks) >= 1 {
		opt.Networks = []nova.ServerOptNetwork{
			{UUID: common.CONF.Test.Networks[0]},
		}
	} else {
		logging.Warning("boot without network")
	}

	if common.CONF.Test.BootFromVolume {
		opt.BlockDeviceMappingV2 = []nova.BlockDeviceMappingV2{
			{
				UUID:               common.CONF.Test.Images[0],
				VolumeSize:         common.CONF.Test.BootVolumeSize,
				SourceType:         "image",
				DestinationType:    "volume",
				VolumeType:         common.CONF.Test.BootVolumeType,
				DeleteOnTemination: true,
			},
		}
	} else {
		opt.Image = common.CONF.Test.Images[0]
	}
	return opt
}
func createDefaultServer(client *openstack.Openstack, testId int) (*nova.Server, error) {
	bootOption := getServerBootOption(testId)
	server, err := client.NovaV2().Servers().Create(bootOption)
	if err != nil {
		return nil, err
	}
	server, err = client.NovaV2().Servers().Show(server.Id)
	if err != nil {
		return nil, err
	}
	logging.Info("[%s] creating with name %s", server.Id, server.Resource.Name)
	server, err = client.NovaV2().Servers().WaitBooted(server.Id)
	if err != nil {
		return nil, err
	}
	logging.Info("[%s] create success", server.Id)
	return server, nil
}

func preTest(client *openstack.Openstack) {
	logging.Info("check flavors ...")
	for _, flavorId := range common.CONF.Test.Flavors {
		flavor, err := client.NovaV2().Flavors().Found(flavorId)
		utility.LogError(err, fmt.Sprintf("get flavor %s failed", flavorId), true)
		server_actions.TEST_FLAVORS = append(server_actions.TEST_FLAVORS, *flavor)
	}
	logging.Info("check images ...")
	for _, idOrName := range common.CONF.Test.Images {
		_, err := client.GlanceV2().Images().Found(idOrName)
		utility.LogError(err, fmt.Sprintf("get image %s failed", idOrName), true)
	}
	logging.Info("check networks ...")
	for _, idOrName := range common.CONF.Test.Networks {
		_, err := client.NeutronV2().Networks().Show(idOrName)
		utility.LogError(err, fmt.Sprintf("get network %s failed", idOrName), true)
	}
}

func runTest(client *openstack.Openstack, serverId string, testId int, actionInterval int, serverActions []string) error {
	var (
		server     *nova.Server
		err        error
		testFailed bool
	)
	success, failed, skip := 0, 0, 0
	if serverId != "" {
		server, err = client.NovaV2().Servers().Found(serverId)
		if err != nil {
			return fmt.Errorf("get server failed: %s", err)
		}
		logging.Info("use server: %s(%s)", server.Id, server.Name)
	} else {
		if len(server_actions.TEST_FLAVORS) == 0 {
			logging.Error("test flavors is empty")
			return fmt.Errorf("test flavors is empty")
		}
		if len(common.CONF.Test.Images) == 0 {
			return fmt.Errorf("test images is empty")
		}
		server, err = createDefaultServer(client, testId)
		if err != nil {
			return fmt.Errorf("create server failed: %s", err)
		}
		defer func() {
			if !testFailed || common.CONF.Test.DeleteIfError {
				logging.Info("[%s] deleting server", server.Id)
				client.NovaV2().Servers().Delete(server.Id)
				client.NovaV2().Servers().WaitDeleted(server.Id)
			}
		}()

		if common.CONF.Test.QGAChecker.Enabled {
			checker := server_actions.QGAChecker{Client: client}
			if err := checker.MakesureServerBooted(server); err != nil {
				return err
			}
			// if err := checker.MakesureHostname(server, server.Name); err != nil {
			// 	return err
			// }
		}

	}
	for i, actionName := range serverActions {
		action := server_actions.ACTIONS.Get(actionName, server, client)
		if action == nil {
			logging.Error("[%s] action '%s' not found", server.Id, action)
			skip++
			continue
		}
		logging.Info("[%s] %s", server.Id, utility.BlueString("==== "+actionName+" ===="))
		err = runActionTest(action)
		if err != nil {
			failed++
			testFailed = true
			logging.Error("[%s] test action '%s' failed: %s", server.Id, actionName, err)
		} else {
			success++
			logging.Success("[%s] test action '%s' success", server.Id, actionName)
		}
		if i < len(serverActions)-1 {
			time.Sleep(time.Second * time.Duration(actionInterval))
		}
	}

	result := fmt.Sprintf("all actions: %d, success: %d, failed: %d, skip: %d",
		len(serverActions), success, failed, skip)
	if len(serverActions) == success {
		logging.Success("[%s] %s", server.Id, result)
	} else if failed > 0 {
		logging.Error("[%s] %s", server.Id, result)
	} else {
		logging.Warning("[%s] %s", server.Id, result)
	}

	return nil
}

var TestActions []string

var serverAction = &cobra.Command{
	Use:   "server-actions [server]",
	Short: "Test server actions",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		actions, _ := cmd.Flags().GetString("actions")
		if testActions, err := server_actions.ParseServerActions(actions); err != nil {
			return err
		} else {
			TestActions = testActions
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		actionInterval, _ := cmd.Flags().GetInt("action-interval")
		servers, _ := cmd.Flags().GetString("servers")
		// 检查 actions
		if len(TestActions) == 0 {
			if testActions, err := server_actions.ParseServerActions(strings.Join(common.CONF.Test.Actions, ",")); err != nil {
				logging.Fatal("parse action failed: %s", err)
			} else {
				TestActions = testActions
			}
		}
		if len(TestActions) == 0 {
			logging.Warning("test actions is empty")
		} else {
			logging.Info("test actions: %s", strings.Join(TestActions, ", "))
		}
		if servers != "" {
			common.CONF.Test.UseServers = strings.Split(servers, ",")
		}
		client := openstack.DefaultClient()
		preTest(client)
		logging.Info("tasks: %d, workers: %d", common.CONF.Test.Tasks, common.CONF.Test.Workers)

		var taskGroup syncutils.TaskGroup
		if len(common.CONF.Test.UseServers) > 0 {
			logging.Info("test with servers: %s", strings.Join(common.CONF.Test.UseServers, ","))
			taskGroup = syncutils.TaskGroup{
				Items:     arrayutils.Range(1, len(common.CONF.Test.UseServers)+1),
				MaxWorker: common.CONF.Test.Workers,
				Func: func(item interface{}) error {
					index := item.(int)
					err := runTest(client, common.CONF.Test.UseServers[index-1], index, actionInterval, TestActions)
					if err != nil {
						logging.Error("[%s] test failed: %s", common.CONF.Test.UseServers[index-1], err)
					}
					return nil
				},
			}
		} else {
			taskGroup = syncutils.TaskGroup{
				Items:     arrayutils.Range(1, common.CONF.Test.Tasks+1),
				MaxWorker: common.CONF.Test.Workers,
				Func: func(item interface{}) error {
					err := runTest(client, "", item.(int), actionInterval, TestActions)
					if err != nil {
						logging.Error("test failed: %s", err)
					}
					return nil
				},
			}
		}
		taskGroup.Start()
	},
}

func init() {
	supportedActions := []string{}
	for _, actions := range arrayutils.SplitStrings(server_actions.ACTIONS.Keys(), 5) {
		supportedActions = append(supportedActions, strings.Join(actions, ", "))
	}
	serverAction.Flags().String("actions", "", "Test actions\nFormat: <action>[:count], "+
		"multiple actions separate by ','.\nExample: reboot,live_migrate:3\n"+
		"Actions: "+strings.Join(supportedActions, ",\n         "),
	)

	serverAction.Flags().Int("action-interval", 0, "Action interval")
	serverAction.Flags().Int("tasks", 0, "The num of task")
	serverAction.Flags().Int("workers", 0, "The num of worker")
	serverAction.Flags().String("servers", "", "Use existing servers")

	viper.BindPFlag("test.tasks", serverAction.Flags().Lookup("tasks"))
	viper.BindPFlag("test.workers", serverAction.Flags().Lookup("workers"))
	viper.BindPFlag("test.actionInterval", serverAction.Flags().Lookup("action-interval"))

	TestCmd.AddCommand(serverAction)
}
