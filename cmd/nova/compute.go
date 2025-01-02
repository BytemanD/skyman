package nova

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/cmd/views"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var (
	csListFlags    flags.ComputeServiceListFlags
	csDisableFlags flags.ComputeServiceDisableFlags
)
var Compute = &cobra.Command{Use: "compute"}
var computeService = &cobra.Command{Use: "service"}

var csList = &cobra.Command{
	Use:   "list",
	Short: "List compute services",
	Run: func(cmd *cobra.Command, _ []string) {
		client := openstack.DefaultClient()

		query := url.Values{}

		if *csListFlags.Binary != "" {
			query.Set("binary", *csListFlags.Binary)
		}
		if *csListFlags.Host != "" {
			query.Set("host", *csListFlags.Host)
		}

		services, _ := client.NovaV2().Service().List(query)
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
		if *csListFlags.Zone != "" {
			pt.ShortColumns[3].Filters = []string{*csListFlags.Zone}
		}
		pt.AddItems(services)
		common.PrintPrettyTable(pt, *csListFlags.Long)
	},
}

var csEnable = &cobra.Command{
	Use:   "enable <host> <binary>",
	Short: "Enable compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		service, err := client.NovaV2().Service().Enable(args[0], args[1])
		utility.LogError(err, "set service disable failed", true)
		views.PrintServiceTable(*service)
	},
}
var csDisable = &cobra.Command{
	Use:   "disable <host> <binary>",
	Short: "Disable compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		service, err := client.NovaV2().Service().Disable(args[0], args[1], *csDisableFlags.Reason)
		if err != nil {
			utility.LogError(err, "set service disable failed", true)
		}
		views.PrintServiceTable(*service)
	},
}
var csUp = &cobra.Command{
	Use:   "up <host> <binary>",
	Short: "Unset force down compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		service, err := client.NovaV2().Service().Up(args[0], args[1])
		if err != nil {
			utility.LogError(err, "set service up failed", true)
		}
		views.PrintServiceTable(*service)
	},
}
var csDown = &cobra.Command{
	Use:   "down <host> <binary>",
	Short: "Force down compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		service, err := client.NovaV2().Service().Down(args[0], args[1])
		if err != nil {
			utility.LogError(err, "set service down failed", true)
		}
		views.PrintServiceTable(*service)
	},
}
var csDelete = &cobra.Command{
	Use:   "delete <host> <binary>",
	Short: "Delete compute service",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		err := client.NovaV2().Service().Delete(args[0], args[1])
		utility.LogError(err, "delete service failed", true)
	},
}

func init() {
	// compute service
	csListFlags = flags.ComputeServiceListFlags{
		Binary: csList.Flags().String("binary", "", "Search by binary"),
		Host:   csList.Flags().String("host", "", "Search by hostname"),
		Zone:   csList.Flags().String("zone", "", "Search by zone"),
		State:  csList.Flags().StringArrayP("state", "s", nil, "Search by server status"),
		Long:   csList.Flags().BoolP("long", "l", false, "List additional fields in output"),
	}
	csDisableFlags = flags.ComputeServiceDisableFlags{
		Reason: csDisable.Flags().String("reason", "", "Reason"),
	}

	computeService.AddCommand(csList, csEnable, csDisable, csUp, csDown, csDelete)

	Compute.AddCommand(computeService)
}
