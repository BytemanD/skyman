package test

import (
	"fmt"
	"regexp"
	"strconv"

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

		timeout, _ := cmd.Flags().GetInt("timeout")
		count, _ := cmd.Flags().GetInt("count")

		serverInstance, err := client.NovaV2().Servers().Show(args[0])
		utility.LogError(err, "get server failed", true)
		clientInstance, err := client.NovaV2().Servers().Show(args[1])
		utility.LogError(err, "get client failed", true)

		logging.Debug("get server host and client host ...")
		serverHost, err := client.NovaV2().Hypervisors().Found(serverInstance.Host)
		utility.LogError(err, "get server host failed", true)

		clientHost, err := client.NovaV2().Hypervisors().Found(clientInstance.Host)
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

		logging.Debug("ping %s -> %s", serverAddresses[0], clientAddresses[0])
		result := serverGuest.Ping(clientAddresses[0], timeout, count, serverAddresses[0])

		if result.ErrData != "" {
			fmt.Println(result.ErrData)
			return

		}
		reg := regexp.MustCompile(`(\d+)% +packet +loss`)
		matchedNoLossed := reg.FindAllStringSubmatch(result.OutData, -1)
		if len(matchedNoLossed) >= 1 && len(matchedNoLossed[0]) >= 2 {
			lossPackage, _ := strconv.Atoi(matchedNoLossed[0][1])
			switch {
			case lossPackage == 0:
				fmt.Println(color.YellowString(result.OutData))
			case lossPackage == 100:
				fmt.Println(color.RedString(result.OutData))
			default:
				fmt.Println(result.OutData)
			}
			return
		}
		fmt.Println(color.RedString(result.OutData))
	},
}

func init() {
	serverPing.Flags().Int("interval", 1, "Interval")
	serverPing.Flags().Int("count", 10, "count")
	TestCmd.AddCommand(serverPing)
}
