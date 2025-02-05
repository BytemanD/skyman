package server

import (
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/cmd/views"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var serverFind = &cobra.Command{
	Use:   "find <id or name>",
	Short: "Find server in all regions",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		c := common.DefaultClient()
		regions, err := c.KeystoneV3().Region().List(nil)
		utility.LogError(err, "get regions failed", true)
		var server *nova.Server
		for _, region := range regions {
			console.Info("try to find server in region '%s'", region.Id)
			server, err = c.SetRegion(region.Id).NovaV2().Server().Find(args[0])
			if err != nil {
				console.Warn("server %s not found in region %s: %s", args[0], region.Id, err)
				continue
			}
			console.Info("found server %s in region '%s'", args[0], region.Id)
			views.PrintServer(*server, c)
			return
		}
		console.Fatal("server %s not found in all regions", args[0])
	},
}

func init() {
	ServerCommand.AddCommand(serverFind)
}
