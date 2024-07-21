package image

import (
	"fmt"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/utility"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var Image = &cobra.Command{Use: "image"}

var ImageList = &cobra.Command{
	Use:   "list",
	Short: "List images",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		visibility, _ := cmd.Flags().GetString("visibility")
		if visibility != "" && !stringutils.ContainsString(glance.IMAGE_VISIBILITIES, visibility) {
			return fmt.Errorf("invalid visibility, valid: %v", glance.IMAGE_VISIBILITIES)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")
		human, _ := cmd.Flags().GetBool("human")
		name, _ := cmd.Flags().GetString("name")
		limit, _ := cmd.Flags().GetInt("limit")
		pageSize, _ := cmd.Flags().GetUint("page-size")
		visibility, _ := cmd.Flags().GetString("visibility")

		query := utility.UrlValues(map[string]string{
			"name":       name,
			"visibility": visibility,
		})
		if pageSize != 0 {
			query.Set("limit", fmt.Sprintf("%d", pageSize))
		}

		c := openstack.DefaultClient().GlanceV2()
		images, err := c.Images().List(query, limit)
		utility.LogError(err, "get imges failed", true)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "Status", AutoColor: true},
				{Name: "Size", Align: text.AlignRight,
					Slot: func(item interface{}) interface{} {
						p, _ := item.(glance.Image)
						if human {
							return p.HumanSize()
						} else {
							return p.Size
						}
					}},
				{Name: "DiskFormat"}, {Name: "ContainerFormat"},
			},
			LongColumns: []common.Column{
				{Name: "Visibility"}, {Name: "Protected"},
			},
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
		idOrName := args[0]
		human, _ := cmd.Flags().GetBool("human")
		c := openstack.DefaultClient().GlanceV2()

		image, err := c.Images().Found(idOrName)
		utility.LogError(err, "Get image failed", true)
		printImage(*image, human)
	},
}
var imageCreate = &cobra.Command{
	Use:   "create",
	Short: "Create image",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		containerFormat, _ := cmd.Flags().GetString("container-format")
		diskFormat, _ := cmd.Flags().GetString("disk-format")
		file, _ := cmd.Flags().GetString("file")
		visibility, _ := cmd.Flags().GetString("visibility")

		if file != "" {
			if containerFormat == "" {
				return fmt.Errorf("must provide --container-format when using --file")
			}
			if diskFormat == "" {
				return fmt.Errorf("must provide --disk-format when using --file")
			}
		} else if name == "" {
			return fmt.Errorf("must provide --name when not using --file")
		}
		if containerFormat != "" && !stringutils.ContainsString(glance.IMAGE_CONTAINER_FORMATS, containerFormat) {
			return fmt.Errorf("invalid container format, valid: %v", glance.IMAGE_CONTAINER_FORMATS)
		}
		if diskFormat != "" && !stringutils.ContainsString(glance.IMAGE_DISK_FORMATS, diskFormat) {
			return fmt.Errorf("invalid disk format, valid: %v", glance.IMAGE_DISK_FORMATS)
		}
		if visibility != "" && !stringutils.ContainsString(glance.IMAGE_VISIBILITIES, visibility) {
			return fmt.Errorf("invalid visibility, valid: %v", glance.IMAGE_VISIBILITIES)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient().GlanceV2()

		name, _ := cmd.Flags().GetString("name")
		containerFormat, _ := cmd.Flags().GetString("container-format")
		diskFormat, _ := cmd.Flags().GetString("disk-format")
		file, _ := cmd.Flags().GetString("file")

		protect, _ := cmd.Flags().GetBool("file")
		visibility, _ := cmd.Flags().GetString("visibility")
		// osDistro, _ := cmd.Flags().GetString("os-distro")

		reqImage := glance.Image{
			ContainerFormat: containerFormat,
			DiskFormat:      diskFormat,
			Protected:       protect,
			Visibility:      visibility,
		}
		if name == "" && file != "" {
			name, _ = common.PathExtSplit(file)
		}
		reqImage.Name = name

		c := openstack.DefaultClient().GlanceV2()

		logging.Info("create image")
		image, err := c.Images().Create(reqImage)
		utility.LogError(err, "Create image failed", true)
		if file != "" {
			logging.Info("upload image")
			err = client.Images().Upload(image.Id, file)
			utility.LogError(err, "Upload image failed", true)
			image, err = c.Images().Show(image.Id)
			utility.LogError(err, "get image failed", true)
		}

		printImage(*image, true)
	},
}

