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
		client, err := cli.GetClient()
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}

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
		hypervisors.PrintTable(long, withServers)
	},
}

func init() {
	// flavor list flags
	hypervisorList.Flags().StringP("name", "n", "", "Show hypervisors matched by name")
	hypervisorList.Flags().BoolP("long", "l", false, "List additional fields in output")
	hypervisorList.Flags().Bool("with-servers", false, "List hypervisors with servers")
	// Server create flags

	Hypervisor.AddCommand(hypervisorList)
}
