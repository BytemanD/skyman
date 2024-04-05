package compute

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

var Volume = &cobra.Command{Use: "volume", Short: "Update server volume"}

func printVolumeAttachments(items []nova.VolumeAttachment) {
	pt := common.PrettyTable{
		ShortColumns: []common.Column{
			{Name: "Id", Text: "Attachment Id"},
			{Name: "VolumeId"}, {Name: "Device", Sort: true},
		},
	}
	pt.AddItems(items)
	common.PrintPrettyTable(pt, false)
}

var volumeList = &cobra.Command{
	Use:   "list <server>",
	Short: "List service volumes",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		attachments, err := client.NovaV2().Servers().ListVolumes(args[0])
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
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		attachment, err := client.NovaV2().Servers().AddVolume(args[0], args[1])
		if err != nil {
			fmt.Printf("Attach volume %s to server failed: %v", args[1], err)
			os.Exit(1)
		}
		printVolumeAttachments([]nova.VolumeAttachment{*attachment})
	},
}
var volumeDetach = &cobra.Command{
	Use:   "detach <host> <binary>",
	Short: "Detach volume from service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		err := client.NovaV2().Servers().DeleteVolume(args[0], args[1])
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
