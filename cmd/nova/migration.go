package nova

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/common"
)

var (
	migrationListFlags flags.MigrationListFlags
)
var Migration = &cobra.Command{Use: "migration"}

var migrationList = &cobra.Command{
	Use:   "list",
	Short: "List server migrations",
	Run: func(cmd *cobra.Command, _ []string) {
		client := common.DefaultClient()

		query := url.Values{}

		if *migrationListFlags.Status != "" {
			query.Set("status", *migrationListFlags.Status)
		}
		if *migrationListFlags.Host != "" {
			query.Set("host", *migrationListFlags.Host)
		}
		if *migrationListFlags.Instance != "" {
			query.Set("instance_uuid", *migrationListFlags.Instance)
		}
		if *migrationListFlags.Type != "" {
			query.Set("migration_type", *migrationListFlags.Type)
		}
		migrations, err := client.NovaV2().Migration().List(query)
		if err != nil {
			console.Fatal("%s", err)
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
		common.PrintPrettyTable(dataListTable, *migrationListFlags.Long)
	},
}

func init() {
	// migration list flags
	migrationListFlags = flags.MigrationListFlags{
		Host:     migrationList.Flags().String("host", "", "List migration matched by host"),
		Status:   migrationList.Flags().String("status", "", "List migration matched by status"),
		Instance: migrationList.Flags().String("instance", "", "List migration matched by instance uuid"),
		Type:     migrationList.Flags().String("type", "", "List migration matched by migration type"),

		Long: migrationList.Flags().BoolP("long", "l", false, "List additional fields in output"),
	}

	Migration.AddCommand(migrationList)
}
