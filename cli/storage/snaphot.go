package storage

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Snapshot = &cobra.Command{Use: "snapshot"}

var snapshotList = &cobra.Command{
	Use:   "list",
	Short: "List snapshots",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := openstack.DefaultClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		status, _ := cmd.Flags().GetString("status")
		all, _ := cmd.Flags().GetBool("all")

		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		if status != "" {
			query.Set("status", status)
		}
		if all {
			query.Set("all_tenants", "true")
		}
		snapshots, err := client.CinderV2().Snapshot().List(query)
		utility.LogError(err, "list snapshot falied", true)
		table := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Status", AutoColor: true},
			},
			LongColumns: []common.Column{},
		}
		table.AddItems(snapshots)
		if long {
			table.StyleSeparateRows = true
		}
		common.PrintPrettyTable(table, long)
	},
}

var snapshotShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		idOrName := args[0]
		snapshot, err := client.CinderV2().Snapshot().Found(idOrName)
		utility.LogError(err, "get snapshot failed", true)
		printSnapshot(*snapshot)
	},
}
var snapshotDelete = &cobra.Command{
	Use:   "delete <snapshot1> [<snapshot2> ...]",
	Short: "Delete snapshot",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		for _, idOrName := range args {
			snapshot, err := client.CinderV2().Snapshot().Found(idOrName)
			if err != nil {
				utility.LogError(err, "get snapshot failed", false)
				continue
			}
			err = client.CinderV2().Snapshot().Delete(snapshot.Id)
			if err == nil {
				fmt.Printf("Requested to delete snapshot %s\n", idOrName)
			} else {
				fmt.Println(err)
			}
		}
	},
}

// var snapshotCreate = &cobra.Command{
// 	Use:   "create <name>",
// 	Short: "Create snapshot",
// 	Args:  cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		size, _ := cmd.Flags().GetUint("size")
// 		snapshotType, _ := cmd.Flags().GetString("type")

// 		params := map[string]interface{}{
// 			"name": args[0],
// 		}
// 		if size > 0 {
// 			params["size"] = size
// 		}
// 		if snapshotType != "" {
// 			params["snapshot_type"] = snapshotType
// 		}

// 		client := openstack.DefaultClient()

// 		snapshot, err := client.CinderV2().Snapshot().Create(params)
// 		if err != nil {
// 			fmt.Println(err)
// 			os.Exit(1)
// 		}
// 		printSnapshot(*snapshot)
// },
// }

func init() {
	snapshotList.Flags().BoolP("long", "l", false, "List additional fields in output")
	snapshotList.Flags().Bool("all", false, "List snapshots of all tenants")
	snapshotList.Flags().StringP("name", "n", "", "Search by snapshot name")
	snapshotList.Flags().String("status", "", "Search by snapshot status")

	Snapshot.AddCommand(
		snapshotList, snapshotShow,
		snapshotDelete,
	)
}
