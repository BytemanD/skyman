package networking

import (
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Security = &cobra.Command{Use: "security"}

var group = &cobra.Command{Use: "group"}
var SG = &cobra.Command{Use: "sg", Aliases: []string{"security group"}}

var sgList = &cobra.Command{
	Use:   "list",
	Short: "List security groups",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient()

		long, _ := cmd.Flags().GetBool("long")
		projectIdOrName, _ := cmd.Flags().GetString("project")
		query := url.Values{}

		if projectIdOrName != "" {
			project, err := c.KeystoneV3().Project().Found(projectIdOrName)
			utility.LogError(err, "get project failed", true)
			query.Set("project_id", project.Id)
		}

		sgs, err := c.NeutronV2().SecurityGroup().List(query)
		utility.LogError(err, "list security group failed", true)

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Description"}, {Name: "ProjectId"},
			},
			LongColumns: []common.Column{
				{Name: "Tags"}, {Name: "Default"},
			},
		}
		pt.AddItems(sgs)
		common.PrintPrettyTable(pt, long)
	},
}
var sgShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show security group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient()

		sg, err := c.NeutronV2().SecurityGroup().Found(args[0])
		utility.LogError(err, "get security group failed", true)
		pt := common.PrettyItemTable{
			Item: *sg,
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"},
				{Name: "Description"},
				{Name: "RevisionNumber"},
				{Name: "CreatedAt"},
				{Name: "UpdatedAt"},
				{Name: "Rules", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.SecurityGroup)
					rules := []string{}
					for _, rule := range p.Rules {
						rules = append(rules, rule.String())
					}
					return strings.Join(rules, "\n")
				}},
				{Name: "ProjectId"},
			},
		}
		common.PrintPrettyItemTable(pt)
	},
}

func init() {
	sgList.Flags().BoolP("long", "l", false, "List additional fields in output")
	sgList.Flags().StringP("project", "", "", "List according to the project")

	group.AddCommand(sgList, sgShow)
	Security.AddCommand(group)
	SG.AddCommand(sgList, sgShow)
}
