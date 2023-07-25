package storage

import (
	"strings"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/openstack/storage"
)

func printVolume(volume storage.Volume) {
	dataTable := cli.DataTable{
		Item: volume,
		ShortFields: []cli.Field{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Status"}, {Name: "TaskStatus"},
			{Name: "Size"}, {Name: "Bootable"},
			{Name: "Attachments"}, {Name: "VolumeType"},
			{Name: "Metadata"},
			{Name: "AvailabilityZone"}, {Name: "Host"},
			{Name: "Multiattach"}, {Name: "GroupId"}, {Name: "SourceVolid"},
			{Name: "VolumeImageMetadata"},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
			{Name: "UserId"}, {Name: "TenantId"},
		},
		Slots: map[string]func(item interface{}) interface{}{
			"Attachments": func(item interface{}) interface{} {
				p, _ := item.(storage.Volume)
				return strings.Join(p.GetAttachmentList(), "\n")
			},
			"Metadata": func(item interface{}) interface{} {
				p, _ := item.(storage.Volume)
				return strings.Join(p.GetMetadataList(), "\n")
			},
			"VolumeImageMetadata": func(item interface{}) interface{} {
				p, _ := item.(storage.Volume)
				return strings.Join(p.GetImageMetadataList(), "\n")
			},
		},
	}
	dataTable.Print(false)
}
