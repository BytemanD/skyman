package prune

import (
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/spf13/cobra"
)

var serverPrune = &cobra.Command{
	Use:   "server",
	Short: "Prune server(s)",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		yes, _ := cmd.Flags().GetBool("yes")
		name, _ := cmd.Flags().GetString("name")
		host, _ := cmd.Flags().GetString("host")
		statusList, _ := cmd.Flags().GetStringArray("status")

		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		if host != "" {
			query.Set("host", host)
		}
		for _, status := range statusList {
			query.Add("status", status)
		}
		c := common.DefaultClient()
		c.PruneServers(query, yes, true)
	},
}

func init() {
	serverPrune.Flags().StringP("name", "n", "", "Search by server name")
	serverPrune.Flags().String("host", "", "Search by hostname")
	serverPrune.Flags().StringArrayP("status", "s", nil, "Search by server status")
	serverPrune.Flags().BoolP("yes", "y", false, i18n.T("answerYes"))
}