var imageDelete = &cobra.Command{
	Use:   "delete <image1> [<image2> ...]",
	Short: "Delete image",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().GlanceV2()

		for _, idOrName := range args {
			image, err := c.Images().Found(idOrName)
			if err != nil {
				utility.LogError(err, fmt.Sprintf("get image %v failed", idOrName), false)
				continue
			}
			err = c.Images().Delete(image.Id)
			if err != nil {
				utility.LogError(err, fmt.Sprintf("delete image %s failed", idOrName), false)
				continue
			}
			fmt.Printf("Requested to delete image %s\n", idOrName)
		}
	},
}

var imageSave = &cobra.Command{
	Use:     "save <image>",
	Short:   "Save image",
	Aliases: []string{"download"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileName, _ := cmd.Flags().GetString("file")

		c := openstack.DefaultClient().GlanceV2()
		image, err := c.Images().Found(args[0])
		utility.LogError(err, fmt.Sprintf("get image %v failed", args[0]), true)
		if fileName == "" {
			fileName = fmt.Sprintf("%s.%s", image.Name, image.DiskFormat)
		}

		logging.Info("Saving image to %s", fileName)
		err = c.Images().Download(image.Id, fileName, true)
		utility.LogError(err, fmt.Sprintf("download image %v failed", args[0]), true)
		logging.Info("Image saved")
	},
}

var IMAGE_ATTRIBUTIES = map[string]string{
	"name":             "name",
	"visibility":       "visibility",
	"container-format": "container_format",
	"disk-format":      "disk_format",
	"kernel-id":        "kernel_id",
	"os-distro":        "os_distro",
	"os-version":       "os_version",
}

var imageSet = &cobra.Command{
	Use:   "set <id or name>",
	Short: "Set image properties",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().GlanceV2()

		human, _ := cmd.Flags().GetBool("human")
		image, err := c.Images().Found(args[0])
		utility.LogError(err, "Get image failed", true)
		params := map[string]interface{}{}
		for flag, attribute := range IMAGE_ATTRIBUTIES {
			value, _ := cmd.Flags().GetString(flag)
			if value == "" {
				continue
			}
			params[attribute] = value
		}
		image, err = c.Images().Set(image.Id, params)
		utility.LogError(err, "update image failed", true)
		printImage(*image, human)
	},
}

func init() {
	ImageList.Flags().BoolP("long", "l", false, "List additional fields in output")
	ImageList.Flags().StringP("name", "n", "", "Search by image name")
	ImageList.Flags().Uint("page-size", 0, "Number of images to request in each paginated request")
	ImageList.Flags().Int("limit", 0, "Maximum number of images to get")
	ImageList.Flags().String("visibility", "", "The visibility of the images to display.")

	imageCreate.Flags().StringP("name", "n", "", "The name of image, if --name is empty use the name of file")
	imageCreate.Flags().String("file", "", "Local file that contains disk image to be uploaded during creation.")
	imageCreate.Flags().Bool("protect", false, "Prevent image from being deleted")
	imageCreate.Flags().String("visibility", "private", "Scope of image accessibility Valid values")

	imageCreate.Flags().String("os-distro", "", "Common name of operating system distribution")

	// TODO: show valid values.
	imageCreate.Flags().String("container-format", "", fmt.Sprintf("Format of the container. Valid:\n%v", glance.IMAGE_CONTAINER_FORMATS))
	imageCreate.Flags().String("disk-format", "", fmt.Sprintf("Format of the disk. Valid:\n%v", glance.IMAGE_DISK_FORMATS))

	imageSave.Flags().String("file", "", "Downloaded image save filename.")
	Image.PersistentFlags().Bool("human", false, "Human size")

	for k, v := range IMAGE_ATTRIBUTIES {
		imageSet.Flags().String(k, "", fmt.Sprintf("Set %s of image", v))
	}

	Image.AddCommand(ImageList, ImageShow, imageCreate, imageDelete, imageSave, imageSet)
}
