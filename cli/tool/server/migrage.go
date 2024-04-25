package server

import (
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var migrate = &cobra.Command{
	Use:   "migrate <server>",
	Short: "migrate server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverId := args[0]

		times, _ := cmd.Flags().GetInt("times")
		interval, _ := cmd.Flags().GetInt("inerval")
		live, _ := cmd.Flags().GetBool("live")

		client := openstack.DefaultClient()

		server, err := client.NovaV2().Servers().Show(serverId)
		utility.LogError(err, "show server failed:", true)

		for i := 1; i <= times; i++ {
			sourceHost := server.Host
			startTime := time.Now()
			logging.Info("[server: %s] migrating(%d), source host: %s ...", server.Id, i, sourceHost)
			if live {
				// TODO: block migration
				err = client.NovaV2().Servers().LiveMigrate(serverId, "auto", "")
			} else {
				err = client.NovaV2().Servers().Migrate(serverId, "")
			}
			if err != nil {
				logging.Error("[server: %s] request to migrate failed %v.", server.Id, err)
				break
			}
			for {
				server, err = client.NovaV2().Servers().Show(serverId)
				if err == nil {
					if server.Host != sourceHost {
						logging.Info("[server: %s] migrated, %s -> %s, used: %v",
							server.Id, sourceHost, server.Host, time.Since(startTime))
						break
					}
					logging.Info("[server: %s] migrate progress: %v", server.Id, server.Progress)
					if server.IsError() {
						logging.Error("[server: %s] status is error", server.Id)
						break
					}
				} else {
					logging.Info("[server: %s] refresh error: %s", server.Id, err)
				}
				time.Sleep(time.Second * 5)
			}
			if server.IsError() {
				break
			}
			time.Sleep(time.Second * time.Duration(interval))
		}
	},
}

func init() {
	migrate.Flags().Int("times", 1, "Migrate times")
	migrate.Flags().Int("interval", 1, "Migrate interval")
	migrate.Flags().Bool("live", false, "Live migrate")

	ServerCommand.AddCommand(migrate)
}
