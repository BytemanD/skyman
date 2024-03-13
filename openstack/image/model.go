package image

import (
	"fmt"

	"github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/utility"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
)

var IMAGE_CONTAINER_FORMATS = []string{"bare", "ami", "ari", "aki", "bare",
	"ovf", "ova", "docker", "community", "shared"}
var IMAGE_DISK_FORMATS = []string{"ami", "ari", "aki", "vhd", "vhdx", "vmdk",
	"raw", "qcow2", "vdi", "iso", "ploop", "luks"}
var IMAGE_VISIBILITIES = []string{"public", "private", "community", "shared"}

func humanSize(size uint) string {
	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float32(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float32(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float32(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

type Image struct {
	common.Resource
	DiskFormat      string   `json:"disk_format,omitempty"`
	ContainerFormat string   `json:"container_format,omitempty"`
	DirectUrl       string   `json:"direct_url,omitempty"`
	Checksum        string   `json:"checksum,omitempty"`
	Size            uint     `json:"size,omitempty"`
	VirtualSize     uint     `json:"virtual_size,omitempty"`
	MinDisk         uint     `json:"min_disk,omitempty"`
	MinRam          uint     `json:"min_ram,omitempty"`
	Owner           string   `json:"owner,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Protected       bool     `json:"protected,omitempty"`
	Visibility      string   `json:"visibility,omitempty"`
	ProcessInfo     float32  `json:"progress_info,omitempty"`
	OSHashAlgo      string   `json:"os_hash_algo,omitempty"`
	OSHashValue     string   `json:"os_hash_value,omitempty"`
	Schema          string   `json:"schema,omitempty"`
	File            string   `json:"file,omitempty"`
	raw             map[string]interface{}
}

func (img Image) HumanSize() string {
	return humanSize(img.Size)
}

func (img Image) GetProperties() map[string]interface{} {
	proerties := map[string]interface{}{}
	tags := append(utility.GetStructTags(img), utility.GetStructTags(img.Resource)...)

	for k, v := range img.raw {
		if utility.StringsContain(tags, k) {
			continue
		}
		proerties[k] = v
	}
	delete(proerties, "self")
	return proerties
}
func (img Image) GetPropertyList() []string {

	properties := []string{}
	for k, v := range img.GetProperties() {
		properties = append(properties, fmt.Sprintf("%s=%s", k, v))
	}
	return properties
}

type Images []Image

type ImagesResp struct {
	Images []Image `json:"images,omitempty"`
	Next   string  `json:"next,omitempty"`
}

type AttributeOp struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
	Op    string      `json:"op"`
}
