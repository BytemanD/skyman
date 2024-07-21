package server

import (
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/views"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var serverFind = &cobra.Command{
	Use:   "find <id or name>",
	Short: "Find server in all regions",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		c := openstack.DefaultClient()
		regions, err := c.KeystoneV3().Region().List(nil)
		utility.LogError(err, "get regions failed", true)
		for _, region := range regions {
			logging.Info("try to find server in region '%s'", region.Id)
			client2 := openstack.Client(region.Id).NovaV2()
			server, err := client2.Server().Found(args[0])
			if err != nil {
				continue
			}
			logging.Info("found server in region '%s'", region.Id)
			views.PrintServer(*server)
			break
		}
	},
}

func init() {
	ServerCommand.AddCommand(serverFind)
}
