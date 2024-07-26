package storage

import (
	"strings"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/cinder"
)

func printVolume(volume cinder.Volume) {
	pt := common.PrettyItemTable{
		Item: volume,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Status"}, {Name: "TaskStatus"},
			{Name: "Size"}, {Name: "Bootable"},
			{Name: "Attachments", Slot: func(item interface{}) interface{} {
				p, _ := item.(cinder.Volume)
				return strings.Join(p.GetAttachmentList(), "\n")
			}},
			{Name: "VolumeType"},
			{Name: "Metadata", Slot: func(item interface{}) interface{} {
				p, _ := item.(cinder.Volume)
				return strings.Join(p.GetMetadataList(), "\n")
			}},
			{Name: "AvailabilityZone"}, {Name: "Host"},
			{Name: "Multiattach"}, {Name: "GroupId"}, {Name: "SourceVolid"},
			{Name: "VolumeImageMetadata", Slot: func(item interface{}) interface{} {
				p, _ := item.(cinder.Volume)
				return strings.Join(p.GetImageMetadataList(), "\n")
			}},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
			{Name: "UserId"}, {Name: "TenantId"},
		},
	}
	common.PrintPrettyItemTable(pt)
}
func printVolumeType(volumeType cinder.VolumeType) {
	pt := common.PrettyItemTable{
		Item: volumeType,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "IsPublic"}, {Name: "IsEncrypted"},
			{Name: "QosSpecsId"},
			{Name: "ExtraSpecs", Slot: func(item interface{}) interface{} {
				p, _ := item.(cinder.VolumeType)
				return strings.Join(p.GetExtraSpecsList(), "\n")
			}},
		},
	}
	common.PrintPrettyItemTable(pt)
}
func printSnapshot(volume cinder.Snapshot) {
	pt := common.PrettyItemTable{
		Item: volume,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Status"},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
			{Name: "UserId"}, {Name: "TenantId"},
		},
	}
	common.PrintPrettyItemTable(pt)
}
