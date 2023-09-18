package image

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/image"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var Image = &cobra.Command{Use: "image"}

var ImageList = &cobra.Command{
	Use:   "list",
	Short: "List images",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		human, _ := cmd.Flags().GetBool("human")
		name, _ := cmd.Flags().GetString("name")
		limit, _ := cmd.Flags().GetInt("limit")
		pageSize, _ := cmd.Flags().GetUint("page-size")

		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		if pageSize != 0 {
			query.Set("limit", fmt.Sprintf("%d", pageSize))
		}
		images := client.Image.ImageList(query, limit)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "Status", AutoColor: true},
				{Name: "Size", Slot: func(item interface{}) interface{} {
					p, _ := item.(image.Image)
					if human {
						return p.HumanSize()
					} else {
						return p.Size
					}
				}},
			},
			LongColumns: []common.Column{
				{Name: "DiskFormat"}, {Name: "ContainerFormat"},
				{Name: "Visibility"}, {Name: "Protected"},
			},
			ColumnConfigs: []table.ColumnConfig{{Number: 4, Align: text.AlignRight}},
		}

		pt.AddItems(images)
		common.PrintPrettyTable(pt, long)
	},
}
var ImageShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show image",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		idOrName := args[0]
		human, _ := cmd.Flags().GetBool("human")
		image, err := client.Image.ImageFound(idOrName)
		common.LogError(err, "Get image faled", true)
		printImage(*image, human)
	},
}

func init() {
	ImageList.Flags().BoolP("long", "l", false, "List additional fields in output")
	ImageList.Flags().StringP("name", "n", "", "Search by image name")
	ImageList.Flags().Uint("page-size", 0, "Number of images to request in each paginated request")
	ImageList.Flags().Int("limit", 0, "Maximum number of images to get")

	Image.PersistentFlags().Bool("human", false, "Human size")

	Image.AddCommand(ImageList, ImageShow)
}
