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
		verbose, _ := cmd.Flags().GetBool("verbose")
		serversTable := ImagesTable{Images: imageClient.ImageList(nil)}
		serversTable.Print(long, verbose)
	},
}

func init() {
	// Server list flags
	ImageList.Flags().BoolP("long", "l", false, "List additional fields in output")
	ImageList.Flags().BoolP("verbose", "v", false, "List verbose fields in output")

	Image.AddCommand(ImageList)
}
