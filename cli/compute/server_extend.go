package compute

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/spf13/cobra"
)

var serverInspect = &cobra.Command{
	Use:   "inspect <id>",
	Short: "inspect server ",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		serverId := args[0]
		format, _ := cmd.Flags().GetString("format")

		serverInspect, err := client.ServerInspect(serverId)
		common.LogError(err, "inspect sever faield", true)

		switch format {
		case "json":
			output, err := common.GetIndentJson(serverInspect)
			if err != nil {
				logging.Fatal("print json failed, %s", err)
			}
			fmt.Println(output)
		case "yaml":
			output, err := common.GetYaml(serverInspect)
			if err != nil {
				logging.Fatal("print json failed, %s", err)
			}
			fmt.Println(output)
		default:
			serverInspect.Print()
		}
	},
}

var serverFound = &cobra.Command{
	Use:   "find <id or name>",
	Short: "Find server in all regions",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := cli.GetClient()
		regions, err := client.Identity.RegionList()
		common.LogError(err, "get regions failed", true)
		for _, region := range regions {
			logging.Info("try to find server in region '%s'", region.Id)
			client2 := cli.GetClientWithRegion(region.Id)
			computeClient := client2.MustGenerateComputeClient()
			server, err := computeClient.ServerFound(args[0])
			if err != nil {
				continue
			}
			logging.Info("found server in region '%s'", region.Id)
			printServer(*server)
			break
		}
	},
}
