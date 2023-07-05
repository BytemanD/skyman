package commands

import (
	"net/url"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
		if name != "" {
			query.Set("name", name)
		}

		tableWriter := table.NewWriter()
		header := table.Row{
			"ID", "Name", "Status", "Task State", "Power State", "Networks",
		}
		if long {
			header = append(header, "AZ")
			header = append(header, "Host")
			header = append(header, "InstanceName")
		}

		tableWriter.AppendHeader(header)
		tableWriter.SetOutputMirror(os.Stdout)
		// tableWriter.SetStyle(table.StyleLight)
		tableWriter.Style().Format.Header = text.FormatDefault
		for _, server := range computeClient.ServerListDetails(query) {
			row := table.Row{
				server.Id, server.Name, server.Status,
				server.GetTaskState(), server.GetPowerState(),
				strings.Join(server.GetNetworks(), "\n"),
			}
			if long {
				row = append(row, server.AZ, server.Host, server.InstanceName)
			}
			tableWriter.AppendRow(row)
		}
		tableWriter.Render()

	},
}
var ServerShow = &cobra.Command{
	Use:   "Show",
	Short: "Show server details",
	Run: func(cmd *cobra.Command, args []string) {
		logging.Info("list servers")
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
	Use:   "delete",
	Short: "Delete server",
	Run: func(cmd *cobra.Command, args []string) {
		logging.Info("list servers")
	},
}

func init() {
	ServerList.Flags().StringP("name", "n", "", "server name")
	ServerList.Flags().BoolP("long", "l", false, "server name")

	Server.AddCommand(ServerList)
	Server.AddCommand(ServerCreate)
	Server.AddCommand(ServerDelete)
	Server.AddCommand(ServerSet)
}
