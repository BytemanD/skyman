package storage

import (
	"encoding/json"
	"strings"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/cinder"
)

func printResource(resource any, fields []common.Column) {
	pt := common.PrettyItemTable{
		Item:        resource,
		ShortFields: fields,
	}
	common.PrintPrettyItemTable(pt)
}
func printVolumeType(volumeType cinder.VolumeType) {
	printResource(
		volumeType,
		[]common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "IsPublic"}, {Name: "IsEncrypted"},
			{Name: "QosSpecsId"},
			{Name: "ExtraSpecs", Slot: func(item interface{}) interface{} {
				p, _ := item.(cinder.VolumeType)
				return strings.Join(p.GetExtraSpecsList(), "\n")
			}},
		},
	)
}

func printVolume(volume cinder.Volume) {
	printResource(
		volume,
		[]common.Column{
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
	)
}
func printSnapshot(snapshot cinder.Snapshot) {
	printResource(
		snapshot,
		[]common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Status"},
			{Name: "VolumeId"},
			{Name: "Size"},
			{Name: "Metadata", Slot: func(item interface{}) interface{} {
				p, _ := item.(cinder.Snapshot)
				if p.Metadata == nil {
					return ""
				}
				metatadata, _ := json.Marshal(p.Metadata)
				return string(metatadata)
			}},
			{Name: "Progress"},
			{Name: "ProjectId"},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
		},
	)
}
func printBackup(backup cinder.Backup) {
	printResource(
		backup,
		[]common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Status"},
			{Name: "VolumeId"},
			{Name: "Size"},
			{Name: "Metadata", Slot: func(item interface{}) interface{} {
				p, _ := item.(cinder.Snapshot)
				if p.Metadata == nil {
					return ""
				}
				metatadata, _ := json.Marshal(p.Metadata)
				return string(metatadata)
			}},
			{Name: "Progress"},
			{Name: "ProjectId"},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
		},
	)
}
