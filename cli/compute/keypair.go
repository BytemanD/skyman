package compute

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
)

var Keypair = &cobra.Command{Use: "keypair"}

var keypairList = &cobra.Command{
	Use:   "list",
	Short: "List keypairs",
	Run: func(cmd *cobra.Command, _ []string) {
		client, err := cli.GetClient()
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}
		keypairs, err := client.Compute.KeypairList(nil)
		if err != nil {
			logging.Fatal("%s", err)
		}
		dataTable := cli.DataListTable{
			ShortHeaders: []string{
				"Name", "Type", "Fingerprint",
			},
		}
		for _, keypair := range keypairs {
			dataTable.Items = append(dataTable.Items, keypair.Keypair)
		}
		dataTable.Print(false)
	},
}

func init() {
	Keypair.AddCommand(keypairList)
}
