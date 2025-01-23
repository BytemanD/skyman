package neutron

import (
	"fmt"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/duke-git/lancet/v2/slice"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Vpc = &cobra.Command{Use: "vpc"}

var vpcCreate = &cobra.Command{
	Use:   "create <name> <cidr>",
	Short: "Create VPC",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}

		ipVersion, _ := cmd.Flags().GetString("ip-version")
		ipVersions := strings.Split(ipVersion, ",")
		for _, v := range ipVersions {
			switch v {
			case "4":
				if !common.ValidIpv4(args[1], 30) {
					return fmt.Errorf("invalid IPv4 address: %s", args[1])
				}
			case "6":
				continue
			default:
				return fmt.Errorf("invalid ip vresion: %s", v)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		vpc, cidr := args[0], args[1]

		c := openstack.DefaultClient().NeutronV2()
		routerName := fmt.Sprintf("%s-router", vpc)
		networkName := fmt.Sprintf("%s-network", vpc)
		subnetName := fmt.Sprintf("%s-subnet", vpc)
		ipVersion, _ := cmd.Flags().GetString("ip-version")
		ipVersions := strings.Split(ipVersion, ",")

		// create router
		routerParams := map[string]interface{}{"name": routerName}
		console.Info("create router %s", routerName)
		router, err := c.Router().Create(routerParams)
		utility.LogIfError(err, true, "create router %s failed", routerName)
		// create network
		networkParams := map[string]interface{}{"name": networkName}
		console.Info("create network %s", networkName)
		network, err := c.Network().Create(networkParams)
		utility.LogIfError(err, true, "create network %s failed", networkParams)
		// create router
		for _, v := range ipVersions {
			subneVerionName := fmt.Sprintf("%s-v%s", subnetName, v)
			subnetParams := map[string]interface{}{
				"name":       subneVerionName,
				"network_id": network.Id,
				"ip_version": v,
				"cidr":       cidr,
			}
			console.Info("create subnet %s", subneVerionName)
			subnet, err := c.Subnet().Create(subnetParams)
			utility.LogIfError(err, true, "create subnet %s failed", subnetName)
			// add router interface
			console.Info("add subnet %s to router %s", subneVerionName, routerName)
			err = c.Router().AddSubnet(router.Id, subnet.Id)
			utility.LogIfError(err, true, "add subnet %s to router %s failed", subnetName, routerName)
			console.Info("create VPC %s success", vpc)
		}
	},
}
var vpcDelete = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete VPC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpc := args[0]
		vpcRouter, _ := cmd.Flags().GetString("router")

		if vpcRouter == "" {
			vpcRouter = fmt.Sprintf("%s-router", vpc)
		}

		c := openstack.DefaultClient().NeutronV2()
		// get vpc router
		console.Info("get router %s", vpcRouter)
		router, err := c.Router().Find(vpcRouter)
		utility.LogIfError(err, true, "get router %s failed", vpcRouter)
		// remove router ports
		routerPorts, err := c.Port().ListByDeviceId(router.Id)
		utility.LogIfError(err, true, "list router ports failed")
		subnets := []string{}
		for _, port := range routerPorts {
			for _, fixedIp := range port.FixedIps {
				console.Info("remove subnet %s from router %s", fixedIp.SubnetId, router.Id)
				c.Router().RemoveSubnet(router.Id, fixedIp.SubnetId)
				if !slice.Contain(subnets, fixedIp.SubnetId) {
					subnets = append(subnets, fixedIp.SubnetId)
				}
			}
		}
		// delete vpc networks
		for _, subnetId := range subnets {
			subnet, err := c.Subnet().Show(subnetId)
			utility.LogIfError(err, true, "get subnet %s failed", subnetId)
			console.Info("delete vpc network %s", subnet.NetworkId)
			err = c.Network().Delete(subnet.NetworkId)
			utility.LogIfError(err, true, "delete network %s failed", subnet.NetworkId)
		}

		// delete vpc router
		console.Info("delete vpc router %s", vpcRouter)
		err = c.Router().Delete(router.Id)
		utility.LogIfError(err, true, "dele router %s failed", vpcRouter)
		console.Info("VPC %s delete success", vpc)
	},
}

func init() {
	vpcCreate.Flags().StringP("ip-version", "v", "4", "IP version")

	vpcDelete.Flags().StringP("router", "r", "", "Router id or name")

	Vpc.AddCommand(vpcCreate, vpcDelete)
}
