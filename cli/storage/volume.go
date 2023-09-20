package storage

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	openstackCommon "github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/storage"
	"github.com/spf13/cobra"
)

var Volume = &cobra.Command{Use: "volume"}

var volumeList = &cobra.Command{
	Use:   "list",
	Short: "List volumes",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

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
		volumes, err := client.Storage.VolumeListDetail(query)
		if err != nil {
			openstackCommon.RaiseIfError(err, "list volume falied")
		}
		table := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"},
				{Name: "Name"},
				{Name: "Status", AutoColor: true},
				{Name: "Size"},
				{Name: "Bootable"},
				{Name: "Attachments", Slot: func(item interface{}) interface{} {
					obj, _ := (item).(storage.Volume)
					return strings.Join(obj.GetAttachmentList(), "\n")
				}},
			},
			LongColumns: []common.Column{
				{Name: "Metadata", Slot: func(item interface{}) interface{} {
					obj, _ := (item).(storage.Volume)
					return strings.Join(obj.GetMetadataList(), "\n")
				}},
			},
		}
		table.AddItems(volumes)
		if long {
			table.StyleSeparateRows = true
		}
		common.PrintPrettyTable(table, long)
	},
}

var volumeShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		idOrName := args[0]
		volume, err := client.Storage.VolumeFound(idOrName)
		common.LogError(err, "get volume failed", true)
		printVolume(*volume)
	},
}
var volumeDelete = &cobra.Command{
	Use:   "delete <volume1> [<volume2> ...]",
	Short: "Delete volume",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, idOrName := range args {
			volume, err := client.Storage.VolumeFound(idOrName)
			if err != nil {
				common.LogError(err, "get volume failed", false)
				continue
			}
			err = client.Storage.VolumeDelete(volume.Id)
			if err == nil {
				fmt.Printf("Requested to delete volume %s\n", idOrName)
			} else {
				fmt.Println(err)
			}
		}
	},
}
var volumePrune = &cobra.Command{
	// TODO: support volume id or name
	Use:   "prune",
	Short: "Prune volume(s)",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		yes, _ := cmd.Flags().GetBool("yes")
		statusList, _ := cmd.Flags().GetStringArray("status")

		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		for _, status := range statusList {
			query.Add("status", status)
		}
		client := cli.GetClient()
		client.Storage.VolumePrune(query, yes, false)

	},
}
var volumeCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		size, _ := cmd.Flags().GetUint("size")

		params := map[string]interface{}{
			"name": args[0],
		}
		if size > 0 {
			params["size"] = size
		}

		client := cli.GetClient()
		volume, err := client.Storage.VolumeCreate(params)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		printVolume(*volume)
	},
}
var volumeExtend = &cobra.Command{
	Use:   "extend <volume> <new size>",
	Short: "Extend volume",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if _, err := strconv.Atoi(args[1]); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		idOrName := args[0]
		size, _ := strconv.Atoi(args[1])
		client := cli.GetClient()
		volume, err := client.Storage.VolumeFound(idOrName)
		common.LogError(err, "get volume falied", true)

		err = client.Storage.VolumeExtend(volume.Id, size)
		common.LogError(err, "extend volume falied", true)
	},
}
var volumeRetype = &cobra.Command{
	Use:   "retype <volume> <new type>",
	Short: "Retype volume",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		migrationPolicy, _ := cmd.Flags().GetString("migration-policy")
		if err := storage.InvalidMIgrationPoicy(migrationPolicy); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		idOrName := args[0]
		newType := args[1]
		migrationPolicy, _ := cmd.Flags().GetString("migration-policy")

		client := cli.GetClient()
		volume, err := client.Storage.VolumeFound(idOrName)
		common.LogError(err, "get volume falied", true)

		err = client.Storage.VolumeRetype(volume.Id, newType, migrationPolicy)
		common.LogError(err, "extend volume falied", true)
	},
}

func init() {
	volumeList.Flags().BoolP("long", "l", false, "List additional fields in output")
	volumeList.Flags().Bool("all", false, "List volumes of all tenants")
	volumeList.Flags().StringP("name", "n", "", "Search by volume name")
	volumeList.Flags().String("status", "", "Search by volume status")

	volumePrune.Flags().StringP("name", "n", "", "Search by volume name")
	volumePrune.Flags().StringArrayP("status", "s", nil, "Search by server status")
	volumePrune.Flags().BoolP("yes", "y", false, i18n.T("answerYes"))

	volumeCreate.Flags().Uint("size", 0, "Volume size (GB)")
	volumeCreate.MarkFlagRequired("size")

	volumeRetype.Flags().StringP("migration-policy", "p", "never",
		fmt.Sprintf("Migration policy during retype of volume,\ninvalid values: %s",
			storage.MIGRATION_POLICYS))

	Volume.AddCommand(
		volumeList, volumeShow, volumeCreate, volumeExtend, volumeRetype,
		volumeDelete, volumePrune,
	)
}
