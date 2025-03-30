package server

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var attachInterfaces = &cobra.Command{
	Use:   "add-interfaces <server> <network1> [<network2>...]",
	Short: "Add interfaces to server",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		nums, _ := cmd.Flags().GetInt("nums")
		parallel, _ := cmd.Flags().GetInt("parallel")
		useNetId, _ := cmd.Flags().GetBool("use-net-id")
		sg, _ := cmd.Flags().GetString("sg")

		client := common.DefaultClient()
		neutronClient := client.NeutronV2()
		server, err := client.NovaV2().FindServer(args[0])
		utility.LogError(err, "show server failed:", true)
		if server.IsError() {
			utility.LogIfError(err, true, "server %s is Error", args[0])
		}
		var securityGroup *neutron.SecurityGroup
		if sg != "" {
			securityGroup, err = neutronClient.FindSecurityGroup(sg)
			utility.LogIfError(err, true, "get security group %s failed:", sg)
		}

		netIds := []string{}
		for _, idOrName := range args[1:] {
			// tenant id
			net, err := client.NeutronV2().FindNetwork(idOrName)
			utility.LogIfError(err, true, "get net %s failed:", idOrName)
			netIds = append(netIds, net.Id)
		}

		netStrRing := utility.StringRing{Items: netIds}
		if nums == 0 {
			nums = len(netIds)
		}
		nets := netStrRing.Sample(nums)

		interfaces := []Interface{}
		mu := sync.Mutex{}

		if !useNetId {
			taskGroup := syncutils.TaskGroup{
				Items:        lo.Range(len(nets)),
				MaxWorker:    parallel,
				Title:        fmt.Sprintf("create %d port(s)", len(nets)),
				ShowProgress: true,
				Func: func(item interface{}) error {
					p := item.(int)
					name := fmt.Sprintf("skyman-port-%d", p+1)
					console.Debug("creating port %s", name)
					options := map[string]interface{}{
						"name": name, "network_id": nets[p],
					}
					if securityGroup != nil {
						options["security_groups"] = []string{securityGroup.Id}
					}
					port, err := neutronClient.CreatePort(options)

					if err != nil {
						console.Error("create port failed: %v", err)
						return err
					}
					console.Info("created port: %v (%v)", port.Name, port.Id)
					mu.Lock()
					interfaces = append(interfaces, Interface{PortId: port.Id, Name: name})
					mu.Unlock()
					return nil
				},
			}
			console.Info("creating %d port(s), waiting ...", nums)
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
			Title:        fmt.Sprintf("attach %d port(s)", len(interfaces)),
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(Interface)
				attachment, err := client.NovaV2().ServerAddInterface(server.Id, p.NetId, p.PortId)
				if err != nil {
					console.Error("[interface: %s] attach failed: %v", p, err)
					return err
				}
				console.Debug("[interface: %s] attaching", attachment.PortId)
				for {
					port, err := client.NeutronV2().GetPort(attachment.PortId)

					if securityGroup != nil && !slice.Contain(port.SecurityGroups, securityGroup.Id) {
						console.Info("[interface: %s] update port security group to %s(%s)", port.Id, sg, securityGroup.Id)
						_, err = client.NeutronV2().UpdatePort(
							port.Id,
							map[string]interface{}{"security_groups": []string{securityGroup.Id}},
						)
						utility.LogIfError(err, true, "[interface: %s]update port security group failed", port.Id)
					}

					if port != nil {
						console.Info("[interface: %s] vif type is %s", port.Id, port.BindingVifType)
						if err == nil && !port.IsUnbound() {
							console.Info("[interface: %s] attached", port.Id)
							break
						}
					}
					time.Sleep(time.Second * 3)
				}

				err = utility.RetryError(
					utility.RetryCondition{Timeout: time.Second * 60, IntervalMin: time.Second},
					func() (bool, error) {
						interfaces, err := client.NovaV2().ListServerInterfaces(server.Id)
						if err != nil {
							console.Error("list server interfaces failed: %s", err)
							return false, err
						}
						for _, vif := range interfaces {
							if vif.PortId == attachment.PortId {
								return false, nil
							}
						}
						return true, nil
					},
				)
				if err != nil {
					console.Error("[interface: %s] attach failed", attachment.PortId)
					attachFailed = true
				} else {
					console.Info("[interface: %s] attach success", attachment.PortId)
				}

				return nil
			},
		}
		console.Info("attaching ...")
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
	attachInterfaces.Flags().String("sg", "", "security group")

	ServerCommand.AddCommand(attachInterfaces)
}
