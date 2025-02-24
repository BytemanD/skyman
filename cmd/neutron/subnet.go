package neutron

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Subnet = &cobra.Command{Use: "subnet"}

var subnetList = &cobra.Command{
	Use:   "list",
	Short: "List subnets",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := common.DefaultClient().NeutronV2()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		subnets, err := c.Subnet().List(query)
		utility.LogError(err, "get subnets failed", true)
		common.PrintSubnets(subnets, long)
	},
}
var subnetCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create subnet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()

		noDhcp, _ := cmd.Flags().GetBool("no-dhcp")
		description, _ := cmd.Flags().GetString("description")
		netIdOrName, _ := cmd.Flags().GetString("network")
		cidr, _ := cmd.Flags().GetString("cidr")
		ipVersion, _ := cmd.Flags().GetInt("ip-version")

		network, err := c.Network().Find(netIdOrName)
		utility.LogError(err, "get network failed", true)

		subnet, err := c.Subnet().Create(map[string]interface{}{
			"name":        args[0],
			"cidr":        cidr,
			"network_id":  network.Id,
			"ip_version":  ipVersion,
			"enable_dhcp": !noDhcp,
			"description": description,
		})
		utility.LogError(err, "create subnet failed", true)
		common.PrintSubnet(*subnet)
	},
}
var subnetShow = &cobra.Command{
	Use:   "show <subnet>",
	Short: "Show subnet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()
		subnet, err := c.Subnet().Find(args[0])
		utility.LogError(err, "show subnet failed", true)
		common.PrintSubnet(*subnet)
	},
}
var subnetDelete = &cobra.Command{
	Use:   "delete <subnet> [subnet ...]",
	Short: "Delete subnet(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()
		for _, idOrName := range args {
			fmt.Printf("Reqeust to delete subnet %s\n", idOrName)
			subnet, err := c.Subnet().Find(idOrName)
			utility.LogIfError(err, true, "get subnet %s failed", idOrName)
			err = c.Subnet().Delete(subnet.Id)
			utility.LogIfError(err, false, "Delete subnet %s failed", idOrName)
		}
	},
}

func init() {
	subnetList.Flags().BoolP("long", "l", false, "List additional fields in output")
	subnetList.Flags().StringP("name", "n", "", "Search by router name")

	subnetCreate.Flags().String("description", "", "Set subnet description")
	subnetCreate.Flags().String("network", "", "Set subnet description")
	subnetCreate.Flags().String("cidr", "", "Subnet range in CIDR notation")
	subnetCreate.Flags().Bool("no-dhcp", false, "Disable DHCP")
	subnetCreate.Flags().Int("ip-version", 4, "IP version (default is 4).")

	subnetCreate.MarkFlagRequired("network")
	subnetCreate.MarkFlagRequired("cidr")

	Subnet.AddCommand(subnetList, subnetCreate, subnetDelete, subnetShow)
}
