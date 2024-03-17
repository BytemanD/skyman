package image

import (
	"strings"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/glance"
)

func printImage(img glance.Image, human bool) {
	pt := common.PrettyItemTable{
		Item: img,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Checksum"}, {Name: "Schema"},
			{Name: "DirectUrl"}, {Name: "Status"},
			{Name: "ContainerFormat"}, {Name: "DiskFormat"},
			{Name: "File"},
			{Name: "Size", Slot: func(item interface{}) interface{} {
				p, _ := item.(glance.Image)
				if human {
					return p.HumanSize()
				} else {
					return p.Size
				}
			}},
			{Name: "Properties", Slot: func(item interface{}) interface{} {
				p, _ := item.(glance.Image)
				return strings.Join(p.GetPropertyList(), "\n")
			}},
			{Name: "VirtualSize"}, {Name: "ProcessInfo"}, {Name: "Protected"},
			{Name: "Visibility"},
			{Name: "OSHashAlgo", Text: "OS Hash Algo"},
			{Name: "OSHashValue", Text: "OS Hash Value"},
			{Name: "Tags"}, {Name: "Owner"},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
		},
	}
	common.PrintPrettyItemTable(pt)
}
