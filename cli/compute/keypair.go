package compute

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/compute"
)

var Keypair = &cobra.Command{Use: "keypair"}

var keypairList = &cobra.Command{
	Use:   "list",
	Short: "List keypairs",
	Run: func(_ *cobra.Command, _ []string) {
		client := cli.GetClient()
		keypairs, err := client.Compute.KeypairList(nil)
		if err != nil {
			logging.Fatal("%s", err)
		}
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Name", Slot: func(item interface{}) interface{} {
					p, _ := item.(compute.Keypair)
					return p.Keypair.Name
				}},
				{Name: "Type", Slot: func(item interface{}) interface{} {
					p, _ := item.(compute.Keypair)
					return p.Keypair.Type
				}},
				{Name: "Fingerprint", Slot: func(item interface{}) interface{} {
					p, _ := item.(compute.Keypair)
					return p.Keypair.Fingerprint
				}},
			},
		}
		pt.AddItems(keypairs)
		common.PrintPrettyTable(pt, false)
	},
}

func init() {
	Keypair.AddCommand(keypairList)
}
