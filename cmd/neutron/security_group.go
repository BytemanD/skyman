package neutron

import (
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/datatable"
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
			project, err := c.KeystoneV3().Project().Find(projectIdOrName)
			utility.LogError(err, "get project failed", true)
			query.Set("project_id", project.Id)
		}

		// result := c.NeutronV2().SecurityGroup().List2(query)
		// sgs, err := result.Items()
		// console.Debug("request id: %s", result.RequestId())
		sgs, err := c.NeutronV2().SecurityGroup().List(query)
		utility.LogError(err, "list security group failed", true)

		table := datatable.DataTable[neutron.SecurityGroup]{
			Items: sgs,
			Columns: []datatable.Column[neutron.SecurityGroup]{
				{Name: "Id"}, {Name: "Name"},
				{Name: "ProjectId"},
				{Name: "RevisionNumber"},
				{Name: "Rules", RenderFunc: func(item neutron.SecurityGroup) interface{} {
					rules := []string{}
					for _, rule := range item.Rules {
						rules = append(rules, rule.String())
					}
					return strings.Join(rules, "\n")
				}},
			},
			MoreColumns: []datatable.Column[neutron.SecurityGroup]{
				// {Name: "Description"},
				{Name: "CreatedAt"},
				{Name: "UpdatedAt"},
			},
		}
		common.PrintDataTable[neutron.SecurityGroup](&table, long)
	},
}
var sgShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show security group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient()

		sg, err := c.NeutronV2().SecurityGroup().Find(args[0])
		utility.LogError(err, "get security group failed", true)
		table := datatable.DataIterator[neutron.SecurityGroup]{
			Items: []neutron.SecurityGroup{*sg},
			Fields: []datatable.Field[neutron.SecurityGroup]{
				{Name: "Id"}, {Name: "Name"},
				{Name: "Description"},
				{Name: "RevisionNumber"},
				{Name: "CreatedAt"},
				{Name: "UpdatedAt"},
				{Name: "Rules", RenderFunc: func(item neutron.SecurityGroup) interface{} {
					rules := []string{}
					for _, rule := range item.Rules {
						rules = append(rules, rule.String())
					}
					return strings.Join(rules, "\n")
				}},
				{Name: "ProjectId"},
			},
		}
		common.PrintDataTable[neutron.SecurityGroup](&table, false)
	},
}

func init() {
	sgList.Flags().BoolP("long", "l", false, "List additional fields in output")
	sgList.Flags().StringP("project", "", "", "List according to the project")

	group.AddCommand(sgList, sgShow)
	Security.AddCommand(group)
	SG.AddCommand(sgList, sgShow)
}
