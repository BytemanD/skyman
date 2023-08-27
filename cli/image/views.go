package image

import (
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/image"
)

func printImage(img image.Image, human bool) {
	dataTable := common.DataTable{
		Item: img,
		ShortFields: []common.Field{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Checksum"},
			{Name: "Schema"},
			{Name: "DirectUrl"},
			{Name: "Status"},
			{Name: "ContainerFormat"},
			{Name: "DiskFormat"},
			{Name: "Size"}, {Name: "VirtualSize"},

			{Name: "ProcessInfo"},
			{Name: "Protected"},
			{Name: "OSHashAlgo"}, {Name: "OSHashValue"},
			{Name: "Tags"},
			{Name: "Owner"},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
		},
		Slots: map[string]func(item interface{}) interface{}{
			"Size": func(item interface{}) interface{} {
				p, _ := item.(image.Image)
				if human {
					return p.HumanSize()
				} else {
					return p.Size
				}
			},
		},
	}
	dataTable.Print(false)
}
