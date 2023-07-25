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
	Size       uint   `json:"size,omitempty"`
	VolumeType string `json:"volume_type,omitempty"`
	Bootable   string `json:"bootable"`

	Attachments         []Attachment      `json:"attachments"`
	Metadata            map[string]string `json:"metadata"`
	AvailabilityZone    string            `json:"availability_zone"`
	Host                string            `json:"os-vol-host-attr:host"`
	Multiattach         bool              `json:"multiattach"`
	SourceVolid         string            `json:"source_volid"`
	GroupId             string            `json:"group_id"`
	TaskStatus          string            `json:"task_status"`
	VolumeImageMetadata map[string]string `json:"volume_image_metadata"`
}

type Volumes []Volume

type VolumesBody struct {
	Volumes Volumes `json:"volumes"`
}

type VolumeBody struct {
	Volume *Volume `json:"volume"`
}

func (volume Volume) GetAttachmentList() []string {
	attachmentList := []string{}
	for _, attachment := range volume.Attachments {
		attachmentList = append(attachmentList,
			fmt.Sprintf("%s @ %s", attachment.Device, attachment.ServerId),
		)
	}
	return attachmentList
}
func (volume Volume) IsBootable() bool {
	return volume.Bootable == "true"
}

func (volume Volume) GetMetadataList() []string {
	metadataList := []string{}
	for k, v := range volume.Metadata {
		metadataList = append(metadataList, fmt.Sprintf("%s=%s", k, v))
	}
	return metadataList
}
func (volume Volume) GetImageMetadataList() []string {
	metadataList := []string{}
	for k, v := range volume.VolumeImageMetadata {
		metadataList = append(metadataList, fmt.Sprintf("%s=%s", k, v))
	}
	return metadataList
}
func (volume Volume) PrintTable() {
	header := table.Row{"Property", "Value"}
	tableWriter := table.NewWriter()
	// Use reject
	tableWriter.AppendRows([]table.Row{
		{"Id", volume.Id}, {"name", volume.Name},
		{"description", volume.Description},
		{"status", volume.Status},
		{"task_status", volume.TaskStatus},
		{"size", volume.Size},
		{"bootable", volume.Bootable},
		{"attachments", strings.Join(volume.GetAttachmentList(), "\n")},
		{"volume_type", volume.VolumeType},
		{"metadata", strings.Join(volume.GetMetadataList(), "\n")},
		{"availability_zone", volume.AvailabilityZone},
		{"host", volume.Host},
		{"multiattach", volume.Multiattach},
		{"group_id", volume.GroupId},
		{"source_volid", volume.SourceVolid},
		{"volume_image_metadata", strings.Join(volume.GetImageMetadataList(), "\n")},

		{"created", volume.CreatedAt},
		{"updated", volume.UpdatedAt},
	})
	tableWriter.AppendHeader(header)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.Render()
}
