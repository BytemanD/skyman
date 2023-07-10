package commands

import (
	"net/url"

	"github.com/spf13/cobra"
)

var Volume = &cobra.Command{Use: "volume"}

var VolumeList = &cobra.Command{
	Use:   "list",
	Short: "List volumes",
	Run: func(cmd *cobra.Command, _ []string) {
		CLIENT := getVolumeClient()

		long, _ := cmd.Flags().GetBool("long")
		human, _ := cmd.Flags().GetBool("human")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		volumes := CLIENT.VolumeListDetail(query)
		volumes.PrintTable(long, human)
	},
}

// var VolumeShow = &cobra.Command{
// 	Use:   "show <id or name>",
// 	Short: "Show volume",
// 	Args:  cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		CLIENT := getVolumeClient()
// 		id := args[0]
// 		human, _ := cmd.Flags().GetBool("human")
// 		volume, err := CLIENT.VolumeShow(id)
// 		if err != nil {
// 			volumes := CLIENT.ImageListByName(id)
// 			if len(volumes) > 1 {
// 				fmt.Printf("Found multi volumes named %s\n", id)
// 			} else if len(volumes) == 1 {
// 				volume = &volumes[0]
// 			} else if len(volumes) > 1 {
// 				fmt.Println(err)
// 			}
// 		}
// 		if volume != nil {
// 			volume.PrintTable(human)
// 		}
// 	},
// }

func init() {
	VolumeList.Flags().BoolP("long", "l", false, "List additional fields in output")
	VolumeList.Flags().StringP("name", "n", "", "Search by volume name")

	Volume.PersistentFlags().Bool("human", false, "Human size")
	Volume.AddCommand(VolumeList)
	// Image.AddCommand(VolumeShow)
}
