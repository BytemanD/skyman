package storage

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/openstack/storage"
	"github.com/spf13/cobra"
)

var Volume = &cobra.Command{Use: "volume"}

var volumeList = &cobra.Command{
	Use:   "list",
	Short: "List volumes",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		volumes := client.Storage.VolumeListDetail(query)
		dataTable := cli.DataListTable{
			ShortHeaders: []string{
				"Id", "Name", "Status", "Size", "Bootable", "Attachments"},
			LongHeaders: []string{"VolumeType", "Metadata"},
			HeaderLabel: map[string]string{
				"VolumeType": "Volume Type",
			},
			Slots: map[string]func(item interface{}) interface{}{
				"Attachments": func(item interface{}) interface{} {
					obj, _ := (item).(storage.Volume)
					return strings.Join(obj.GetAttachmentList(), "\n")
				},
				"Metadata": func(item interface{}) interface{} {
					obj, _ := (item).(storage.Volume)
					return strings.Join(obj.GetMetadataList(), "\n")
				},
			},
		}
		if long {
			dataTable.StyleSeparateRows = true
		}
		dataTable.AddItems(volumes)
		dataTable.Print(long)
	},
}

var volumeShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		idOrName := args[0]
		volume, err := client.Storage.VolumeShow(idOrName)
		if err != nil {
			logging.Fatal("%s", err)
		}
		if err != nil {
			volumes := client.Storage.VolumeListDetailByName(idOrName)
			if len(volumes) > 1 {
				fmt.Printf("Found multi volumes named %s\n", idOrName)
				os.Exit(1)
			} else if len(volumes) == 1 {
				volume = &volumes[0]
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		printVolume(*volume)
	},
}
var volumeDelete = &cobra.Command{
	// TODO: support volume id or name
	Use:   "delete <volume id> [<volume id> ...]",
	Short: "Delete volume",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, volumeId := range args {
			err := client.Storage.VolumeDelete(volumeId)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("Requested to delete volume %s\n", volumeId)
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

func init() {
	volumeList.Flags().BoolP("long", "l", false, "List additional fields in output")
	volumeList.Flags().StringP("name", "n", "", "Search by volume name")

	volumePrune.Flags().StringP("name", "n", "", "Search by volume name")
	volumePrune.Flags().StringArrayP("status", "s", nil, "Search by server status")
	volumePrune.Flags().BoolP("yes", "y", false, "所有问题自动回答'是'")

	volumeCreate.Flags().Uint("size", 0, "Volume size")

	Volume.AddCommand(
		volumeList, volumeShow, volumeCreate,
		volumeDelete, volumePrune,
	)
}
