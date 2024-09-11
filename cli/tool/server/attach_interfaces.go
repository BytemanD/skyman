package server

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/BytemanD/easygo/pkg/arrayutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var attachInterfaces = &cobra.Command{
	Use:   "add-interfaces <server> <network1> [<network2>...]",
	Short: "Add interfaces to server",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		nums, _ := cmd.Flags().GetInt("nums")
		parallel, _ := cmd.Flags().GetInt("parallel")
		// clean, _ := cmd.Flags().GetBool("clean")
		useNetId, _ := cmd.Flags().GetBool("use-net-id")

		client := openstack.DefaultClient()
		neutronClient := client.NeutronV2()
		server, err := client.NovaV2().Server().Found(args[0])
		utility.LogError(err, "show server failed:", true)

		netIds := []string{}
		for _, idOrName := range args[1:] {
			net, err := client.NeutronV2().Network().Found(idOrName)
			utility.LogIfError(err, true, "get net %s failed:", idOrName)
			netIds = append(netIds, net.Id)
		}

		netStrRing := utility.StringRing{Items: netIds}
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
					port, err := neutronClient.Port().Create(
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
		attachFailed := false
		taskGroup2 := syncutils.TaskGroup{
			Items:        interfaces,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(Interface)
				attachment, err := client.NovaV2().Server().AddInterface(server.Id, p.NetId, p.PortId)
				if err != nil {
					logging.Error("[interface: %s] attach failed: %v", p, err)
					return err
				}
				logging.Debug("[interface: %s] attaching", attachment.PortId)

				interfaces, err := client.NovaV2().Server().ListInterfaces(server.Id)
				if err != nil {
					utility.LogError(err, "list server interfaces failed:", false)
					return err
				}
				for {
					port, err := client.NeutronV2().Port().Show(attachment.PortId)
					if port != nil {
						logging.Info("[interface: %s] vif type is %s", port.Id, port.BindingVifType)
						if err == nil && !port.IsUnbound() {
							logging.Info("[interface: %s] attached", port.Id)
							break
						}
					}
					time.Sleep(time.Second * 3)
				}

				for _, vif := range interfaces {
					if vif.PortId == p.PortId {
						logging.Info("[interface: %s] attach success", attachment.PortId)
						return nil
					}
				}
				logging.Error("[interface: %s] attach failed", attachment.PortId)
				attachFailed = true
				return nil
			},
		}
		logging.Info("attaching ...")
		taskGroup2.Start()
		if attachFailed {
			os.Exit(1)
		}
	},
}

func init() {
	attachInterfaces.Flags().Int("nums", 1, "nums of interfaces")
	attachInterfaces.Flags().Int("parallel", runtime.NumCPU(), "nums of parallel")
	attachInterfaces.Flags().Bool("use-net-id", false, "attach interface with network id rather than port id")

	ServerCommand.AddCommand(attachInterfaces)
}
