package nova

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
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
		client := common.DefaultClient()
		server, err := client.NovaV2().FindServer(args[0])
		utility.LogIfError(err, true, "get server %s faield", args[0])
		attachments, err := client.NovaV2().ListServerVolumes(server.Id)
		if err != nil {
			println(err)
		}
		printVolumeAttachments(attachments)
	},
}

var volumeAttach = &cobra.Command{
	Use:   "attach <server> <volume-id>",
	Short: "Attach volome to service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()

		server, err := client.NovaV2().FindServer(args[0])
		utility.LogIfError(err, true, "get server %s faield", args[0])

		volume, err := client.CinderV2().FindVolume(args[1])
		utility.LogIfError(err, true, "get volume %s faield", args[1])

		attachment, err := client.NovaV2().ServerAddVolume(server.Id, volume.Id)
		utility.LogError(err, "Attach volume to server failed", true)
		printVolumeAttachments([]nova.VolumeAttachment{*attachment})
	},
}
var volumeDetach = &cobra.Command{
	Use:   "detach <server> <volume id>",
	Short: "Detach volume from service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()
		server, err := client.NovaV2().FindServer(args[0])
		utility.LogIfError(err, true, "get server %s faield", args[0])

		volume, err := client.CinderV2().FindVolume(args[1])
		utility.LogIfError(err, true, "get volume %s faield", args[1])

		err = client.NovaV2().ServerDeleteVolume(server.Id, volume.Id)
		utility.LogError(err, "Detach volume from server failed", true)
	},
}

func init() {
	// compute service
	Volume.AddCommand(volumeList, volumeAttach, volumeDetach)

	Server.AddCommand(Volume)
}
