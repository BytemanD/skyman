package test

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/utility"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var ServerPing = &cobra.Command{
	Use:   "server-ping <SERVER> <CLIENT>",
	Short: "Run ping from server to client",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()

		interval, _ := cmd.Flags().GetFloat32("interval")
		count, _ := cmd.Flags().GetInt("count")

		serverInstance, err := client.NovaV2().Server().Find(args[0])
		utility.LogError(err, "get server failed", true)
		clientInstance, err := client.NovaV2().Server().Find(args[1])
		utility.LogError(err, "get client failed", true)

		console.Debug("get server host and client host ...")
		serverHost, err := client.NovaV2().Hypervisor().Find(serverInstance.Host)
		utility.LogError(err, "get server host failed", true)

		clientHost, err := client.NovaV2().Hypervisor().Find(clientInstance.Host)
		utility.LogError(err, "get client host failed", true)

		serverGuest := guest.Guest{Connection: serverHost.HostIp, Domain: serverInstance.Id}
		clientGuest := guest.Guest{Connection: clientHost.HostIp, Domain: clientInstance.Id}

		serverAddresses := serverGuest.GetIpaddrs()
		clientAddresses := clientGuest.GetIpaddrs()
		console.Debug("客户端实例IP地址: %s", clientAddresses)
		console.Debug("服务端实例IP地址: %s", serverAddresses)
		if len(clientAddresses) == 0 || len(serverAddresses) == 0 {
			console.Fatal("客户端和服务端实例必须至少有一张启用的网卡")
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
					console.Info("waiting, stop by ctrl+C ...")
					sigCh := make(chan os.Signal, 1)
					signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
					sig := <-sigCh
					console.Info("received signal: %s\n", sig)
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
	ServerPing.Flags().Float32("interval", 1.0, "Interval")
	ServerPing.Flags().Int("count", 0, "count")
}
