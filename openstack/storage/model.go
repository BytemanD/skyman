package storage

import (
	"fmt"
	"os"
	"strings"

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
		return fmt.Sprintf("%f GB", float32(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%2f MB", float32(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%f KB", float32(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

type Attachment struct {
	Id           string `json:"id,omitempty"`
	ServerId     string `json:"server_id,omitempty"`
	AttachmentId string `json:"attachment_id,omitempty"`
	Device       string `json:"device,omitempty"`
	AttachmentAt string `json:"attached_at,omitempty"`
	HostName     string `json:"host_name,omitempty"`
	VolumeId     string `json:"volume_id,omitempty"`
}
type Volume struct {
	common.Resource
	Size uint `json:"size,omitempty"`

	Attachments []Attachment `json:"attachments,omitempty"`
}
type Volumes []Volume

type VolumesBody struct {
	Volumes Volumes `json:"volumes"`
}

type VolumeBody Volume

func (volume Volume) GetAttachmentStrings() []string {
	attachmentString := []string{}
	for _, attachment := range volume.Attachments {
		attachmentString = append(attachmentString,
			fmt.Sprintf("Attached to %s on %s", attachment.ServerId, attachment.Device),
		)
	}
	return attachmentString
}

func (volume Volume) PrintTable(human bool) {
	header := table.Row{"Property", "Value"}
	tableWriter := table.NewWriter()
	// Use reject
	tableWriter.AppendRows([]table.Row{
		{"ID", volume.Id}, {"name", volume.Name},
		{"description", volume.Description},
		{"status", volume.Status},
		{"size", volume.Size},

		{"created", volume.CreatedAt},
		{"updated", volume.UpdatedAt},
	})
	tableWriter.AppendHeader(header)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()
}

func (volumes Volumes) PrintTable(long bool, human bool) {
	header := table.Row{"ID", "Name", "Status", "Size", "Attached to"}
	if long {
		header = append(header, "Properties")
	}
	tableWriter := table.NewWriter()
	for _, volume := range volumes {
		row := table.Row{volume.Id, volume.Name, volume.Status}
		if human {
			row = append(row, humanSize(volume.Size))
		} else {
			row = append(row, volume.Size)
		}
		row = append(row, strings.Join(volume.GetAttachmentStrings(), "\n"))
		if long {
			row = append(row, "TODO")
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
	tableWriter.Render()
}
