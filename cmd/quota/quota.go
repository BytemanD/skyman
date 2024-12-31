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

		pt := common.PrettyItemTable{
			ShortFields: []common.Column{
				{Name: "Instances"},
				{Name: "Cores"}, {Name: "Ram"},
				{Name: "MetadataItems"},
				{Name: "SecurityGroups"},
				{Name: "SecurityGroupsMembers"},
				{Name: "InjectedFiles"},
				{Name: "InjectedFileContentBytes"},
				{Name: "InjectedFilePathBytes"},
			},
		}
		// computeVersion, _ := client.NovaV2().GetCurrentVersion()
		if !client.NovaV2().MicroVersionLargeEqual("2.36") {
			pt.ShortFields = append(pt.ShortFields, []common.Column{
				{Name: "FloatingIps"}, {Name: "FixedIps"},
			}...)
		}
		quotaSet, err := client.NovaV2().Quota().Show(projectId)
		utility.LogError(err, "show quota failed, %v", true)
		pt.Item = *quotaSet
		common.PrintPrettyItemTable(pt)
	},
}

func init() {
	QuotaCmd.AddCommand(show)
}
