package compute

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var Compute = &cobra.Command{Use: "compute"}
var computeService = &cobra.Command{Use: "service"}

var csList = &cobra.Command{
	Use:   "list",
	Short: "List compute services",
	Run: func(cmd *cobra.Command, _ []string) {
		client := openstack.DefaultClient()

		query := url.Values{}
		binary, _ := cmd.Flags().GetString("binary")
		host, _ := cmd.Flags().GetString("host")
		zone, _ := cmd.Flags().GetString("zone")
		long, _ := cmd.Flags().GetBool("long")

		if binary != "" {
			query.Set("binary", binary)
		}
		if host != "" {
			query.Set("host", host)
		}

		services, _ := client.NovaV2().Services().List(query)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Binary"},
				{Name: "Host"}, {Name: "Zone"},
				{Name: "Status", AutoColor: true},
				{Name: "State", AutoColor: true},
				{Name: "UpdatedAt"},
			},
			LongColumns: []common.Column{
				{Name: "DisabledReason"}, {Name: "ForcedDown"},
			},
			Filters: map[string]string{},
		}
		if zone != "" {
			pt.ShortColumns[3].Filters = []string{zone}
		}
		pt.AddItems(services)
		common.PrintPrettyTable(pt, long)
	},
}

var csEnable = &cobra.Command{
	Use:   "enable <host> <binary>",
	Short: "Enable compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		service, err := client.NovaV2().Services().Enable(args[0], args[1])
		if err != nil {
			fmt.Printf("Set service diabled failed: %v", err)
			os.Exit(1)
		}
		printServiceTable(*service)
	},
}
var csDisable = &cobra.Command{
	Use:   "disable <host> <binary>",
	Short: "Disable compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		reason, _ := cmd.Flags().GetString("reason")
		service, err := client.NovaV2().Services().Disable(args[0], args[1], reason)
		if err != nil {
			fmt.Printf("Set service diabled failed: %v", err)
			os.Exit(1)
		}
		printServiceTable(*service)
	},
}
var csUp = &cobra.Command{
	Use:   "up <host> <binary>",
	Short: "Unset force down compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		service, err := client.NovaV2().Services().Up(args[0], args[1])
		if err != nil {
			fmt.Printf("Set service diabled failed: %v", err)
			os.Exit(1)
		}
		printServiceTable(*service)
	},
}
var csDown = &cobra.Command{
	Use:   "down <host> <binary>",
	Short: "Force down compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		service, err := client.NovaV2().Services().Down(args[0], args[1])
		if err != nil {
			fmt.Printf("Set service diabled failed: %v", err)
			os.Exit(1)
		}
		printServiceTable(*service)
	},
}
var csDelete = &cobra.Command{
	Use:   "delete <host> <binary>",
	Short: "Delete compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		err := client.NovaV2().Services().Delete(args[0], args[1])
		utility.LogError(err, "Delete service failed, %v", true)
	},
}

func init() {
	// compute service
	csList.Flags().String("binary", "", "Search by binary")
	csList.Flags().String("host", "", "Search by hostname")
	csList.Flags().String("zone", "", "Search by zone")
	csList.Flags().StringArrayP("state", "s", nil, "Search by server status")
	csList.Flags().BoolP("long", "l", false, "List additional fields in output")

	csDisable.Flags().String("reason", "", "Disable reason")

	computeService.AddCommand(csList, csEnable, csDisable, csUp, csDown, csDelete)

	Compute.AddCommand(computeService)
}
