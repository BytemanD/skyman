package test

import (
	"fmt"
	"sync"
	"time"

	"github.com/BytemanD/easygo/pkg/arrayutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{Use: "attach", Short: "attach ports to a server"}

var interfaceAttachPorts = &cobra.Command{
	Use:   "ports <server> <network id>",
	Short: "Attach ports to server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		nums, _ := cmd.Flags().GetInt("nums")
		parallel, _ := cmd.Flags().GetInt("parallel")
		clean, _ := cmd.Flags().GetBool("clean")

		client := openstack.DefaultClient()
		neutronClient := client.NeutronV2()
		server, err := client.NovaV2().Servers().Show(args[0])
		utility.LogError(err, "show server failed:", true)

		ports := []neutron.Port{}
		mu := sync.Mutex{}

		taskGroup := syncutils.TaskGroup{
			Items:        arrayutils.Range(1, nums+1),
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(int)
				name := fmt.Sprintf("skyman-port-%d", p)
				logging.Debug("creating port %s", name)
				port, err := neutronClient.Ports().Create(
					map[string]interface{}{"name": name, "network_id": args[1]})
				if err != nil {
					logging.Error("create port failed: %v", err)
					return err
				}
				logging.Info("created port: %v (%v)", port.Name, port.Id)
				mu.Lock()
				ports = append(ports, *port)
				mu.Unlock()
				return nil
			},
		}
		logging.Info("creating %d port(s), waiting ...", nums)
		taskGroup.Start()

		taskGroup2 := syncutils.TaskGroup{
			Items:        ports,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(neutron.Port)
				logging.Debug("[port: %s] attaching", p.Id)
				_, err := client.NovaV2().Servers().AddInterface(server.Id, "", p.Id)
				if err != nil {
					logging.Error("[port: %s] attach failed: %v", p.Id, err)
					return err
				}
				interfaces, err := client.NovaV2().Servers().ListInterfaces(server.Id)
				if err != nil {
					utility.LogError(err, "list server interfaces failed:", false)
					return err
				}
				for _, vif := range interfaces {
					if vif.PortId == p.Id {
						logging.Info("[port: %s] attach success", p.Id)
						return nil
					}
				}
				logging.Error("[port: %s] attach failed", p.Id)
				return nil
			},
		}
		logging.Info("attaching ...")
		taskGroup2.Start()
		if !clean {
			return
		}
		taskGroup3 := syncutils.TaskGroup{
			Items:        ports,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(neutron.Port)
				logging.Debug("[port: %s] detaching", p.Id)
				err := client.NovaV2().Servers().DeleteInterface(server.Id, p.Id)
				if err != nil {
					logging.Error("[port: %s] attach failed: %v", p.Id, err)
					return err
				}
				for {
					interfaces, err := client.NovaV2().Servers().ListInterfaces(server.Id)
					if err != nil {
						utility.LogError(err, "list server interfaces failed:", false)
						time.Sleep(time.Second * 1)
						continue
					}
					detached := true
					for _, vif := range interfaces {
						if vif.PortId == p.Id {
							detached = false
							break
						}
					}
					if detached {
						logging.Info("[port: %s] detached", p.Id)
						break
					} else {
						time.Sleep(time.Second * 5)
					}
				}
				return nil
			},
		}
		logging.Info("detaching ...")
		taskGroup3.Start()
		taskGroup4 := syncutils.TaskGroup{
			Items:        ports,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(neutron.Port)
				err := client.NeutronV2().Ports().Delete(p.Id)
				if err != nil {
					logging.Error("[port: %s] delete failed: %v", p.Id, err)
					return err
				}
				return nil
			},
		}
		logging.Info("cleaning ...")
		taskGroup4.Start()
	},
}

var interfaceAttachNets = &cobra.Command{
	Use:   "attach-nets <server> <network1> [<network2>...]",
	Short: "Attach nets to server",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		nums, _ := cmd.Flags().GetInt("nums")
		workers, _ := cmd.Flags().GetInt("workers")

		usedNets := args[1:]

		client := openstack.DefaultClient()
		server, err := client.NovaV2().Servers().Show(args[0])
		utility.LogError(err, "show server failed:", true)

		nets := []string{}
		index := 0

		for {
			if len(nets) >= nums {
				break
			}
			nets = append(nets, usedNets[index])
			index++
			if index >= len(usedNets)-1 {
				index = 0
			}
		}
		logging.Debug("attach nets: %s", nets)
		taskGroup := syncutils.TaskGroup{
			Items:     nets,
			MaxWorker: workers,
			Func: func(item interface{}) error {
				p := item.(string)
				logging.Info("[net: %s] attaching", p)
				attachment, err := client.NovaV2().Servers().AddInterface(server.Id, p, "")

				if err != nil {
					logging.Info("[net: %s] attach failed: %v", p, err)
					return err
				}
				interfaces, err := client.NovaV2().Servers().ListInterfaces(server.Id)
				if err != nil {
					utility.LogError(err, "list server interfaces failed:", false)
					return err
				}
				for _, vif := range interfaces {
					if vif.PortId == attachment.PortId {
						logging.Info("[port: %s] attach success", attachment.PortId)
						return nil
					}
				}
				logging.Error("[net: %s] attach failed", p)
				return nil
			},
		}
		taskGroup.Start()
	},
}

func init() {
	interfaceAttachPorts.Flags().Int("nums", 1, "nums of interfaces")
	interfaceAttachPorts.Flags().Int("parallel", 0, "nums of parallel")
	interfaceAttachPorts.Flags().Bool("clean", false, "detach and delete interfaces")

	interfaceAttachNets.Flags().Int("nums", 1, "nums of interfaces")
	interfaceAttachNets.Flags().Int("parallel", 1, "nums of parallel")

	attachCmd.AddCommand(interfaceAttachPorts, interfaceAttachNets)
}
