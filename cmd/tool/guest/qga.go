package guest

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
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
	Example: "qga-exec HOST@DOMAIN_UUID 'ls -l'",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		guestConnector, command := args[0], args[1:]
		uuid, _ := cmd.Flags().GetBool("uuid")

		domainGuest, err := getGuest(guestConnector)
		utility.LogError(err, "parse guest failed", true)
		if uuid {
			domainGuest.ByUUID = uuid
		}
		console.Debug("connect to guest: %s", domainGuest)
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

		console.Debug("connect to guest: %s", domainGuest)
		err = domainGuest.Connect()
		utility.LogError(err, "连接domain失败", true)

		guestFile, err := domainGuest.CopyFile(localFile, guestPath)
		if err != nil {
			console.Error("copy file failed: %s", err)
			os.Exit(1)
		} else {
			console.Info("the path of file is %s", guestFile)
		}
	},
}
var qgaPasswd = &cobra.Command{
	Use:     "qga-passwd <domain> <PASSWORD>",
	Short:   "使用 QGA 修改密码",
	Long:    "使用 chpasswd 命令修改 Linux 密码",
	Example: "qga-passwd <HOST>@<DOMAIN_UUID>",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		connection, password := args[0], args[1]
		user, _ := cmd.Flags().GetString("user")

		domainGuest, err := getGuest(connection)
		utility.LogError(err, "parse guest connection failed", true)
		console.Info("连接 guest domain")
		err = domainGuest.Connect()
		utility.LogError(err, "连接domain失败", true)

		filename := "user-passwd.sh"

		file, err := os.Create(filename)
		utility.LogError(err, "creating file failed: %s", true)

		defer func() {
			file.Close()
			os.Remove(filename)
		}()

		_, err = file.WriteString(fmt.Sprintf("echo '%s:%s' | /usr/sbin/chpasswd\n", user, password))
		utility.LogError(err, "writing to file failed", true)

		guestFile, err := domainGuest.CopyFile(filename, "/tmp")
		utility.LogError(err, "copy to guest failed", true)
		execResult := domainGuest.Exec(fmt.Sprintf("sh %s", guestFile), true)
		if execResult.Failed || execResult.ErrData != "" {
			console.Error("execute failed: %s", execResult.ErrData)
		} else {
			console.Success("设置成功")
		}
	},
}

var iperf3Test = &cobra.Command{
	Use:     "iperf3-test <domain> --client <domain>",
	Short:   "测试实例BPS/PPS",
	Long:    "基于 iperf3 工具测试实例的网络BPS/PPS",
	Args:    cobra.ExactArgs(1),
	Example: "iperf3-test <hostA>@<domain1-uuid> --client <hostB>@<domain2-uuid>",
	Run: func(cmd *cobra.Command, args []string) {
		client, _ := cmd.Flags().GetString("client")
		pps, _ := cmd.Flags().GetBool("pps")
		iperf3Package, _ := cmd.Flags().GetString("iperf3-package")
		serverOptions, _ := cmd.Flags().GetString("server-options")
		cilentOptions, _ := cmd.Flags().GetString("client-options")

		serverGuest, err := getGuest(args[0])
		utility.LogError(err, "parse server guest failed", true)

		clientGuest, err := getGuest(client)
		utility.LogError(err, "parse client guest failed", true)

		job := guest.NetQosTest{
			ClientGuest:     *clientGuest,
			ServerGuest:     *serverGuest,
			PPS:             pps,
			LocalIperf3File: iperf3Package,
			ServerOptions:   serverOptions,
			ClientOptions:   cilentOptions,
		}
		_, _, err = job.Run()
		utility.LogError(err, "测试失败", true)
	},
}

func init() {
	qgaExec.Flags().Bool("uuid", false, "通过 UUID 查找")

	iperf3Test.Flags().StringP("client", "C", "", "客户端实例UUID")
	iperf3Test.Flags().Bool("pps", false, "测试PPS")
	iperf3Test.Flags().String("iperf3-package", "", "iperf3 安装包")
	iperf3Test.MarkFlagRequired("client")

	iperf3Test.Flags().String("server-options", "", "iperf3 服务端参数")
	iperf3Test.Flags().String("client-options", "", "iperf3 客户端参数")

	qgaPasswd.Flags().StringP("user", "u", "root", "用户名")

	GuestCommand.AddCommand(qgaExec, qgaCopy, qgaPasswd, iperf3Test)
}
