package compute

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
)

var Keypair = &cobra.Command{Use: "keypair"}

var keypairList = &cobra.Command{
	Use:   "list",
	Short: "List keypairs",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()
		keypairs, err := client.Compute.KeypairList(nil)
		if err != nil {
			logging.Fatal("%s", err)
		}
		dataListTable := common.DataListTable{
			ShortHeaders: []string{"Name", "Type", "Fingerprint"},
		}
		dataListTable.AddItems(keypairs)
		common.PrintDataListTable(dataListTable, false)
	},
}

func init() {
	Keypair.AddCommand(keypairList)
}
