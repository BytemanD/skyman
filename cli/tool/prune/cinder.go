package prune

import (
	"net/url"
	"strconv"

	"github.com/BytemanD/skyman/cli/flags"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/spf13/cobra"
)

var volumePruneFlags flags.PruneVolumeFlags

var volumePrune = &cobra.Command{
	Use:   "volume",
	Short: "Prune volume(s)",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		query := url.Values{}
		// query.Set("sort", "created_at:desc")
		if *volumePruneFlags.Status != "" {
			query.Add("status", *volumePruneFlags.Status)
		}
		if *volumePruneFlags.All {
			query.Add("all_tenants", "1")
		}
		if *volumePruneFlags.Limit > 0 {
			query.Add("limit", strconv.Itoa(int(*volumePruneFlags.Limit)))
		}
		if *volumePruneFlags.Marker != "" {
			query.Add("marker", *volumePruneFlags.Marker)
		}
		c := openstack.DefaultClient()
		c.PruneVolumes(query, *volumePruneFlags.Name, *volumePruneFlags.Type, *volumePruneFlags.Yes)
	},
}

func init() {
	volumePruneFlags = flags.PruneVolumeFlags{
		Name:   volumePrune.Flags().StringP("name", "n", "", "Filter by volume name"),
		Status: volumePrune.Flags().StringP("status", "s", "error", "Search by volume status, e.g. available, error"),
		Type:   volumePrune.Flags().StringP("type", "t", "", "Search by volume type"),
		All:    volumePrune.Flags().Bool("all", false, "Search by all tenants"),
		Yes:    volumePrune.Flags().BoolP("yes", "y", false, i18n.T("answerYes")),
		Marker: volumePrune.Flags().String("marker", "", "Marker"),
		Limit:  volumePrune.Flags().Uint("limit", 1000, "Limit"),
	}
}
