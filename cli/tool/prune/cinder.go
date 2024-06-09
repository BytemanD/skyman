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

		query := url.Values{}
		if status != "" {
			query.Add("status", status)
		}
		client := openstack.DefaultClient()
		client.CinderV2().Volumes().Prune(query, name, yes)
	},
}

func init() {
	volumePrune.Flags().StringP("name", "n", "", "Filter by volume name")
	volumePrune.Flags().StringP("status", "s", "error", "Search by server status")
	volumePrune.Flags().BoolP("yes", "y", false, i18n.T("answerYes"))
}
