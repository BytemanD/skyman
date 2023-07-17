package commands

import (
	"fmt"
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
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		volumes := CLIENT.VolumeListDetail(query)
		volumes.PrintTable(long)
	},
}

var VolumeShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CLIENT := getVolumeClient()
		idOrName := args[0]
		volume, err := CLIENT.VolumeShow(idOrName)
		if err != nil {
			volumes := CLIENT.VolumeListDetailByName(idOrName)
			if len(volumes) > 1 {
				fmt.Printf("Found multi volumes named %s\n", idOrName)
			} else if len(volumes) == 1 {
				volume = &volumes[0]
			} else {
				fmt.Println(err)
			}
		}
		if volume != nil {
			volume.PrintTable()
		}
	},
}

func init() {
	VolumeList.Flags().BoolP("long", "l", false, "List additional fields in output")
	VolumeList.Flags().StringP("name", "n", "", "Search by volume name")

	Volume.AddCommand(VolumeList)
	Volume.AddCommand(VolumeShow)
}
