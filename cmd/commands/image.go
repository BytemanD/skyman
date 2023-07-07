package commands

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var Image = &cobra.Command{Use: "image"}

var ImageList = &cobra.Command{
	Use:   "list",
	Short: "List images",
	Run: func(cmd *cobra.Command, _ []string) {
		imageClient := getImageClient()

		long, _ := cmd.Flags().GetBool("long")
		human, _ := cmd.Flags().GetBool("human")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		images := imageClient.ImageList(query)
		images.PrintTable(long, human)
	},
}
var ImageShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show image",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageClient := getImageClient()
		id := args[0]
		human, _ := cmd.Flags().GetBool("human")
		image, err := imageClient.ImageShow(id)
		if err != nil {
			images := imageClient.ImageListByName(id)
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
	Image.AddCommand(ImageList)
	Image.AddCommand(ImageShow)
}
