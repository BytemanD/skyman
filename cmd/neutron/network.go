package neutron

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Network = &cobra.Command{Use: "network"}

var networkList = &cobra.Command{
	Use:   "list",
	Short: "List networks",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := common.DefaultClient().NeutronV2()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		networks, err := c.ListNetwork(query)
		utility.LogIfError(err, true, "list network failed")
		common.PrintNetworks(networks, long)
	},
}
var networkDelete = &cobra.Command{
	Use:   "delete <network> [network ...]",
	Short: "Delete network(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()

		for _, idOrName := range args {
			fmt.Printf("Reqeust to delete network %s\n", idOrName)
			network, err := c.FindNetwork(idOrName)
			utility.LogIfError(err, true, "get network %s failed", idOrName)
			err = c.DeleteNetwork(network.Id)
			utility.LogIfError(err, false, "delete network %s failed", idOrName)
		}
	},
}
var networkShow = &cobra.Command{
	Use:   "show <network>",
	Short: "Show network",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()

		network, err := c.FindNetwork(args[0])
		utility.LogError(err, "show network failed", true)
		common.PrintNetwork(*network)
	},
}
var networkCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create network",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()

		// name, _ := cmd.Flags().GetString("name")
		disable, _ := cmd.Flags().GetBool("disable")
		description, _ := cmd.Flags().GetString("description")
		params := map[string]any{
			"name": args[0],
		}
		if disable {
			params["enable"] = false
		}
		if description != "" {
			params["description"] = description
		}
		network, err := c.CreateNetwork(params)
		utility.LogError(err, "create network failed", true)
		common.PrintNetwork(*network)
	},
}

func init() {
	networkList.Flags().BoolP("long", "l", false, "List additional fields in output")
	networkList.Flags().StringP("name", "n", "", "Search by router name")

	networkCreate.Flags().String("description", "", "Set network description")
	networkCreate.Flags().Bool("disable", false, "Disable router")

	Network.AddCommand(networkList, networkShow, networkDelete, networkCreate)
	// Network.AddCommand(agentCmd)
}
