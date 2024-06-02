package test

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/arrayutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/server_actions"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

func runActionTest(action server_actions.ServerAction) error {
	defer action.Cleanup()
	if err := action.Start(); err != nil {
		return err
	}
	return nil
}

func getServerBootOption() nova.ServerOpt {
	opt := nova.ServerOpt{
		Name:             fmt.Sprintf("skyman-server-%v", time.Now().Format("20060102-150405")),
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
func createDefaultServer(client *openstack.Openstack) (*nova.Server, error) {
	bootOption := getServerBootOption()
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
	logging.Info("[%s] create %s", server.Id, utility.GreenString("success"))
	return server, nil
}

func preTest(client *openstack.Openstack) {
	logging.Info("refresh flavors")
	for _, flavorId := range common.CONF.Test.Flavors {
		flavor, err := client.NovaV2().Flavors().Found(flavorId)
		utility.LogError(err, fmt.Sprintf("get flavor %s failed", flavorId), true)
		server_actions.TEST_FLAVORS = append(server_actions.TEST_FLAVORS, *flavor)
	}
	logging.Info("refresh images")
	for _, idOrName := range common.CONF.Test.Images {
		_, err := client.GlanceV2().Images().Found(idOrName)
		utility.LogError(err, fmt.Sprintf("get image %s failed", idOrName), true)
	}
	logging.Info("refresh networks")
	for _, idOrName := range common.CONF.Test.Networks {
		_, err := client.NeutronV2().Networks().Show(idOrName)
		utility.LogError(err, fmt.Sprintf("get network %s failed", idOrName), true)
	}
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
		var (
			server *nova.Server
			boot   bool
		)
		if len(args) == 1 {
			server, err = client.NovaV2().Servers().Found(args[0])
			utility.LogError(err, "get server failed", true)
			logging.Info("use server: %s(%s)", server.Id, server.Name)
		} else {
			if len(server_actions.TEST_FLAVORS) == 0 {
				logging.Info("test flavors is empty")
				os.Exit(1)
			}
			if len(common.CONF.Test.Images) == 0 {
				logging.Info("test images is empty")
				os.Exit(1)
			}
			server, err = createDefaultServer(client)
			utility.LogError(err, "create server failed", true)
			boot = true
		}

		success, failed, skip := 0, 0, 0
		testFailed := false
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
				logging.Error("[%s] test action '%s' %s: %s", server.Id, actionName,
					utility.RedString("failed"), err)
			} else {
				success++
				logging.Info("[%s] test action '%s' %s", server.Id, actionName, utility.GreenString("success"))
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
		logging.Info("[%s] total: %d, success: %d, failed: %d, skip: %d",
			server.Id, len(serverActions), success, failed, skip)
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

	// serverAction.MarkFlagRequired("actions")

	TestCmd.AddCommand(serverAction)
}
