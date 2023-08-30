package compute

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
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
		migration_type, _ := cmd.Flags().GetString("type")
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
		if migration_type != "" {
			query.Set("migration_type", migration_type)
		}
		migrations, err := client.Compute.MigrationList(query)
		if err != nil {
			logging.Fatal("%s", err)
		}
		dataListTable := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id", Sort: true}, {Name: "MigrationType", Text: "Type"},
				{Name: "Status", AutoColor: true},
				{Name: "SourceNode"},
				{Name: "DestNode"}, {Name: "DestCompute"},
				{Name: "InstanceUUID", Text: "Instance UUID"},
			},
			LongColumns: []common.Column{
				{Name: "DestHost"},
				{Name: "OldInstanceTypeId", Text: "Old Flavor"},
				{Name: "NewInstanceTypeId", Text: "New Flavor"},
				{Name: "SourceRegion"}, {Name: "DestRegion"},
				{Name: "CreatedAt"}, {Name: "UpdatedAt"},
			},
		}
		dataListTable.AddItems(migrations)
		common.PrintPrettyTable(dataListTable, long)
	},
}

func init() {
	// migration list flags
	migrationList.Flags().String("host", "", "List migration matched by host")
	migrationList.Flags().String("status", "", "List migration matched by status")
	migrationList.Flags().String("instance", "", "List migration matched by instance uuid")
	migrationList.Flags().String("type", "", "List migration matched by migration type")

	migrationList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Migration.AddCommand(migrationList)
}
