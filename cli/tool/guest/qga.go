package guest

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/guest"
	"github.com/BytemanD/skyman/utility"
)

var GuestCommand = &cobra.Command{Use: "guest", Short: "guest tools"}

var qgaExec = &cobra.Command{
	Use:   "qga-exec <domain> <command>",
	Short: "执行QGA命令",
	Long:  "执行 Libvirt QGA(qemu-guest-agent) 命令",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		guestConnector, command := args[0], args[1]
		uuid, _ := cmd.Flags().GetBool("uuid")

		domainGuest, err := guest.ParseGuest(guestConnector)
		utility.LogError(err, "parse guest connection failed", true)
		if uuid {
			domainGuest.ByUUID = uuid
		}
		logging.Debug("connect to guest: %s", guestConnector)
		err = domainGuest.Connect()
		if err != nil {
			logging.Error("连接domain失败 %s", err)
			return
		}
		execResult := domainGuest.Exec(command, true)
		if execResult.OutData != "" {
			fmt.Println(execResult.OutData)
		}
		if execResult.ErrData != "" {
			fmt.Println(execResult.ErrData)
		}
	},
}

func init() {
	qgaExec.Flags().Bool("uuid", false, "通过 UUID 查找")

	GuestCommand.AddCommand(qgaExec)
}
