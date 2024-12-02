package compute

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cli/flags"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var (
	consoleLogFlags flags.ConsoleLogFlags
)
var Console = &cobra.Command{Use: "console"}

var consoleLog = &cobra.Command{
	Use:   "log <server>",
	Short: "Show console log of server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		server, err := client.NovaV2().Server().Found(args[0])
		utility.LogError(err, "get server failed", true)
		consoleLog, err := client.NovaV2().Server().ConsoleLog(server.Id, *consoleLogFlags.Lines)
		utility.LogError(err, "get console log failed", true)
		fmt.Println(consoleLog.Output)
	},
}

var validType = []string{
	"novnc", "xvpvnc", "rdp-html5",
	"spice-html5", "serial", "webmks", "ssh", "sressh",
}

var consoleUrl = &cobra.Command{
	Use:   "url <server> <type>",
	Short: "Show console url of server",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		var isTypeValid bool
		for _, item := range validType {
			if args[1] == item {
				isTypeValid = true
				break
			}
		}
		if !isTypeValid {
			fmt.Printf("invalid type: %s, supported types: %v\n", args[1], validType)
			os.Exit(1)
		}

		server, err := client.NovaV2().Server().Found(args[0])
		utility.LogError(err, "get server failed", true)
		console, err := client.NovaV2().Server().ConsoleUrl(server.Id, args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pt := common.PrettyItemTable{
			Item:        *console,
			ShortFields: []common.Column{{Name: "Type"}, {Name: "Url"}},
		}
		common.PrintPrettyItemTable(pt)
	},
}

func init() {
	consoleLogFlags = flags.ConsoleLogFlags{
		Lines: consoleLog.Flags().UintP("lines", "l", 0, "Number of lines to display from the end of the log"),
	}

	Console.AddCommand(consoleLog, consoleUrl)
}
