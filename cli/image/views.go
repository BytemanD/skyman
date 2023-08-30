package image

import (
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/image"
)

func printImage(img image.Image, human bool) {
	pt := common.PrettyItemTable{
		Item: img,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Checksum"}, {Name: "Schema"},
			{Name: "DirectUrl"}, {Name: "Status"},
			{Name: "ContainerFormat"}, {Name: "DiskFormat"},
			{Name: "Size", Slot: func(item interface{}) interface{} {
				p, _ := item.(image.Image)
				if human {
					return p.HumanSize()
				} else {
					return p.Size
				}
			}},
			{Name: "VirtualSize"}, {Name: "ProcessInfo"}, {Name: "Protected"},
			{Name: "OSHashAlgo", Text: "OS Hash Algo"},
			{Name: "OSHashValue", Text: "OS Hash Value"},
			{Name: "Tags"}, {Name: "Owner"},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
		},
	}
	common.PrintPrettyItemTable(pt)
}
