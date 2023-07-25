package compute

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/openstack/compute"
)

var serverInterface = &cobra.Command{Use: "interface"}

func printinterfaceAttachments(items []compute.InterfaceAttachment) {
	dataTable := cli.DataListTable{
		ShortHeaders: []string{"PortState", "PortId", "NetId", "FixedIps", "MacAddr"},
		HeaderLabel: map[string]string{
			"PortState": "Port State",
			"PortId":    "Port Id",
			"NetId":     "Net Id",
			"FixedIps":  "IP Addresses",
			"MacAddr":   "Mac Addr",
		},
		Slots: map[string]func(item interface{}) interface{}{
			"FixedIps": func(item interface{}) interface{} {
				attachment, _ := item.(compute.InterfaceAttachment)
				return strings.Join(attachment.GetIPAddresses(), ", ")
			},
		},
	}
	dataTable.AddItems(items)
	dataTable.Print(false)
}

var interfaceList = &cobra.Command{
	Use:   "list <server>",
	Short: "List server interfaces",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		attachments, err := client.Compute.ServerInterfaceList(args[0])
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
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		attachment, err := client.Compute.ServerAddPort(args[0], args[1])
		if err != nil {
			fmt.Printf("Attach port %s to server failed: %v", args[1], err)
			os.Exit(1)
		}
		printinterfaceAttachments([]compute.InterfaceAttachment{*attachment})
	},
}
var interfaceAttachNet = &cobra.Command{
	Use:   "attach-net <server> <network id>",
	Short: "Attach network to server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		attachment, err := client.Compute.ServerAddNet(args[0], args[1])
		if err != nil {
			fmt.Printf("Attach network %s to server failed: %v", args[1], err)
			os.Exit(1)
		}
		printinterfaceAttachments([]compute.InterfaceAttachment{*attachment})
	},
}
var interfaceDetach = &cobra.Command{
	Use:   "detach <server> <port id>",
	Short: "Detach port from server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		err := client.Compute.ServerInterfaceDetach(args[0], args[1])
		if err != nil {
			fmt.Printf("Detach port %s from server failed: %v", args[1], err)
			os.Exit(1)
		}
	},
}

func init() {
	// compute service
	serverInterface.AddCommand(
		interfaceList, interfaceAttachNet, interfaceAttachPort,
		interfaceDetach)

	Server.AddCommand(serverInterface)
}
