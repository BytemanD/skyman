package quota

import (
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var QuotaCmd = &cobra.Command{Use: "quota"}

var show = &cobra.Command{
	Use:   "show",
	Short: "Show quotas for project or class.",
	Args:  cobra.ExactArgs(0),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		projectId, err := client.ProjectId()
		utility.LogError(err, "get project id failed, %v", true)

		quotaSet, err := client.NovaV2().Quota().Show(projectId)
		utility.LogError(err, "show quota failed, %v", true)
		var showFipAndFixedIps bool
		if !client.NovaV2().MicroVersionLargeEqual("2.36") {
			showFipAndFixedIps = true
		}
		common.PrintQuotaSet(*quotaSet, showFipAndFixedIps)
	},
}

func init() {
	QuotaCmd.AddCommand(show)
}
