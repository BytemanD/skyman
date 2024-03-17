package prune

import (
	"net/url"

	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/spf13/cobra"
)

var VolumePrune = &cobra.Command{
	Use:   "volume",
	Short: "Prune volume(s)",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		yes, _ := cmd.Flags().GetBool("yes")
		status, _ := cmd.Flags().GetString("status")

		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		if status != "" {
			query.Add("status", status)
		}
		client := openstack.DefaultClient()
		client.CinderV2().Volumes().Prune(query, yes)

	},
}

func init() {
	VolumePrune.Flags().StringP("name", "n", "", "Search by volume name")
	VolumePrune.Flags().StringP("status", "s", "error", "Search by server status")
	VolumePrune.Flags().BoolP("yes", "y", false, i18n.T("answerYes"))
}
