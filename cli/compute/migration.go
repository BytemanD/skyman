package compute

import (
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/openstack/compute"
)

var Migration = &cobra.Command{Use: "migration"}

var migrationList = &cobra.Command{
	Use:   "list",
	Short: "List server migrations",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		query := url.Values{}
		status, _ := cmd.Flags().GetString("status")
		host, _ := cmd.Flags().GetString("host")
		instance, _ := cmd.Flags().GetString("instance")
		long, _ := cmd.Flags().GetBool("long")

		if status != "" {
			query.Set("status", status)
		}
		if host != "" {
			query.Set("host", host)
		}
		if instance != "" {
			query.Set("instance_uuid", instance)
		}

		migrations, err := client.Compute.MigrationList(query)
		if err != nil {
			logging.Fatal("%s", err)
		}
		dataTable := cli.DataListTable{
			ShortHeaders: []string{
				"Id", "MigrationType", "Status", "SourceNode", "SourceCompute",
				"DestNode", "DestCompute", "InstanceUUID",
			},
			LongHeaders: []string{
				"DestHost", "OldInstanceTypeId", "NewInstanceTypeId",
				"CreatedAt", "UpdatedAt"},
			HeaderLabel: map[string]string{
				"MigrationType":     "Type",
				"SourceNode":        "Source Node",
				"SourceCompute":     "Source Compute",
				"DestNode":          "Dest Node",
				"DestCompute":       "Deest Compute",
				"DestHost":          "Dest Host",
				"InstanceUUID":      "Instance UUID",
				"OldInstanceTypeId": "Old Flavor",
				"NewInstanceTypeId": "New Flavor",
				"CreatedAt":         "Created At",
				"UpdatedAt":         "Updated At",
			},
			SortBy: []table.SortBy{{Name: "Id", Mode: table.Asc}},
			Slots: map[string]func(item interface{}) interface{}{
				"Status": func(item interface{}) interface{} {
					p, _ := item.(compute.Migration)
					return cli.BaseColorFormatter.Format(p.Status)
				},
			},
		}
		dataTable.AddItems(migrations)
		dataTable.Print(long)
	},
}

func init() {
	// migration list flags
	migrationList.Flags().String("host", "", "List migration matched by host")
	migrationList.Flags().String("status", "", "List migration matched by status")
	migrationList.Flags().String("instance", "", "List migration matched by instance uuid")

	migrationList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Migration.AddCommand(migrationList)
}
