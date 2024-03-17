package compute

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var serverInspect = &cobra.Command{
	Use:   "inspect <id>",
	Short: "inspect server ",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		serverId := args[0]
		format, _ := cmd.Flags().GetString("format")

		serverInspect, err := client.ServerInspect(serverId)
		utility.LogError(err, "inspect sever faield", true)

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
		c := openstack.DefaultClient()
		regions, err := c.KeystoneV3().Regions().List(nil)
		utility.LogError(err, "get regions failed", true)
		for _, region := range regions {
			logging.Info("try to find server in region '%s'", region.Id)
			client2 := openstack.Client(region.Id).NovaV2()
			server, err := client2.Servers().Found(args[0])
			if err != nil {
				continue
			}
			logging.Info("found server in region '%s'", region.Id)
			printServer(*server)
			break
		}
	},
}
