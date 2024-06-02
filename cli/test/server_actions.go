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
	defer action.Cleanup()
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
	logging.Info("[%s] creating server %s", server.Id, server.Resource.Name)
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
		boot       bool
		testFailed bool
	)
	success, failed, skip := 0, 0, 0
	if serverId != "" {
		server, err = client.NovaV2().Servers().Found(serverId)
		utility.LogError(err, "get server failed", true)
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
		boot = true
	}
	for i, actionName := range serverActions {
		action, err := server_actions.GetTestAction(actionName, server, client)
		if err != nil {
			logging.Error("[%s] get action failed: %s", server.Id, err)
			skip++
			continue
		}
		logging.Info("[%s] %s", server.Id, utility.BlueString("==== test "+actionName+" start ===="))
		err = runActionTest(action)
		if err != nil {
			failed++
			testFailed = true
			logging.Error("[%s] test action '%s' failed: %s", server.Id, actionName, err)
		} else {
			success++
			logging.Success("[%s] test action '%s' success", server.Id, actionName)
		}
		logging.Info("[%s] %s", server.Id, utility.BlueString("==== test "+actionName+" finished ===="))
		if i < len(serverActions)-1 {
			time.Sleep(time.Second * time.Duration(actionInterval))
		}
	}
	if boot && (!testFailed || common.CONF.Test.DeleteIfError) {
		logging.Info("[%s] deleting", server.Id)
		client.NovaV2().Servers().Delete(server.Id)
		client.NovaV2().Servers().WaitDeleted(server.Id)
	}
	if len(serverActions) == success {
		logging.Success("[%s] total actions: %d, success: %d, failed: %d, skip: %d",
			server.Id, len(serverActions), success, failed, skip)
	} else if failed > 0 {
		logging.Error("[%s] total actions: %d, success: %d, failed: %d, skip: %d",
			server.Id, len(serverActions), success, failed, skip)
	} else {
		logging.Warning("[%s] total actions: %d, success: %d, failed: %d, skip: %d",
			server.Id, len(serverActions), success, failed, skip)
	}

	return nil
}

var serverAction = &cobra.Command{
	Use:   "server-actions [server]",
	Short: "Test server actions",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		actions, _ := cmd.Flags().GetString("actions")
		actionInterval, _ := cmd.Flags().GetInt("action-interval")
		// 检查 actions
		if actions == "" {
			actions = strings.Join(common.CONF.Test.Actions, ",")
		}
		serverActions, err := server_actions.ParseServerActions(actions)
		if err != nil {
			utility.LogError(err, "parse server actions failed", true)
		}
		logging.Info("test actions: %s", strings.Join(serverActions, ", "))

		client := openstack.DefaultClient()
		preTest(client)
		if len(args) == 1 {
			runTest(client, args[0], actionInterval, 1, serverActions)
		} else {
			logging.Info("tasks: %d, workers: %d", common.CONF.Test.Tasks, common.CONF.Test.Workers)
			taskGroup := syncutils.TaskGroup{
				Items:     arrayutils.Range(1, common.CONF.Test.Tasks+1),
				MaxWorker: common.CONF.Test.Workers,
				Func: func(item interface{}) error {
					err := runTest(client, "", item.(int), actionInterval, serverActions)
					if err != nil {
						logging.Error("test failed: %v", err)
						return nil
					}
					return nil
				},
			}
			taskGroup.Start()
		}
	},
}

func init() {
	supportedActions := []string{}
	for _, actions := range arrayutils.SplitStrings(server_actions.GetActions(), 5) {
		supportedActions = append(supportedActions, strings.Join(actions, ", "))
	}
	serverAction.Flags().String("actions", "", "Test actions\nFormat: <action>[:count], "+
		"multiple actions separate by ','.\nExample: reboot,live_migrate:3\n"+
		"Actions: "+strings.Join(supportedActions, ",\n         "),
	)

	serverAction.Flags().Int("action-interval", 0, "Action interval")
	serverAction.Flags().Int("tasks", 0, "The num of task")
	serverAction.Flags().Int("workers", 0, "The num of worker")

	viper.BindPFlag("test.tasks", serverAction.Flags().Lookup("tasks"))
	viper.BindPFlag("test.workers", serverAction.Flags().Lookup("workers"))

	TestCmd.AddCommand(serverAction)
}
