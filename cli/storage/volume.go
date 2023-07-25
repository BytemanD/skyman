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

var VolumeList = &cobra.Command{
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
		dataTable.AddItems(volumes)
		dataTable.Print(long)
	},
}

var VolumeShow = &cobra.Command{
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

func init() {
	VolumeList.Flags().BoolP("long", "l", false, "List additional fields in output")
	VolumeList.Flags().StringP("name", "n", "", "Search by volume name")

	Volume.AddCommand(VolumeList, VolumeShow)
}
