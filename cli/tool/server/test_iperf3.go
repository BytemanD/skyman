package server

import (
	"fmt"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/table"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

const SERVER_IPERF3_TEST_EXAMPLE = `
## 客户端向服务端打流:
  skyman tool server iperf3-test <服务端虚拟机 UUID> --client <客户端虚拟机 UUID>

## 设置iperf3 客户端参数
1. 设置打流时间为60s, 编辑配置文件(默认为 /etc/skyman/skyman.yaml), 设置如下:
  +------------------------------------------------------+
  | iperf:                                               |
  |   clientOptions: -t 60                               |
  +------------------------------------------------------+
2. 设置打流时间为60s, 并发数为10, 编辑配置文件, 设置如下:
  +------------------------------------------------------+
  | iperf:                                               |
  |   clientOptions: -t 60 -P 10                         |
  +------------------------------------------------------+
`

type QosItem struct {
	Position string
	Value    string
	BpsBurst string
	BpsPeak  string
	PPSBurst string
	PPSPeak  string
}

func printServerQOSItems(server nova.Server) {
	items := []QosItem{
		{
			Position: "入向",
			BpsBurst: server.Flavor.ExtraSpecs.Get("quota:vif_inbound_burst"),
			BpsPeak:  server.Flavor.ExtraSpecs.Get("quota:vif_inbound_peak"),
			PPSBurst: server.Flavor.ExtraSpecs.Get("quota:vif_inbound_pps_burst"),
			PPSPeak:  server.Flavor.ExtraSpecs.Get("quota:vif_inbound_pps_burst"),
		},
		{
			Position: "出向",
			BpsBurst: server.Flavor.ExtraSpecs.Get("quota:vif_inbound_burst"),
			BpsPeak:  server.Flavor.ExtraSpecs.Get("quota:vif_inbound_peak"),
			PPSBurst: server.Flavor.ExtraSpecs.Get("quota:vif_inbound_pps_burst"),
			PPSPeak:  server.Flavor.ExtraSpecs.Get("quota:vif_inbound_pps_burst"),
		},
	}
	t := table.NewItemsTable([]string{"Position", "BpsBurst", "BpsPeak", "PPSBurst", "PPSPeak"}, items)
	result, _ := t.SetStyle(table.StyleLight).Render()
	fmt.Println(result)
}

var testNetQos = &cobra.Command{
	Use:     "iperf3-test <server> --client <client>",
	Short:   i18n.T("testServerNetworkQOS"),
	Long:    "基于iperf3工具测试两个虚拟机的网络QOS",
	Example: strings.TrimRight(SERVER_IPERF3_TEST_EXAMPLE, "\n"),
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		server := args[0]
		client, _ := cmd.Flags().GetString("client")

		pps, _ := cmd.Flags().GetBool("pps")
		localIperf3File, _ := cmd.Flags().GetString("iperf3-package")

		openstackClient := openstack.DefaultClient()
		logging.Info("get server and client")
		serverInstance, err := openstackClient.NovaV2().Servers().Show(server)
		utility.LogError(err, "get server failed", true)
		clientInstance, err := openstackClient.NovaV2().Servers().Show(client)
		utility.LogError(err, "get client failed", true)

		if !serverInstance.IsActive() {
			logging.Fatal("instance %s is not active", serverInstance.Id)
		}
		if !clientInstance.IsActive() {
			logging.Fatal("instance %s is not active", clientInstance.Id)
		}

		logging.Info("get server host and client host")
		serverHost, err := openstackClient.NovaV2().Hypervisors().Found(serverInstance.Host)
		utility.LogError(err, "get server host failed", true)

		clientHost, err := openstackClient.NovaV2().Hypervisors().Found(clientInstance.Host)
		utility.LogError(err, "get client host failed", true)

		serverConn := guest.Guest{Connection: serverHost.HostIp, Domain: serverInstance.Id}
		clientConn := guest.Guest{Connection: clientHost.HostIp, Domain: clientInstance.Id}

		fmt.Println("服务端QOS配置:")
		printServerQOSItems(*serverInstance)
		fmt.Println("客户端QOS配置:")
		printServerQOSItems(*clientInstance)

		logging.Info("start test with QGA")
		_, _, err = guest.TestNetQos(clientConn, serverConn, pps, localIperf3File)
		if err != nil {
			logging.Fatal("test failed, %s", err)
		}
	},
}

func init() {
	testNetQos.Flags().StringP("client", "C", "", "客户端实例")
	testNetQos.Flags().Bool("pps", false, "测试PPS")
	testNetQos.Flags().String("iperf3-package", "", "iperf3 安装包")

	testNetQos.MarkFlagRequired("client")

	ServerCommand.AddCommand(testNetQos)
}
