package cinder

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Backup = &cobra.Command{Use: "backup"}

var backupList = &cobra.Command{
	Use:   "list",
	Short: "List backups",
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
		backups, err := client.CinderV2().Backup().Detail(query)
		utility.LogError(err, "list backup falied", true)
		table := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name"},
				{Name: "Status", AutoColor: true},
				{Name: "Size"},
				{Name: "VolumeId"},
			},
			LongColumns: []common.Column{
				{Name: "Description"},
				{Name: "CreatedAt"},
			},
		}
		table.AddItems(backups)
		if long {
			table.StyleSeparateRows = true
		}
		common.PrintPrettyTable(table, long)
	},
}

var backupShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show backup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		idOrName := args[0]
		backup, err := client.CinderV2().Backup().Find(idOrName)
		utility.LogError(err, "get backup failed", true)
		printBackup(*backup)
	},
}
var backupDelete = &cobra.Command{
	Use:   "delete <backup1> [<backup2> ...]",
	Short: "Delete backup",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		for _, idOrName := range args {
			backup, err := client.CinderV2().Backup().Find(idOrName)
			if err != nil {
				utility.LogError(err, "get backup failed", false)
				continue
			}
			err = client.CinderV2().Backup().Delete(backup.Id)
			if err == nil {
				fmt.Printf("Requested to delete backup %s\n", idOrName)
			} else {
				fmt.Println(err)
			}
		}
	},
}

var backupCreate = &cobra.Command{
	Use:   "create <volume>",
	Short: "Create backup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		name, _ := cmd.Flags().GetString("name")

		client := openstack.DefaultClient()

		volume, err := client.CinderV2().Volume().Find(args[0])
		utility.LogIfError(err, true, "get volume %s failed", args[0])

		backup, err := client.CinderV2().Backup().Create(volume.Id, name, force)
		utility.LogIfError(err, true, "create backup failed")
		backup, err = client.CinderV2().Backup().Show(backup.Id)
		utility.LogIfError(err, true, "show backup failed")
		printBackup(*backup)
	},
}

func init() {
	backupList.Flags().BoolP("long", "l", false, "List additional fields in output")
	backupList.Flags().Bool("all", false, "List backups of all tenants")
	backupList.Flags().StringP("name", "n", "", "Search by backup name")
	backupList.Flags().String("status", "", "Search by backup status")

	backupCreate.Flags().Bool("force", false, "Ignores the current status of the volume ")
	backupCreate.Flags().StringP("name", "n", "", "backup name")

	Backup.AddCommand(
		backupList, backupShow, backupCreate,
		backupDelete,
	)
}
