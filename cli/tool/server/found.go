package server

import (
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/views"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
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
		var server *nova.Server
		for _, region := range regions {
			logging.Info("try to find server in region '%s'", region.Id)
			c.AuthPlugin.SetRegion(region.Id)
			server, err = c.NovaV2().Server().Found(args[0])
			if err != nil {
				logging.Warning("server %s not found in region %s: %s", args[0], region.Id, err)
				continue
			}
			logging.Info("found server in region '%s'", region.Id)
			break
		}
		if server != nil {
			views.PrintServer(*server, c)
		} else {
			logging.Fatal("server %s not found in all regions", args[0])
		}
	},
}

func init() {
	ServerCommand.AddCommand(serverFind)
}
