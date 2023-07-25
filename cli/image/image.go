package image

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var Image = &cobra.Command{Use: "image"}

var ImageList = &cobra.Command{
	Use:   "list",
	Short: "List images",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		// human, _ := cmd.Flags().GetBool("human")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		images := client.Image.ImageList(query)
		dataTable := cli.DataListTable{
			ShortHeaders: []string{"Id", "Name", "Status", "Size"},
			LongHeaders: []string{
				"DiskFormat", "ContainerFormat", "Visibility", "Protected"},
			HeaderLabel: map[string]string{
				"DiskFormat":      "Disk Format",
				"ContainerFormat": "Container Format"},
			SortBy: []table.SortBy{
				{Name: "Name", Mode: table.Asc},
			},
			ColumnConfigs: []table.ColumnConfig{
				{Number: 4, Align: text.AlignRight},
			},
		}
		dataTable.AddItems(images)
		dataTable.Print(long)
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
			} else if len(images) == 1 {
				image = &images[0]
			} else if len(images) > 1 {
				fmt.Println(err)
			}
		}
		if image != nil {
			image.PrintTable(human)
		}
	},
}

func init() {
	ImageList.Flags().BoolP("long", "l", false, "List additional fields in output")
	ImageList.Flags().StringP("name", "n", "", "Search by image name")

	Image.PersistentFlags().Bool("human", false, "Human size")

	Image.AddCommand(ImageList, ImageShow)
}
