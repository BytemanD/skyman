package server

import (
	"fmt"
	"runtime"
	"time"

	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var detachInterfaces = &cobra.Command{
	Use:   "remove-interfaces <server>",
	Short: "Remove interfaces from server",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nums, _ := cmd.Flags().GetInt("nums")
		parallel, _ := cmd.Flags().GetInt("parallel")
		clean, _ := cmd.Flags().GetBool("clean")

		client := common.DefaultClient()
		neutronClient := client.NeutronV2()
		server, err := client.NovaV2().Server().Find(args[0])
		utility.LogError(err, "show server failed:", true)
		if server.IsError() {
			utility.LogIfError(err, true, "server %s is Error", args[0])
		}
		interfaces, err := client.NovaV2().Server().ListInterfaces(server.Id)
		utility.LogError(err, "list server interfaces failed:", true)

		console.Info("server has %d interfaces", len(interfaces))

		start := max(0, len(interfaces)-nums)
		detachInterfaces := interfaces[start:]
		if len(detachInterfaces) == 0 {
			console.Warn("nothing to do")
			return
		}
		taskGroup2 := syncutils.TaskGroup{
			Items:        detachInterfaces,
			MaxWorker:    parallel,
			Title:        fmt.Sprintf("detach %d interface(s)", len(interfaces)),
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(nova.InterfaceAttachment)
				console.Info("[interface: %s] detaching", p.PortId)
				err := client.NovaV2().Server().DeleteInterfaceAndWait(server.Id, p.PortId, time.Minute*5)
				if err != nil {
					console.Error("[interface: %s] detach failed: %v", p.PortId, err)
					return err
				}
				if clean {
					err = neutronClient.Port().Delete(p.PortId)
					if err == nil {
						console.Info("[interface: %s] deleted", p.PortId)
					} else {
						console.Info("[interface: %s] delete failed: %s", p.PortId, err)
					}
				}
				return nil
			},
		}
		console.Info("detaching ...")
		taskGroup2.Start()
	},
}

func init() {
	detachInterfaces.Flags().Int("nums", 1, "nums of interfaces")
	detachInterfaces.Flags().Int("parallel", runtime.NumCPU(), "nums of parallel")
	detachInterfaces.Flags().Bool("clean", false, "delete interface after detached")

	ServerCommand.AddCommand(detachInterfaces)
}
