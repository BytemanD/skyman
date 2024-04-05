package tool

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/BytemanD/easygo/pkg/arrayutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var attachInterfaces = &cobra.Command{
	Use:   "interfaces <server> <network1> [<network2>...]",
	Short: "Attach interfaces to server",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serverId := args[0]

		nums, _ := cmd.Flags().GetInt("nums")
		parallel, _ := cmd.Flags().GetInt("parallel")
		// clean, _ := cmd.Flags().GetBool("clean")
		useNetId, _ := cmd.Flags().GetBool("use-net-id")

		client := openstack.DefaultClient()
		neutronClient := client.NeutronV2()
		server, err := client.NovaV2().Servers().Show(serverId)
		utility.LogError(err, "show server failed:", true)

		netStrRing := utility.StringRing{Items: args[1:]}
		nets := netStrRing.Sample(nums)

		interfaces := []Interface{}
		mu := sync.Mutex{}

		if !useNetId {
			taskGroup := syncutils.TaskGroup{
				Items:        arrayutils.Range(len(nets)),
				MaxWorker:    parallel,
				ShowProgress: true,
				Func: func(item interface{}) error {
					p := item.(int)
					name := fmt.Sprintf("skyman-port-%d", p)
					logging.Debug("creating port %s", name)
					port, err := neutronClient.Ports().Create(
						map[string]interface{}{"name": name, "network_id": nets[p]})
					if err != nil {
						logging.Error("create port failed: %v", err)
						return err
					}
					logging.Info("created port: %v (%v)", port.Name, port.Id)
					mu.Lock()
					interfaces = append(interfaces, Interface{PortId: port.Id, Name: name})
					mu.Unlock()
					return nil
				},
			}
			logging.Info("creating %d port(s), waiting ...", nums)
			taskGroup.Start()
		} else {
			for _, net := range nets {
				interfaces = append(interfaces, Interface{NetId: net})
			}
		}
		if len(interfaces) == 0 {
			return
		}
		taskGroup2 := syncutils.TaskGroup{
			Items:        interfaces,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(Interface)
				logging.Debug("[interface: %s] attaching", p)
				attachment, err := client.NovaV2().Servers().AddInterface(server.Id, p.NetId, p.PortId)
				if err != nil {
					logging.Error("[interface: %s] attach failed: %v", p, err)
					return err
				}
				if attachment != nil && p.PortId == "" {
					p.PortId = attachment.PortId
				}
				interfaces, err := client.NovaV2().Servers().ListInterfaces(server.Id)
				if err != nil {
					utility.LogError(err, "list server interfaces failed:", false)
					return err
				}
				for _, vif := range interfaces {
					if vif.PortId == p.PortId {
						logging.Info("[interface: %s] attach success", p)
						return nil
					}
				}
				logging.Error("[interface: %s] attach failed", p)
				return nil
			},
		}
		logging.Info("attaching ...")
		taskGroup2.Start()
	},
}

func init() {
	attachInterfaces.Flags().Int("nums", 1, "nums of interfaces")
	attachInterfaces.Flags().Int("parallel", runtime.NumCPU(), "nums of parallel")
	attachInterfaces.Flags().Bool("use-net-id", false, "attach interface with network id rather than port id")

	attachCmd.AddCommand(attachInterfaces)
}
