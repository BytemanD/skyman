package test

import (
	"strings"

	"github.com/BytemanD/easygo/pkg/table"
	"github.com/BytemanD/go-console/console"
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
  skyman test server-iperf3 <服务端虚拟机> --client <客户端虚拟机> --client-options '-t 60'
  
2. 设置打流时间为60s, 并发数为10, 编辑配置文件, 设置如下:
  skyman test server-iperf3 <服务端虚拟机> --client <客户端虚拟机> --client-options '-t 60 -P 10'
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
	println(result)
}

var TestNetQos = &cobra.Command{
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
		serverOptions, _ := cmd.Flags().GetString("server-options")
		cilentOptions, _ := cmd.Flags().GetString("client-options")

		openstackClient := openstack.DefaultClient()
		console.Info("get server and client")
		serverInstance, err := openstackClient.NovaV2().Server().Find(server)
		utility.LogError(err, "get server failed", true)
		clientInstance, err := openstackClient.NovaV2().Server().Find(client)
		utility.LogError(err, "get client failed", true)

		if !serverInstance.IsActive() {
			console.Fatal("instance %s is not active", serverInstance.Id)
		}
		if !clientInstance.IsActive() {
			console.Fatal("instance %s is not active", clientInstance.Id)
		}

		console.Info("get server host and client host")
		serverHost, err := openstackClient.NovaV2().Hypervisor().Find(serverInstance.Host)
		utility.LogError(err, "get server host failed", true)

		clientHost, err := openstackClient.NovaV2().Hypervisor().Find(clientInstance.Host)
		utility.LogError(err, "get client host failed", true)

		serverConn := guest.Guest{Connection: serverHost.HostIp, Domain: serverInstance.Id}
		clientConn := guest.Guest{Connection: clientHost.HostIp, Domain: clientInstance.Id}

		println("服务端QOS配置:")
		printServerQOSItems(*serverInstance)
		println("客户端QOS配置:")
		printServerQOSItems(*clientInstance)

		console.Info("start test with QGA")

		job := guest.NetQosTest{
			ClientGuest:     clientConn,
			ServerGuest:     serverConn,
			PPS:             pps,
			LocalIperf3File: localIperf3File,
			ServerOptions:   serverOptions,
			ClientOptions:   cilentOptions,
		}
		_, _, err = job.Run()
		if err != nil {
			console.Fatal("test failed, %s", err)
		}
	},
}

func init() {
	TestNetQos.Flags().StringP("client", "C", "", "客户端实例")
	TestNetQos.Flags().Bool("pps", false, "测试PPS")
	TestNetQos.Flags().String("iperf3-package", "", "iperf3 安装包")
	TestNetQos.Flags().String("server-options", "", "iperf3 服务端参数")
	TestNetQos.Flags().String("client-options", "", "iperf3 客户端参数")

	TestNetQos.MarkFlagRequired("client")
}
