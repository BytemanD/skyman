package tool

import (
	"runtime"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var detachInterfaces = &cobra.Command{
	Use:   "interfaces <server>",
	Short: "Attach interfaces to server",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverId := args[0]

		nums, _ := cmd.Flags().GetInt("nums")
		parallel, _ := cmd.Flags().GetInt("parallel")
		clean, _ := cmd.Flags().GetBool("clean")

		client := openstack.DefaultClient()
		neutronClient := client.NeutronV2()
		server, err := client.NovaV2().Servers().Show(serverId)
		utility.LogError(err, "show server failed:", true)

		interfaces, err := client.NovaV2().Servers().ListInterfaces(server.Id)
		utility.LogError(err, "list server interfaces failed:", true)

		logging.Info("server has %d interfaces", len(interfaces))

		start := max(0, len(interfaces)-nums)
		detachInterfaces := interfaces[start:]
		if len(detachInterfaces) == 0 {
			logging.Warning("nothing to do")
			return
		}
		taskGroup2 := syncutils.TaskGroup{
			Items:        detachInterfaces,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(nova.InterfaceAttachment)
				logging.Info("[interface: %s] detaching", p.PortId)
				err := client.NovaV2().Servers().DeleteInterfaceAndWait(server.Id, p.PortId, 600)
				if err != nil {
					logging.Error("[interface: %s] detach failed: %v", p.PortId, err)
					return err
				}
				if clean {
					err = neutronClient.Ports().Delete(p.PortId)
					if err == nil {
						logging.Info("[interface: %s] deleted", p.PortId)
					} else {
						logging.Info("[interface: %s] delete failed: %s", p.PortId, err)
					}
				}
				return nil
			},
		}
		logging.Info("detaching ...")
		taskGroup2.Start()
	},
}

func init() {
	detachInterfaces.Flags().Int("nums", 1, "nums of interfaces")
	detachInterfaces.Flags().Int("parallel", runtime.NumCPU(), "nums of parallel")
	detachInterfaces.Flags().Bool("clean", false, "delete interface after detached")

	detachCmd.AddCommand(detachInterfaces)
}
