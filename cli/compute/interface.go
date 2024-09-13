package compute

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var serverInterface = &cobra.Command{Use: "interface"}

func printinterfaceAttachments(items []nova.InterfaceAttachment) {
	dataListTable := common.PrettyTable{
		ShortColumns: []common.Column{
			{Name: "PortState", AutoColor: true},
			{Name: "PortId"},
			{Name: "NetId"},
			{Name: "FixedIps", Text: "IP Addresses", Slot: func(item interface{}) interface{} {
				attachment, _ := item.(nova.InterfaceAttachment)
				return strings.Join(attachment.GetIPAddresses(), ", ")
			}},
			{Name: "MacAddr"},
		},
	}
	dataListTable.AddItems(items)
	common.PrintPrettyTable(dataListTable, false)
}

var interfaceList = &cobra.Command{
	Use:   "list <server>",
	Short: "List server interfaces",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		server, err := client.NovaV2().Server().Found(args[0])
		utility.LogIfError(err, true, "get server %s faield", args[0])
		attachments, err := client.NovaV2().Server().ListInterfaces(server.Id)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		printinterfaceAttachments(attachments)
	},
}

var interfaceAttachPort = &cobra.Command{
	Use:   "attach-port <server> <port id>",
	Short: "Attach port to server",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		server, err := client.NovaV2().Server().Found(args[0])
		utility.LogIfError(err, true, "get server %s failed", args[0])

		attachment, err := client.NovaV2().Server().AddInterface(server.Id, "", args[1])
		utility.LogError(err, fmt.Sprintf("Attach port %s to server failed", args[1]), true)
		printinterfaceAttachments([]nova.InterfaceAttachment{*attachment})
	},
}
var interfaceAttachNet = &cobra.Command{
	Use:   "attach-net <server> <network id>",
	Short: "Attach network to server",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		server, err := client.NovaV2().Server().Found(args[0])
		utility.LogIfError(err, true, "get server %s failed", args[0])

		network, err := client.NeutronV2().Network().Found(args[1])
		utility.LogIfError(err, true, "get network %s failed", args[1])

		attachment, err := client.NovaV2().Server().AddInterface(server.Id, network.Id, "")
		utility.LogError(err, fmt.Sprintf("Attach network %s to server failed", args[1]), true)
		printinterfaceAttachments([]nova.InterfaceAttachment{*attachment})
	},
}
var interfaceDetach = &cobra.Command{
	Use:   "detach <server> <port id>",
	Short: "Detach port from server",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		server, err := client.NovaV2().Server().Found(args[0])
		utility.LogIfError(err, true, "get server %s failed", args[0])
		_, err = client.NovaV2().Server().DeleteInterface(server.Id, args[1])
		utility.LogError(err, "Detach port from server failed", true)
	},
}

func init() {
	serverInterface.AddCommand(
		interfaceList, interfaceAttachNet, interfaceAttachPort,
		interfaceDetach,
	)

	Server.AddCommand(serverInterface)
}
