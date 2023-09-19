package compute

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
)

var Hypervisor = &cobra.Command{Use: "hypervisor"}

var hypervisorList = &cobra.Command{
	Use:   "list",
	Short: "List hypervisors",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		withServers, _ := cmd.Flags().GetBool("with-servers")

		query := url.Values{}
		if withServers {
			// TODO
			query.Set("with_servers", "true")
		}
		if name != "" {
			query.Set("hypervisor_hostname_pattern", name)
		}
		hypervisors, err := client.Compute.HypervisorListDetail(query)
		if err != nil {
			logging.Fatal("%s", err)
		}
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Hostname"}, {Name: "HostIp"},
				{Name: "Status", AutoColor: true}, {Name: "State", AutoColor: true},
			},
			LongColumns: []common.Column{
				{Name: "Type"}, {Name: "Version"},
				{Name: "Vcpus"}, {Name: "VcpusUsed"},
				{Name: "MemoryMB", Text: "Memory(MB)"},
				{Name: "MemoryMBUsed", Text: "Memory Used(MB)"},
			},
		}
		pt.AddItems(hypervisors)
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	// hypervisor list flags
	hypervisorList.Flags().StringP("name", "n", "", "Show hypervisors matched by name")
	hypervisorList.Flags().BoolP("long", "l", false, "List additional fields in output")
	hypervisorList.Flags().Bool("with-servers", false, "List hypervisors with servers")

	Hypervisor.AddCommand(hypervisorList)
}
