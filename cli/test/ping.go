package test

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var serverPing = &cobra.Command{
	Use:   "server-ping <SERVER> <CLIENT>",
	Short: "Run ping from server to client",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		interval, _ := cmd.Flags().GetFloat32("interval")
		count, _ := cmd.Flags().GetInt("count")

		serverInstance, err := client.NovaV2().Server().Show(args[0])
		utility.LogError(err, "get server failed", true)
		clientInstance, err := client.NovaV2().Server().Show(args[1])
		utility.LogError(err, "get client failed", true)

		logging.Debug("get server host and client host ...")
		serverHost, err := client.NovaV2().Hypervisor().Found(serverInstance.Host)
		utility.LogError(err, "get server host failed", true)

		clientHost, err := client.NovaV2().Hypervisor().Found(clientInstance.Host)
		utility.LogError(err, "get client host failed", true)

		serverGuest := guest.Guest{Connection: serverHost.HostIp, Domain: serverInstance.Id}
		clientGuest := guest.Guest{Connection: clientHost.HostIp, Domain: clientInstance.Id}

		serverAddresses := serverGuest.GetIpaddrs()
		clientAddresses := clientGuest.GetIpaddrs()
		logging.Debug("客户端实例IP地址: %s", clientAddresses)
		logging.Debug("服务端实例IP地址: %s", serverAddresses)
		if len(clientAddresses) == 0 || len(serverAddresses) == 0 {
			logging.Fatal("客户端和服务端实例必须至少有一张启用的网卡")
		}

		var stdout, stderr string
		if count > 0 {
			result := serverGuest.Ping(clientAddresses[0], interval, count, serverAddresses[0], count > 0)
			if result.ErrData != "" {
				fmt.Println(result.ErrData)
				return
			}
			stdout, stderr = result.OutData, result.ErrData

		} else {
			stdout, stderr = serverGuest.WithPing(
				clientAddresses[0], interval, serverAddresses[0],
				func() {
					logging.Info("waiting, stop by ctrl+C ...")
					sigCh := make(chan os.Signal, 1)
					signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
					sig := <-sigCh
					logging.Info("received signal: %s\n", sig)
				},
			)
		}

		if stderr != "" {
			fmt.Println(color.RedString(stdout))
			os.Exit(1)
		}
		reg := regexp.MustCompile(`(\d+)% +packet +loss`)
		matchedNoLossed := reg.FindAllStringSubmatch(stdout, -1)
		if len(matchedNoLossed) >= 1 && len(matchedNoLossed[0]) >= 2 {
			lossPackage, _ := strconv.Atoi(matchedNoLossed[0][1])
			switch {
			case lossPackage == 0:
				fmt.Println(stdout)
			case lossPackage == 100:
				fmt.Println(color.RedString(stdout))
			default:
				fmt.Println(color.YellowString(stdout))
			}
			return
		}
		fmt.Println(color.RedString(stdout))
	},
}

func init() {
	serverPing.Flags().Float32("interval", 1.0, "Interval")
	serverPing.Flags().Int("count", 0, "count")
	TestCmd.AddCommand(serverPing)
}
