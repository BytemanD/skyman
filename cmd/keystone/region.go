package keystone

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var Region = &cobra.Command{Use: "region"}

var list = &cobra.Command{
	Use:   "list",
	Short: "List regions",
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient().KeystoneV3()
		regions, err := c.Region().List(nil)
		utility.LogError(err, "list region failed", true)
		common.PrintRegions(regions, false)
	},
}

func init() {
	Region.AddCommand(list)
}
