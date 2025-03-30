package nova

import (
	"fmt"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/common"
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
		client := common.DefaultClient()

		server, err := client.NovaV2().FindServer(args[0])
		utility.LogError(err, "get server failed", true)
		consoleLog, err := client.NovaV2().GetServerConsoleLog(server.Id, *consoleLogFlags.Lines)
		utility.LogError(err, "get console log failed", true)
		println(consoleLog.Output)
	},
}

var validTypes = []string{
	"novnc", "xvpvnc", "rdp-html5",
	"spice-html5", "serial", "webmks", "ssh", "sressh",
}

var consoleUrl = &cobra.Command{
	Use:   "url <server> <type>",
	Short: "Show console url of server",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()
		if !lo.Contains(validTypes, args[1]) {
			fmt.Printf("invalid type: %s, supported types: %v\n", args[1], validTypes)
			os.Exit(1)
		}

		server, err := client.NovaV2().FindServer(args[0])
		utility.LogError(err, "get server failed", true)
		console, err := client.NovaV2().GetServerConsoleUrl(server.Id, args[1])
		if err != nil {
			println(err)
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
