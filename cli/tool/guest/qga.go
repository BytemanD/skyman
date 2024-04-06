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

var qgaExec = &cobra.Command{
	Use:     "qga-exec <domain> <command>",
	Short:   "执行QGA命令",
	Long:    "执行 Libvirt QGA(qemu-guest-agent) 命令",
	Example: "e.g.\nqga-exec HOST@DOMAIN_UUID 'ls -l'",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		guestConnector, command := args[0], args[1:]
		uuid, _ := cmd.Flags().GetBool("uuid")

		matched := REG_HOST_UUID.FindStringSubmatch(guestConnector)
		if matched == nil {
			logging.Fatal("invalid guest connector: %s", guestConnector)
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
		if uuid {
			domainGuest.ByUUID = uuid
		}
		logging.Debug("connect to guest: %s", domainGuest)
		err := domainGuest.Connect()
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
	Example: "e.g.\nqga-copy /the/path/of/file HOST@DOMAIN_UUID:/tmp",
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
		localFile, guestPath := args[0], args[1]

		matched := REG_HOST_UUID_PATH.FindStringSubmatch(guestPath)
		if matched == nil {
			logging.Fatal("invalid remote path: %s", guestPath)
		}
		var domainHost, domainUUID, domainPath string
		for i, name := range REG_HOST_UUID_PATH.SubexpNames() {
			switch name {
			case "host":
				domainHost = matched[i]
			case "uuid":
				domainUUID = matched[i]
			case "path":
				domainPath = matched[i]
			}
		}
		if domainHost == "" {
			domainHost = "localhost"
		}
		if domainPath == "" {
			domainPath = "/"
		}
		domainGuest := guest.Guest{
			Connection: domainHost,
			Domain:     domainUUID,
		}
		logging.Debug("connect to guest: %s", domainGuest)
		err := domainGuest.Connect()
		utility.LogError(err, "连接domain失败", true)

		guestFile, err := domainGuest.CopyFile(localFile, domainPath)
		if err != nil {
			logging.Fatal("copy file failed: %s", err)
		} else {
			logging.Info("the path of file is %s", guestFile)
		}
	},
}

func init() {
	qgaExec.Flags().Bool("uuid", false, "通过 UUID 查找")

	GuestCommand.AddCommand(qgaExec, qgaCopy)
}
