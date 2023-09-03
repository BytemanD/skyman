package compute

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/compute"
)

var Compute = &cobra.Command{Use: "compute"}
var computeService = &cobra.Command{Use: "service"}

func printServiceTable(item interface{}) {
	pt := common.PrettyItemTable{
		Item: item,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Binary"}, {Name: "Host"},
			{Name: "Status", Slot: func(item interface{}) interface{} {
				p, _ := (item).(compute.Service)
				return common.BaseColorFormatter.Format(p.Status)
			}},
			{Name: "State", Slot: func(item interface{}) interface{} {
				p, _ := item.(compute.Service)
				return common.BaseColorFormatter.Format(p.State)
			}},
			{Name: "ForcedDown", Text: "Forced Down"},
			{Name: "DisabledReason", Text: "Disabled Reason"},
		},
	}
	common.PrintPrettyItemTable(pt)
}

var csList = &cobra.Command{
	Use:   "list",
	Short: "List compute services",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		query := url.Values{}
		binary, _ := cmd.Flags().GetString("binary")
		host, _ := cmd.Flags().GetString("host")
		long, _ := cmd.Flags().GetBool("long")

		if binary != "" {
			query.Set("binary", binary)
		}
		if host != "" {
			query.Set("host", host)
		}
		services := client.Compute.ServiceList(query)
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
		client := cli.GetClient()
		service, err := client.Compute.ServiceEnable(args[0], args[1])
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
		client := cli.GetClient()
		reason, _ := cmd.Flags().GetString("reason")
		service, err := client.Compute.ServiceDisable(args[0], args[1], reason)
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
		client := cli.GetClient()
		service, err := client.Compute.ServiceUp(args[0], args[1])
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
		client := cli.GetClient()
		service, err := client.Compute.ServiceDown(args[0], args[1])
		if err != nil {
			fmt.Printf("Set service diabled failed: %v", err)
			os.Exit(1)
		}
		printServiceTable(*service)
	},
}

func init() {
	// compute service
	csList.Flags().String("binary", "", "Search by binary")
	csList.Flags().String("host", "", "Search by hostname")
	csList.Flags().StringArrayP("state", "s", nil, "Search by server status")
	csList.Flags().BoolP("long", "l", false, "List additional fields in output")
	csDisable.Flags().String("reason", "", "Disable reason")

	computeService.AddCommand(csList, csEnable, csDisable, csUp, csDown)

	Compute.AddCommand(computeService)
}
