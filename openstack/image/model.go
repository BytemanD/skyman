package image

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/BytemanD/stackcrud/openstack/common"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
)

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
	DiskFormat      string   `json:"disk_format"`
	ContainerFormat string   `json:"container_format"`
	DirectUrl       string   `json:"direct_url"`
	Checksum        string   `json:"checksum"`
	Size            uint     `json:"size,omitempty"`
	VirtualSize     uint     `json:"virtual_size,omitempty"`
	MinDisk         uint     `json:"min_disk,omitempty"`
	MinRam          uint     `json:"min_ram,omitempty"`
	Owner           string   `json:"owner,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Protected       bool     `json:"protected,omitempty"`
	Visibility      bool     `json:"visibility,omitempty"`
	ProcessInfo     float32  `json:"progress_info,omitempty"`
	OSHashAlgo      string   `json:"os_hash_algo,omitempty"`
	OSHashValue     string   `json:"os_hash_value,omitempty"`
	Schema          string   `json:"schema,omitempty"`
}
type Images []Image

type ImageBody Image

func (image Image) PrintTable(human bool) {
	header := table.Row{"Property", "Value"}
	tableWriter := table.NewWriter()
	// Use reject
	tableWriter.AppendRows([]table.Row{
		{"ID", image.Id}, {"name", image.Name}, {"description", image.Description},

		{"direct_url", image.DirectUrl},
		{"checksum", image.Checksum},
		{"status", image.Status},
		{"disk_format", image.DiskFormat},
		{"container_format", image.ContainerFormat},
		{"size", image.Size},
		{"virtual_size", image.VirtualSize},
		{"process_info", image.ProcessInfo},
		{"peojected", image.Protected},
		{"os_hash_algo", image.OSHashAlgo},
		{"os_hash_value", image.OSHashValue},
		{"tags", image.Tags},
		{"owner", image.Owner},
		{"shema", image.Schema},

		{"created", image.CreatedAt},
		{"updated", image.UpdatedAt},
	})
	tableWriter.AppendHeader(header)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()
}

type ImagesBody struct {
	Images Images `json:"images"`
}

func (images Images) PrintTable(long bool, human bool) {
	header := table.Row{"ID", "Name", "Status", "Size"}
	if long {
		header = append(header, "Disk Format", "Container Format",
			"Visibility", "Protected")
	}
	tableWriter := table.NewWriter()
	for _, image := range images {
		row := table.Row{image.Id, image.Name, image.Status}
		if human {
			row = append(row, humanSize(image.Size))
		} else {
			row = append(row, image.Size)
		}
		if long {
			row = append(row, image.DiskFormat, image.ContainerFormat,
				image.Visibility, image.Protected)
		}
		tableWriter.SortBy([]table.SortBy{
			{Name: "Name", Mode: table.Asc},
		})
		tableWriter.AppendRow(row)
	}

	// tableWriter.SetStyle(table.StyleLight)
	tableWriter.AppendHeader(header)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
	})
	tableWriter.Render()
}
