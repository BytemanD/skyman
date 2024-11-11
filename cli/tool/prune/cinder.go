package prune

import (
	"net/url"

	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/spf13/cobra"
)

var volumePrune = &cobra.Command{
	Use:   "volume",
	Short: "Prune volume(s)",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		yes, _ := cmd.Flags().GetBool("yes")
		status, _ := cmd.Flags().GetString("status")
		volumeType, _ := cmd.Flags().GetString("type")

		query := url.Values{}
		query.Set("sort", "created_at:desc")
		if status != "" {
			query.Add("status", status)
		}
		c := openstack.DefaultClient()
		c.PruneVolumes(query, name, volumeType, yes)
	},
}

func init() {
	volumePrune.Flags().StringP("name", "n", "", "Filter by volume name")
	volumePrune.Flags().StringP("status", "s", "error", "Search by volume status")
	volumePrune.Flags().StringP("type", "t", "", "Search by volume type")
	volumePrune.Flags().BoolP("yes", "y", false, i18n.T("answerYes"))
}
