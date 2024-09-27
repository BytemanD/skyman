package test

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
  skyman test server-iperf3 <服务端虚拟机> --client <客户端虚拟机>

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
	BpsBurst string
	BpsPeak  string
	PPSBurst string
	PPSPeak  string
}

const (
	QUOTA_VIF_INBOUND_PEAK      = "quota:vif_inbound_peak"
	QUOTA_VIF_INBOUND_BURST     = "quota:vif_inbound_burst"
	QUOTA_VIF_INBOUND_PPS_PEAK  = "quota:vif_inbound_pps_peak"
	QUOTA_VIF_INBOUND_PPS_PEADK = "quota:vif_inbound_pps_burst"

	QUOTA_VIF_OUTBOUND_PEAK     = "quota:vif_outbound_peak"
	QUOTA_VIF_OUTBOUND_BURST    = "quota:vif_outbound_burst"
	QUOTA_VIF_OUTBOUND_PPS_PEAK = "quota:vif_outbound_pps_peak"
	QUOTA_VIF_OUTBOUND_PPS_PEAD = "quota:vif_outbound_pps_burst"
)

func printServerQOSItems(server nova.Server) {
	items := []QosItem{
		{
			Position: "入向",
			BpsPeak:  server.Flavor.ExtraSpecs.Get(QUOTA_VIF_INBOUND_PEAK),
			BpsBurst: server.Flavor.ExtraSpecs.Get(QUOTA_VIF_INBOUND_BURST),
			PPSPeak:  server.Flavor.ExtraSpecs.Get(QUOTA_VIF_INBOUND_PPS_PEAK),
			PPSBurst: server.Flavor.ExtraSpecs.Get(QUOTA_VIF_INBOUND_PPS_PEADK),
		},
		{
			Position: "出向",
			BpsPeak:  server.Flavor.ExtraSpecs.Get(QUOTA_VIF_OUTBOUND_PEAK),
			BpsBurst: server.Flavor.ExtraSpecs.Get(QUOTA_VIF_OUTBOUND_BURST),
			PPSPeak:  server.Flavor.ExtraSpecs.Get(QUOTA_VIF_OUTBOUND_PPS_PEAK),
			PPSBurst: server.Flavor.ExtraSpecs.Get(QUOTA_VIF_OUTBOUND_PPS_PEAD),
		},
	}
	t := table.NewItemsTable([]string{"Position", "BpsBurst", "BpsPeak", "PPSBurst", "PPSPeak"}, items)
	result, _ := t.SetStyle(table.StyleLight).Render()
	fmt.Println(result)
}

var testNetQos = &cobra.Command{
	Use:     "server-iperf3 <server> --client <client>",
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
		serverInstance, err := openstackClient.NovaV2().Server().Found(server)
		utility.LogError(err, "get server failed", true)
		clientInstance, err := openstackClient.NovaV2().Server().Found(client)
		utility.LogError(err, "get client failed", true)

		if !serverInstance.IsActive() {
			logging.Fatal("instance %s is not active", serverInstance.Id)
		}
		if !clientInstance.IsActive() {
			logging.Fatal("instance %s is not active", clientInstance.Id)
		}

		logging.Info("get server host and client host")
		serverHost, err := openstackClient.NovaV2().Hypervisor().Found(serverInstance.Host)
		utility.LogError(err, "get server host failed", true)

		clientHost, err := openstackClient.NovaV2().Hypervisor().Found(clientInstance.Host)
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

	TestCmd.AddCommand(testNetQos)
}
