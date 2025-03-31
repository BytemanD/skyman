package glance

import (
	"fmt"

	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/duke-git/lancet/v2/slice"
)

type ImageListFlags struct {
	Name       *string
	Human      *bool
	Visibility *string

	Total *uint
	// page size
	Limit *uint

	Long *bool
}
type ImageShowFlags struct {
	Bytes *bool
}
type ImageCreateFlags struct {
	Name            *string
	File            *string
	Protect         *bool
	Visibility      *string
	OSDistro        *string
	ContainerFormat *string
	DiskFormat      *string
}
type ImageSaveFlags struct {
	Name      *string
	Output    *string
	OverWrite *bool
}

func (f ImageCreateFlags) Valid() error {
	if *f.ContainerFormat == "" {
		return fmt.Errorf("must provide --container-format when using --file")
	}
	if *f.DiskFormat == "" {
		return fmt.Errorf("must provide --disk-format when using --file")
	}
	if *f.File != "" {
		if *f.ContainerFormat == "" {
			return fmt.Errorf("must provide --container-format when using --file")
		}
		if *f.DiskFormat == "" {
			return fmt.Errorf("must provide --disk-format when using --file")
		}
	} else if *f.Name == "" {
		return fmt.Errorf("must provide --name when not using --file")
	}
	if *f.ContainerFormat != "" &&
		!slice.Contain(glance.IMAGE_CONTAINER_FORMATS, *f.ContainerFormat) {
		return fmt.Errorf("invalid container format, valid: %v", glance.IMAGE_CONTAINER_FORMATS)
	}
	if *f.DiskFormat != "" &&
		!slice.Contain(glance.IMAGE_DISK_FORMATS, *f.DiskFormat) {
		return fmt.Errorf("invalid disk format, valid: %v", glance.IMAGE_DISK_FORMATS)
	}
	if *f.Visibility != "" && !slice.Contain(glance.IMAGE_VISIBILITIES, *f.Visibility) {
		return fmt.Errorf("invalid visibility, valid: %v", glance.IMAGE_VISIBILITIES)
	}
	return nil
}

type ImageSetFlags struct {
	Name            *string
	Protect         *bool
	Visibility      *string
	ContainerFormat *string
	DiskFormat      *string
	KernelId        *string
	OSDistro        *string
	OSVersion       *string
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
