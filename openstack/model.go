package openstack

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"strings"

	markdown "github.com/MichaelMure/go-term-markdown"

	"github.com/BytemanD/skyman/openstack/compute"
	"github.com/BytemanD/skyman/openstack/networking"
	"github.com/BytemanD/skyman/openstack/storage"
)

type ServerInspect struct {
	Server          compute.Server                `json:"server"`
	Interfaces      []compute.InterfaceAttachment `json:"interfaces"`
	Volumes         []compute.VolumeAttachment    `json:"volumes"`
	PowerState      string
	InterfaceDetail map[string]networking.Port `json:"interfaceDetail"`
	VolumeDetail    map[string]storage.Volume  `json:"volumeDetail"`
	Actions         []compute.InstanceAction   `json:"actions"`
}

func (serverInspect *ServerInspect) Print() {
	source := `
# Inspect For Instance {{.Server.Name}}

## 基本信息

1. 实例详情
	- **UUID(Name)**: {{.Server.Id}} ({{.Server.Name}})
	- **Root BDM Type**: {{.Server.RootBdmType}}
	- **InstanceName**: {{.Server.InstanceName}}
	- **Created**: {{.Server.Created}}
	- **Updated**: {{.Server.Updated}}
	- **LaunchedAt**: {{.Server.LaunchedAt}}
	- **TerminatedAt**: {{.Server.TerminatedAt}}
	- **Description**: {{.Server.Description}}
1. 状态
	- **VmState**: {{.Server.VmState}}
	- **PowerState**: {{.PowerState}}
	- **TaskState**: {{.Server.TaskState}}
1. 节点
	- **Availability Zone**: {{.Server.AZ}}
	- **Host**: {{.Server.Host}}

## 规格

- **Name**: {{.Server.Flavor.OriginalName}}
- **Vcpus**: {{.Server.Flavor.Vcpus}}
- **Ram**: {{.Server.Flavor.Ram}}
- **Extra specs**:
{{ range $key, $value := .Server.Flavor.ExtraSpecs }}
	1. {{$key}} = {{$value}}
{{end}}

## 网卡
{{ range $index, $interface := .Interfaces }}
1. **PortId**: {{$interface.PortId}}
  - **Mac Address**: {{$interface.MacAddr}}
  - **IP Address**: {{ range $index, $ip := $interface.FixedIps }} {{$ip.IpAddress}} {{end}}
  - **Binding Type**: vnic_type:{{ getPortVnicType $interface.PortId }} vif_type:{{ getPortVifType $interface.PortId }}
  - **Binding Detail**: {{ getPortBindingDetail $interface.PortId }}
{{end}}

## 磁盘
|  **Device** |  **Volume Id** | **Type**  | **Size** | **Image** |
|---|---|---|---|---|
{{ range $index, $volume := .Volumes }} {{$volume.Device}} |{{$volume.VolumeId}} | {{ getVolumeType $volume.VolumeId }} |{{ getVolumeSize $volume.VolumeId }} |{{ getVolumeImage $volume.VolumeId }}|
{{end}}
  
## 操作记录
{{ range $index, $action := .Actions }}
1. **{{$action.StartTime}}** {{$action.RequestId}} {{$action.Action}} {{$action.Message}}
{{end}}
`
	templateBuffer := bytes.NewBuffer([]byte{})
	bufferWriter := bufio.NewWriter(templateBuffer)
	serverInspect.PowerState = serverInspect.Server.GetPowerState()

	tmpl := template.New("inspect")
	tmpl = tmpl.Funcs(template.FuncMap{
		"getVolumeType": func(volumeId string) string {
			if vol, ok := serverInspect.VolumeDetail[volumeId]; ok {
				return vol.VolumeType
			} else {
				return "-"
			}
		},
		"getVolumeSize": func(volumeId string) string {
			if vol, ok := serverInspect.VolumeDetail[volumeId]; ok {
				return fmt.Sprintf("%d GB", vol.Size)
			} else {
				return "-"
			}
		},
		"getVolumeImage": func(volumeId string) string {
			if vol, ok := serverInspect.VolumeDetail[volumeId]; ok {
				var image string
				if imageId, ok := vol.VolumeImageMetadata["image_id"]; ok {
					image = imageId
				}
				if imageName, ok := vol.VolumeImageMetadata["image_name"]; ok {
					image += fmt.Sprintf("(%s)", imageName)
				}
				return image
			} else {
				return "-"
			}
		},
		"getPortVnicType": func(portId string) string {
			if port, ok := serverInspect.InterfaceDetail[portId]; ok {
				return port.BindingVnicType
			} else {
				return "-"
			}
		},
		"getPortVifType": func(portId string) string {
			if port, ok := serverInspect.InterfaceDetail[portId]; ok {
				return port.BindingVifType
			} else {
				return "-"
			}
		},
		"getPortBindingDetail": func(portId string) string {
			if port, ok := serverInspect.InterfaceDetail[portId]; ok {
				return strings.Join(port.VifDetailList(), ", ")
			} else {
				return "-"
			}
		},
	})
	tmpl, _ = tmpl.Parse(source)
	tmpl.Execute(bufferWriter, serverInspect)
	bufferWriter.Flush()
	result := markdown.Render(templateBuffer.String(), 120, 4)
	fmt.Println(string(result))
}
