package compute

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
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
			query.Set("with_servers", "true")
		}
		if name != "" {
			query.Set("hypervisor_hostname_pattern", name)
		}
		hypervisors, err := client.Compute.HypervisorListDetail(query)
		if err != nil {
			logging.Fatal("%s", err)
		}
		dataTable := cli.DataListTable{
			ShortHeaders: []string{
				"Id", "Hostname", "HostIp", "Status", "State"},
			LongHeaders: []string{
				"Type", "Version", "Vcpus", "VcpusUsed",
				"MemoryMB", "MemoryMBUsed"},
			HeaderLabel: map[string]string{
				"HostIp":       "Host IP",
				"VcpusUsed":    "VCPUs Used",
				"MemoryMBUsed": "MemoryMB Used"},
		}
		for _, hypervisor := range hypervisors {
			dataTable.Items = append(dataTable.Items, hypervisor)
		}
		dataTable.Print(long)
	},
}

func init() {
	// hypervisor list flags
	hypervisorList.Flags().StringP("name", "n", "", "Show hypervisors matched by name")
	hypervisorList.Flags().BoolP("long", "l", false, "List additional fields in output")
	hypervisorList.Flags().Bool("with-servers", false, "List hypervisors with servers")

	Hypervisor.AddCommand(hypervisorList)
}
