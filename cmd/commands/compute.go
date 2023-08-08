package commands

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

var Server = &cobra.Command{Use: "server"}

var ServerList = &cobra.Command{
	Use:   "list",
	Short: "List servers",
	Run: func(cmd *cobra.Command, args []string) {
		computeClient := getComputeClient()

		query := url.Values{}
		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		host, _ := cmd.Flags().GetString("host")
		statusList, _ := cmd.Flags().GetStringArray("status")

		if name != "" {
			query.Set("name", name)
		}
		if host != "" {
			query.Set("host", host)
		}
		for _, status := range statusList {
			query.Add("status", status)
		}

		serversTable := ServersTable{Servers: computeClient.ServerListDetails(query)}
		serversTable.Print(long)
	},
}
var ServerShow = &cobra.Command{
	Use:   "show <name or id>",
	Short: "Show server details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		computeClient := getComputeClient()
		nameOrId := args[0]
		server, _ := computeClient.ServerShow(nameOrId)
		serverTable := ServerTable{Server: server}
		serverTable.Print()
	},
}
var ServerCreate = &cobra.Command{
	Use:   "create",
	Short: "Create server(s)",
	Run: func(cmd *cobra.Command, args []string) {
		logging.Info("list servers")
	},
}
var ServerSet = &cobra.Command{
	Use:   "set",
	Short: "Set server properties",
	Run: func(cmd *cobra.Command, args []string) {
		logging.Info("list servers")
	},
}
var ServerDelete = &cobra.Command{
	Use:   "delete <server1> [server2 ...]",
	Short: "Delete server",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		computeClient := getComputeClient()
		for _, id := range args {
			err := computeClient.ServerDelete(id)
			if err != nil {
				logging.Error("Reqeust to delete server failed, %v", err)
			} else {
				fmt.Printf("Requested to delete server: %s\n", id)
			}

		}
	},
}

func init() {
	ServerList.Flags().StringP("name", "n", "", "Search by server name")
	ServerList.Flags().String("host", "", "Search by hostname")
	ServerList.Flags().StringArrayP("status", "s", nil, "Search by server status")
	ServerList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Server.AddCommand(ServerList)
	Server.AddCommand(ServerShow)
	Server.AddCommand(ServerCreate)
	Server.AddCommand(ServerDelete)
	Server.AddCommand(ServerSet)
}
