package guest

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/utility"
)

var REG_HOST_UUID, _ = regexp.Compile("((?P<host>.+)@)*(?P<uuid>[^:@]+)")
var REG_HOST_UUID_PATH, _ = regexp.Compile("((?P<host>.+)@)*(?P<uuid>[^:@]+)(:(?P<path>.+))*")

var GuestCommand = &cobra.Command{Use: "guest", Short: "guest tools"}

func getGuest(guestConnector string) (*guest.Guest, error) {
	matched := REG_HOST_UUID.FindStringSubmatch(guestConnector)
	if matched == nil {
		return nil, fmt.Errorf("invalid guest connector: %s", guestConnector)
	}
	var domainHost, domainUUID string
	for i, name := range REG_HOST_UUID.SubexpNames() {
		switch name {
		case "host":
			domainHost = matched[i]
		case "uuid":
			domainUUID = matched[i]
		}
	}
	domainGuest := guest.Guest{
		Connection: domainHost,
		Domain:     domainUUID,
	}
	return &domainGuest, nil
}

var qgaExec = &cobra.Command{
	Use:     "qga-exec <domain> <command>",
	Short:   "执行QGA命令",
	Long:    "执行 Libvirt QGA(qemu-guest-agent) 命令",
	Example: "e.g.\nqga-exec HOST@DOMAIN_UUID 'ls -l'",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		guestConnector, command := args[0], args[1:]
		uuid, _ := cmd.Flags().GetBool("uuid")

		domainGuest, err := getGuest(guestConnector)
		utility.LogError(err, "parse guest failed", true)
		if uuid {
			domainGuest.ByUUID = uuid
		}
		logging.Debug("connect to guest: %s", domainGuest)
		err = domainGuest.Connect()
		utility.LogError(err, "连接domain失败", true)

		execResult := domainGuest.Exec(strings.Join(command, " "), true)
		if execResult.OutData != "" {
			fmt.Println(execResult.OutData)
		}
		if execResult.ErrData != "" {
			fmt.Println(execResult.ErrData)
		}
	},
}

var qgaCopy = &cobra.Command{
	Use:     "qga-copy <domain> <file> <guest path>",
	Short:   "QGA 拷贝文件",
	Long:    "使用 Libvirt QGA 拷贝小文件",
	Example: "qga-copy /the/path/of/file HOST@DOMAIN_UUID:/tmp",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if args[1] == "" {
			return fmt.Errorf("guest path is empty")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		localFile, guestConnectorPath := args[0], args[1]
		var guestConnector, guestPath string
		if values := strings.Split(guestConnectorPath, ":"); len(values) < 2 {
			guestConnector, guestPath = values[0], "/"
		} else {
			guestConnector, guestPath = values[0], values[1]
		}

		domainGuest, err := getGuest(guestConnector)
		utility.LogError(err, "parse guest failed", true)

		logging.Debug("connect to guest: %s", domainGuest)
		err = domainGuest.Connect()
		utility.LogError(err, "连接domain失败", true)

		guestFile, err := domainGuest.CopyFile(localFile, guestPath)
		if err != nil {
			logging.Fatal("copy file failed: %s", err)
		} else {
			logging.Info("the path of file is %s", guestFile)
		}
	},
}

var iperf3Test = &cobra.Command{
	Use:     "iperf3-test <domain> --client <domain>",
	Short:   "测试实例BPS/PPS",
	Long:    "基于 iperf3 工具测试实例的网络BPS/PPS",
	Args:    cobra.ExactArgs(1),
	Example: "guest-bps-test <hostA>@<domain1-uuid> --client <hostB>@<domain2-uuid>",
	Run: func(cmd *cobra.Command, args []string) {
		client, _ := cmd.Flags().GetString("client")
		pps, _ := cmd.Flags().GetBool("pps")
		iperf3Package, _ := cmd.Flags().GetString("iperf3-package")

		serverGuest, err := getGuest(args[0])
		utility.LogError(err, "parse server guest failed", true)

		clientGuest, err := getGuest(client)
		utility.LogError(err, "parse client guest failed", true)

		_, _, err = guest.TestNetQos(*serverGuest, *clientGuest, pps, iperf3Package)
		utility.LogError(err, "测试失败", true)
	},
}

func init() {
	qgaExec.Flags().Bool("uuid", false, "通过 UUID 查找")

	iperf3Test.Flags().StringP("client", "C", "", "客户端实例UUID")
	iperf3Test.Flags().Bool("pps", false, "测试PPS")
	iperf3Test.Flags().String("iperf3-package", "", "iperf3 安装包")
	iperf3Test.MarkFlagRequired("client")

	GuestCommand.AddCommand(qgaExec, qgaCopy, iperf3Test)
}
