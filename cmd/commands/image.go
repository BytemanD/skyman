package commands

import (
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
		imagesTable := ImagesTable{Images: imageClient.ImageList(nil)}
		imagesTable.Print(long, human)
	},
}

func init() {
	// Server list flags
	ImageList.Flags().BoolP("long", "l", false, "List additional fields in output")
	ImageList.Flags().Bool("human", false, "Human size")

	Image.AddCommand(ImageList)
}
