package test

import (
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

func RebootServer(novaClient openstack.NovaV2, id string, hard bool) {
	err := novaClient.Servers().Reboot(id, hard)
	if err != nil {
		logging.Error("[%s] reboot failed, %v", id, err)
	}

	logging.Info("[%s] rebooting", id)

	_, err = novaClient.Servers().WaitStatus(id, "ACTIVE", 5)
	if err == nil {
		logging.Info("[%s] rebooted", id)
	} else {
		logging.Error("[%s] reboot failed, %v", id, err)
	}
}

var reboot = &cobra.Command{
	Use:   "reboot <server> [<server> ...]",
	Short: "test server reboot",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hard, _ := cmd.Flags().GetBool("hard")
		times, _ := cmd.Flags().GetInt("times")

		client := openstack.DefaultClient()
		task := syncutils.TaskGroup{
			Items: args,
			Func: func(item interface{}) error {
				serviceId := item.(string)
				server, err := client.NovaV2().Servers().Found(serviceId)
				utility.LogError(err, "get service failed", true)
				for i := 0; i < times; i++ {
					RebootServer(*client.NovaV2(), server.Id, hard)
				}
				return nil
			},
		}
		task.Start()
	},
}

func init() {
	reboot.Flags().Bool("hard", false, "Perform a hard reboot")
	reboot.Flags().Int("times", 1, "Reboot times")

	TestCmd.AddCommand(reboot)
}
