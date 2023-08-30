package image

import (
	"fmt"
	"net/url"
	"os"

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

		id := args[0]
		human, _ := cmd.Flags().GetBool("human")
		image, err := client.Image.ImageShow(id)
		if err != nil {
			images := client.Image.ImageListByName(id)
			if len(images) > 1 {
				fmt.Printf("Found multi images named %s\n", id)
				os.Exit(1)
			} else if len(images) == 1 {
				image = &images[0]
			} else if len(images) > 1 {
				fmt.Println(err)
				os.Exit(1)
			}
		}
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
