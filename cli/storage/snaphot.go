package storage

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/cinder"
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
				{Name: "Id"}, {Name: "Name"},
				{Name: "Status", AutoColor: true},
				{Name: "Size"},
				{Name: "VolumeId", Text: "Volume", Slot: func(item interface{}) interface{} {
					p, _ := item.(cinder.Snapshot)
					if p.VolumeId == "" {
						return ""
					}
					if vol, err := client.CinderV2().Volume().Show(p.VolumeId); err != nil {
						logging.Warning("get volume %s failed: %s", p.VolumeId, err)
						return p.VolumeId
					} else {
						return vol.NameOrId()
					}
				}},
			},
			LongColumns: []common.Column{
				{Name: "Description"},
				{Name: "CreatedAt"},
			},
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

var snapshotCreate = &cobra.Command{
	Use:   "create <volume>",
	Short: "Create snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		name, _ := cmd.Flags().GetString("name")

		client := openstack.DefaultClient()

		volume, err := client.CinderV2().Volume().Found(args[0])
		utility.LogIfError(err, true, "get volume %s failed", args[0])

		snapshot, err := client.CinderV2().Snapshot().Create(volume.Id, name, force)
		utility.LogIfError(err, true, "create snaphost failed")
		printSnapshot(*snapshot)
	},
}

func init() {
	snapshotList.Flags().BoolP("long", "l", false, "List additional fields in output")
	snapshotList.Flags().Bool("all", false, "List snapshots of all tenants")
	snapshotList.Flags().StringP("name", "n", "", "Search by snapshot name")
	snapshotList.Flags().String("status", "", "Search by snapshot status")

	snapshotCreate.Flags().Bool("force", false, "Ignores the current status of the volume ")
	snapshotCreate.Flags().StringP("name", "n", "", "snapshot name")

	Snapshot.AddCommand(
		snapshotList, snapshotShow, snapshotCreate,
		snapshotDelete,
	)
}
