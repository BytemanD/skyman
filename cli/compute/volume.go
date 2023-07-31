package compute

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/openstack/compute"
)

var Volume = &cobra.Command{Use: "volume"}

func printVolumeAttachments(items []compute.VolumeAttachment) {
	dataTable := cli.DataListTable{
		ShortHeaders: []string{"Id", "VolumeId", "Device"},
		HeaderLabel:  map[string]string{"Id": "Attachment Id"},
		SortBy: []table.SortBy{
			{Name: "Device", Mode: table.Asc},
		},
	}
	dataTable.AddItems(items)
	dataTable.Print(false)
}

var volumeList = &cobra.Command{
	Use:   "list <server>",
	Short: "List service volumes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		attachments, err := client.Compute.ServerVolumeList(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		printVolumeAttachments(attachments)
	},
}

var volumeAttach = &cobra.Command{
	Use:   "attach <server> <volume>",
	Short: "Attach volome to service",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		attachment, err := client.Compute.ServerVolumeAdd(args[0], args[1])
		if err != nil {
			fmt.Printf("Attach volume %s to server failed: %v", args[1], err)
			os.Exit(1)
		}
		printVolumeAttachments([]compute.VolumeAttachment{*attachment})
	},
}
var volumeDetach = &cobra.Command{
	Use:   "detach <host> <binary>",
	Short: "Detach volume from service",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		err := client.Compute.ServerVolumeDelete(args[0], args[1])
		if err != nil {
			fmt.Printf("Detach volume %s from server failed: %v", args[1], err)
			os.Exit(1)
		}
	},
}

func init() {
	// compute service
	Volume.AddCommand(volumeList, volumeAttach, volumeDetach)

	Server.AddCommand(Volume)
}
