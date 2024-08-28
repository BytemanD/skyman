package storage

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/utility"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var Volume = &cobra.Command{Use: "volume"}

var volumeList = &cobra.Command{
	Use:   "list",
	Short: "List volumes",
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
		volumes, err := client.CinderV2().Volume().Detail(query)
		utility.LogError(err, "list volume falied", true)
		table := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Status", AutoColor: true},
				{Name: "Size", Align: text.AlignRight},
				{Name: "Bootable"}, {Name: "VolumeType"},
				{Name: "Attachments", Slot: func(item interface{}) interface{} {
					obj, _ := (item).(cinder.Volume)
					return strings.Join(obj.GetAttachmentList(), "\n")
				}},
			},
			LongColumns: []common.Column{
				{Name: "Metadata", Slot: func(item interface{}) interface{} {
					obj, _ := (item).(cinder.Volume)
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
		client := openstack.DefaultClient()
		idOrName := args[0]
		volume, err := client.CinderV2().Volume().Found(idOrName)
		utility.LogError(err, "get volume failed", true)
		printVolume(*volume)
	},
}
var volumeDelete = &cobra.Command{
	Use:   "delete <volume1> [<volume2> ...]",
	Short: "Delete volume",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		force, _ := cmd.Flags().GetBool("force")
		cascade, _ := cmd.Flags().GetBool("cascade")

		for _, idOrName := range args {
			volume, err := client.CinderV2().Volume().Found(idOrName)
			if err != nil {
				utility.LogError(err, "get volume failed", false)
				continue
			}
			err = client.CinderV2().Volume().Delete(volume.Id, force, cascade)
			if err == nil {
				fmt.Printf("Requested to delete volume %s\n", idOrName)
			} else {
				fmt.Println(err)
			}
		}
	},
}

var volumeCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		size, _ := cmd.Flags().GetUint("size")
		volumeType, _ := cmd.Flags().GetString("type")
		multiattach, _ := cmd.Flags().GetBool("multiattach")

		params := map[string]interface{}{"name": args[0]}
		if size > 0 {
			params["size"] = size
		}
		if volumeType != "" {
			params["volume_type"] = volumeType
		}
		if multiattach {
			params["multiattach"] = multiattach
		}

		client := openstack.DefaultClient()

		volume, err := client.CinderV2().Volume().Create(params)
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
		client := openstack.DefaultClient()
		volume, err := client.CinderV2().Volume().Found(idOrName)
		utility.LogError(err, "get volume falied", true)

		err = client.CinderV2().Volume().Extend(volume.Id, size)
		utility.LogError(err, "extend volume falied", true)
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
		if err := openstack.InvalidMIgrationPoicy(migrationPolicy); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		idOrName := args[0]
		newType := args[1]
		migrationPolicy, _ := cmd.Flags().GetString("migration-policy")

		client := openstack.DefaultClient()
		volume, err := client.CinderV2().Volume().Found(idOrName)
		utility.LogError(err, "get volume falied", true)

		err = client.CinderV2().Volume().Retype(volume.Id, newType, migrationPolicy)
		utility.LogError(err, "extend volume falied", true)
	},
}

func init() {
	volumeList.Flags().BoolP("long", "l", false, "List additional fields in output")
	volumeList.Flags().Bool("all", false, "List volumes of all tenants")
	volumeList.Flags().StringP("name", "n", "", "Search by volume name")
	volumeList.Flags().String("status", "", "Search by volume status")

	volumeCreate.Flags().Uint("size", 0, "Volume size (GB)")
	volumeCreate.Flags().String("type", "", "Volume type")
	volumeCreate.Flags().Bool("multiattach", false, "Allow multiattach")
	volumeCreate.MarkFlagRequired("size")

	volumeDelete.Flags().Bool("force", false, "Force delete volume.")
	volumeDelete.Flags().Bool("cascade", false, "Remove any snapshots along with volume")

	volumeRetype.Flags().StringP("migration-policy", "p", "never",
		fmt.Sprintf("Migration policy during retype of volume,\ninvalid values: %s",
			openstack.MIGRATION_POLICYS))

	Volume.AddCommand(
		volumeList, volumeShow, volumeCreate, volumeExtend, volumeRetype,
		volumeDelete,
	)
}
