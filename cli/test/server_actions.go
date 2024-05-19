package test

import (
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/server_actions"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func runActionTest(action server_actions.ServerAction) error {
	defer action.Cleanup()
	if err := action.Start(); err != nil {
		return err
	}
	return nil
}

var serverAction = &cobra.Command{
	Use:   "server-actions <server>",
	Short: "Test server actions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		actions, _ := cmd.Flags().GetString("actions")
		actionInterval, _ := cmd.Flags().GetInt("action-interval")

		serverActions, err := server_actions.ParseServerActions(actions)
		if err != nil {
			logging.Error("parse server actions failed: %s", err)
			return
		}
		client := openstack.DefaultClient()
		server, err := client.NovaV2().Servers().Found(args[0])
		utility.LogError(err, "get server failed", true)

		success, failed := 0, 0
		for i, actionName := range serverActions {
			action, err := server_actions.GetTestAction(actionName, server, client)
			if err != nil {
				logging.Error("[%s] get action failed: %s", server.Id, err)
				continue
			}
			logging.Info("[%s] ==== test %s start ====", server.Id, actionName)
			err = runActionTest(action)
			if err != nil {
				failed++
				logging.Error("[%s] test action '%s' %s: %s", server.Id, actionName,
					color.New(color.FgRed).Sprintf("failed"),
					err)
			} else {
				success++
				logging.Info("[%s] test action '%s' %s", server.Id, actionName, color.New(color.FgGreen).Sprintf("success"))
			}
			logging.Info("[%s] ==== test %s finished ====", server.Id, actionName)
			if i < len(serverActions)-1 {
				time.Sleep(time.Second * time.Duration(actionInterval))
			}
		}
		logging.Info("[%s] total: %d, success: %d, failed: %d", server.Id, len(serverActions), success, failed)
	},
}

func init() {
	serverAction.Flags().String("actions", "", "Test actions\nFormat: <action>[:count], "+
		"multiple actions separate by ','.\nExample: reboot,live_migrate:3")
	serverAction.Flags().Int("action-interval", 0, "Action interval")

	serverAction.MarkFlagRequired("actions")

	TestCmd.AddCommand(serverAction)
}
