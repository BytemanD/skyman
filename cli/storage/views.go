package storage

import (
	"strings"

	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/storage"
)

func printVolume(volume storage.Volume) {
	dataTable := common.DataTable{
		Item: volume,
		ShortFields: []common.Field{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Status"}, {Name: "TaskStatus"},
			{Name: "Size"}, {Name: "Bootable"},
			{Name: "Attachments", Slot: func(item interface{}) interface{} {
				p, _ := item.(storage.Volume)
				return strings.Join(p.GetAttachmentList(), "\n")
			}},
			{Name: "VolumeType"},
			{Name: "Metadata", Slot: func(item interface{}) interface{} {
				p, _ := item.(storage.Volume)
				return strings.Join(p.GetMetadataList(), "\n")
			}},
			{Name: "AvailabilityZone"}, {Name: "Host"},
			{Name: "Multiattach"}, {Name: "GroupId"}, {Name: "SourceVolid"},
			{Name: "VolumeImageMetadata", Slot: func(item interface{}) interface{} {
				p, _ := item.(storage.Volume)
				return strings.Join(p.GetImageMetadataList(), "\n")
			}},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
			{Name: "UserId"}, {Name: "TenantId"},
		},
	}
	common.PrintDataTable(dataTable)
}
