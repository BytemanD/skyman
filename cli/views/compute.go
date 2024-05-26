package views

import (
	"encoding/json"
	"fmt"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/nova"
)

func PrintServer(server nova.Server) {
	pt := common.PrettyItemTable{
		Item: server,
		ShortFields: []common.Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
			{Name: "Flavor:original_name", Text: "Flavor:original_name",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.OriginalName
				}},
			{Name: "Flavor:ram", Text: "Flavor:ram",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.Ram
				}},
			{Name: "Flavor:vcpus", Text: "Flavor:vcpus",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.Vcpus
				}},
			{Name: "Flavor:disk", Text: "Flavor:disk",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.Disk
				}},
			{Name: "Flavor:swap", Text: "Flavor:swap",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Flavor.Swap
				}},
			{Name: "Flavor:extra_specs", Text: "Flavor:extra_specs",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.GetFlavorExtraSpecsString()

				}},
			{Name: "Image",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					if p.Image.Name != "" {
						return fmt.Sprintf("%s (%s)", p.Image.Name, p.Image.Id)
					} else {
						return p.Image.Id
					}
				}},
			{Name: "AZ", Text: "AZ"}, {Name: "Host"},
			{Name: "Status"}, {Name: "TaskState"}, {Name: "PowerState"},
			{Name: "RootBdmType"}, {Name: "RootDeviceName"},
			{Name: "SecurityGroups", Slot: func(item interface{}) interface{} {
				p := item.(nova.Server)
				bytes, _ := json.Marshal(p.SecurityGroups)
				return string(bytes)
			}},
			{Name: "Created"}, {Name: "LaunchedAt"}, {Name: "Updated"}, {Name: "TerminatedAt"},

			{Name: "Fault:code", Text: "Fault:code",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Fault.Code
				}},
			{Name: "Fault:message", Text: "Fault:message",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Fault.Message
				}},
			{Name: "Fault:details", Text: "Fault:details",
				Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Server)
					return p.Fault.Details
				}},
			{Name: "UserId"},
		},
	}
	common.PrintPrettyItemTable(pt)
}
