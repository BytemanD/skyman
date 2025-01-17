package glance

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/utility"
)

var Image = &cobra.Command{Use: "image"}

var ImageList = &cobra.Command{
	Use:   "list",
	Short: "List images",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		if *imageListFlags.Visibility != "" &&
			!stringutils.ContainsString(glance.IMAGE_VISIBILITIES, *imageListFlags.Visibility) {
			return fmt.Errorf("invalid visibility %s, valid: %v", *imageListFlags.Visibility, glance.IMAGE_VISIBILITIES)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, _ []string) {
		query := url.Values{}
		if *imageListFlags.Name != "" {
			query.Set("name", *imageListFlags.Name)
		}
		if *imageListFlags.Visibility != "" {
			query.Set("visibility", *imageListFlags.Visibility)
		}
		if *imageListFlags.Limit != 0 {
			query.Set("limit", fmt.Sprintf("%d", *imageListFlags.Limit))
		}

		c := openstack.DefaultClient().GlanceV2()
		images, err := c.Images().List(query, int(*imageListFlags.Total))
		utility.LogError(err, "get imges failed", true)
		common.PrintImages(images, *imageListFlags.Long)
	},
}
var ImageShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show image",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().GlanceV2()

		image, err := c.Images().Find(args[0])
		utility.LogIfError(err, true, "Get image %s failed", args[0])
		common.PrintImage(*image, *imageShowFlags.Human)
	},
}
var imageCreate = &cobra.Command{
	Use:   "create",
	Short: "Create image",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		return imageCreateFlags.Valid()
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

		console.Info("create image name=%s", name)
		image, err := client.Images().Create(reqImage)
		utility.LogError(err, "Create image failed", true)
		if file != "" {
			console.Info("upload image")
			err = client.Images().Upload(image.Id, file)
			if err != nil {
				client.Images().Delete(image.Id)
				utility.LogError(err, "Upload image failed", true)
			}
			image, err = client.Images().Show(image.Id)
			utility.LogError(err, "get image failed", true)
		}
		common.PrintImage(*image, false)
	},
}

var imageDelete = &cobra.Command{
	Use:   "delete <image1> [<image2> ...]",
	Short: "Delete image",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().GlanceV2()

		for _, idOrName := range args {
			image, err := c.Images().Find(idOrName)
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
		c := openstack.DefaultClient().GlanceV2()

		image, err := c.Images().Find(args[0])
		utility.LogError(err, fmt.Sprintf("get image %v failed", args[0]), true)

		fileName := utility.OneOfString(
			*imageSaveFlags.File, fmt.Sprintf("%s.%s", image.Name, image.DiskFormat),
		)

		console.Info("Saving image to %s", fileName)
		err = c.Images().Download(image.Id, fileName, true)
		utility.LogError(err, fmt.Sprintf("download image %v failed", args[0]), true)
		console.Info("Image saved")
	},
}

var imageSet = &cobra.Command{
	Use:   "set <id or name>",
	Short: "Set image properties",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().GlanceV2()

		image, err := c.Images().Find(args[0])
		utility.LogError(err, "Get image failed", true)
		params := map[string]interface{}{}

		if *imageSetFlags.Name != "" {
			params["name"] = *imageSetFlags.Name
		}
		if *imageSetFlags.Visibility != "" {
			params["visibility"] = *imageSetFlags.Visibility
		}
		if *imageSetFlags.ContainerFormat != "" {
			params["container_format"] = *imageSetFlags.ContainerFormat
		}
		if *imageSetFlags.DiskFormat != "" {
			params["disk_format"] = *imageSetFlags.DiskFormat
		}
		if *imageSetFlags.KernelId != "" {
			params["kernel_id"] = *imageSetFlags.KernelId
		}
		if cmd.Flags().Changed("protect") {
			params["protect"] = *imageSetFlags.Protect
		}
		if len(params) == 0 {
			console.Warn("nothing to do")
			return
		}
		image, err = c.Images().Set(image.Id, params)
		utility.LogError(err, "update image failed", true)
		common.PrintImage(*image, false)
	},
}
var (
	imageListFlags   ImageListFlags
	imageShowFlags   ImageShowFlags
	imageCreateFlags ImageCreateFlags
	imageSaveFlags   ImageSaveFlags
	imageSetFlags    ImageSetFlags
)

func init() {
	imageListFlags = ImageListFlags{
		Name:       ImageList.Flags().StringP("name", "n", "", "Search by image name"),
		Limit:      ImageList.Flags().Uint("limit", 0, "Number of images to request in each paginated request"),
		Total:      ImageList.Flags().Uint("total", 0, "Maximum number of images to get"),
		Visibility: ImageList.Flags().String("visibility", "", "The visibility of the images to display."),
		Human:      ImageList.Flags().Bool("human", false, "Human size"),
		Long:       ImageList.Flags().BoolP("long", "l", false, "List additional fields in output"),
	}

	imageShowFlags = ImageShowFlags{
		Human: ImageShow.Flags().Bool("human", false, "Human size"),
	}
	imageCreateFlags = ImageCreateFlags{
		Name:            imageCreate.Flags().StringP("name", "n", "", "The name of image, if --name is empty use the name of file"),
		File:            imageCreate.Flags().String("file", "", "Local file that contains disk image to be uploaded during creation."),
		Protect:         imageCreate.Flags().Bool("protect", false, "Prevent image from being deleted"),
		Visibility:      imageCreate.Flags().String("visibility", "private", "Scope of image accessibility Valid values"),
		OSDistro:        imageCreate.Flags().String("os-distro", "", "Common name of operating system distribution"),
		ContainerFormat: imageCreate.Flags().String("container-format", "", fmt.Sprintf("Format of the container. Valid:\n%v", glance.IMAGE_CONTAINER_FORMATS)),
		DiskFormat:      imageCreate.Flags().String("disk-format", "", fmt.Sprintf("Format of the disk. Valid:\n%v", glance.IMAGE_DISK_FORMATS)),
	}
	imageSaveFlags = ImageSaveFlags{
		File: imageSave.Flags().String("file", "", "Downloaded image save filename."),
	}
	imageSetFlags = ImageSetFlags{
		Name:            imageSet.Flags().String("file", "", "Set name of image"),
		Protect:         imageSet.Flags().Bool("protoct", false, "Set protect of image"),
		Visibility:      imageSet.Flags().String("visibility", "", "Set visibility of image"),
		ContainerFormat: imageSet.Flags().String("container-format", "", "Set container format of image"),
		DiskFormat:      imageSet.Flags().String("disk-format", "", "Set disk format of image"),
		OSDistro:        imageSet.Flags().String("os-distro", "", "Set name os distro image"),
		OSVersion:       imageSet.Flags().String("os-version", "", "Set os version of image"),
		KernelId:        imageSet.Flags().String("kernel-id", "", "Set os kernel id of image"),
	}

	Image.AddCommand(ImageList, ImageShow, imageCreate, imageDelete, imageSave, imageSet)
}
