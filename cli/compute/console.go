package compute

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
)

var Console = &cobra.Command{Use: "console"}

var consoleLog = &cobra.Command{
	Use:   "log <server>",
	Short: "Show console log of server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		lines, _ := cmd.Flags().GetUint("lines")
		consoleLog, err := client.NovaV2().Servers().ConsoleLog(args[0], lines)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(consoleLog.Output)
	},
}

var validType []string

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

		console, err := client.NovaV2().Servers().ConsoleUrl(args[0], args[1])
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
	validType = append(validType, "novnc", "xvpvnc", "rdp-html5",
		"spice-html5", "serial", "webmks", "ssh", "sressh")

	consoleLog.Flags().UintP("lines", "l", 0, "Number of lines to display from the end of the log")

	Console.AddCommand(consoleLog, consoleUrl)
}
