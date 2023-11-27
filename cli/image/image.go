package image

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	openstackCommon "github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/image"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var IMAGE_VISIBILITIES = []string{"public", "private", "community", "shared"}

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
		images, err := client.ImageClient().ImageList(query, limit)
		common.LogError(err, "get imges failed", true)
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
		image, err := client.ImageClient().ImageFound(idOrName)
		common.LogError(err, "Get image faled", true)
		printImage(*image, human)
	},
}
var imageCreate = &cobra.Command{
	Use:   "create <id or name>",
	Short: "Create image",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		containerFormat, _ := cmd.Flags().GetString("container-format")
		diskFormat, _ := cmd.Flags().GetString("disk-format")
		file, _ := cmd.Flags().GetString("file")
		visibility, _ := cmd.Flags().GetString("visibility")

		if file != "" {
			if containerFormat == "" {
				return fmt.Errorf("Must provide --container-format where using --file")
			}
			if diskFormat == "" {
				return fmt.Errorf("Must provide --disk-format when using --file")
			}
		}
		if visibility != "" {
			if !openstackCommon.ContainsString(IMAGE_VISIBILITIES, visibility) {
				return fmt.Errorf("invalid visibility, inlvalid: %v", IMAGE_VISIBILITIES)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		name, _ := cmd.Flags().GetString("name")
		containerFormat, _ := cmd.Flags().GetString("container-format")
		diskFormat, _ := cmd.Flags().GetString("disk-format")
		file, _ := cmd.Flags().GetString("file")

		protect, _ := cmd.Flags().GetBool("file")
		visibility, _ := cmd.Flags().GetString("visibility")
		// osDistro, _ := cmd.Flags().GetString("os-distro")

		reqImage := image.Image{
			ContainerFormat: containerFormat,
			DiskFormat:      diskFormat,
			Protected:       protect,
			Visibility:      visibility,
		}
		if name == "" && file != "" {
			name, _ = common.PathExtSplit(file)
		}
		reqImage.Name = name

		imageClient := client.ImageClient()
		image, err := imageClient.ImageCreate(reqImage)
		common.LogError(err, "Create image faled", true)
		if file != "" {
			err = imageClient.ImageUpload(image.Id, file)
			common.LogError(err, "Upload image failed", true)
			image, err = imageClient.ImageShow(image.Id)
			common.LogError(err, "get image failed", true)
		}

		printImage(*image, true)
	},
}

var imageDelete = &cobra.Command{
	Use:   "delete <image1> [<image2> ...]",
	Short: "Delete image",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, idOrName := range args {
			image, err := client.ImageClient().ImageFound(idOrName)
			if err != nil {
				common.LogError(err, fmt.Sprintf("get image %v failed", idOrName), false)
				continue
			}
			err = client.ImageClient().ImageDelete(image.Id)
			if err != nil {
				common.LogError(err, fmt.Sprintf("delete image %s failed", idOrName), false)
				continue
			}
			fmt.Printf("Requested to delete image %s\n", idOrName)
		}
	},
}
var imageSave = &cobra.Command{
	Use:   "save <image>",
	Short: "Save image",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		image, err := client.ImageClient().ImageFound(args[0])
		common.LogError(err, fmt.Sprintf("get image %v failed", args[0]), false)

		fileName, _ := cmd.Flags().GetString("file")
		if fileName == "" {
			fileName = fmt.Sprintf("%s.%s", image.Name, image.DiskFormat)
		}

		err = client.ImageClient().ImageDownload(image.Id, fileName, true)
		common.LogError(err, fmt.Sprintf("download image %v failed", args[0]), false)
		fmt.Printf("Saved image to %s.\n", fileName)
	},
}

func init() {
	ImageList.Flags().BoolP("long", "l", false, "List additional fields in output")
	ImageList.Flags().StringP("name", "n", "", "Search by image name")
	ImageList.Flags().Uint("page-size", 0, "Number of images to request in each paginated request")
	ImageList.Flags().Int("limit", 0, "Maximum number of images to get")

	imageCreate.Flags().StringP("name", "n", "", "The name of image")
	imageCreate.Flags().String("file", "", "Local file that contains disk image to be uploaded during creation.")
	imageCreate.Flags().Bool("protect", false, "Prevent image from being deleted")
	imageCreate.Flags().String("visibility", "private", "Scope of image accessibility Valid values")

	imageCreate.Flags().String("os-distro", "", "Common name of operating system distribution")

	// TODO: show valid values.
	imageCreate.Flags().String("container-format", "", "Format of the container.")
	imageCreate.Flags().String("disk-format", "", "Format of the disk.")

	imageSave.Flags().String("file", "", "Downloaded image save filename.")
	Image.PersistentFlags().Bool("human", false, "Human size")

	Image.AddCommand(ImageList, ImageShow, imageCreate, imageDelete, imageSave)
}
