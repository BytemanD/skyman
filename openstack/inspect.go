package openstack

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	markdown "github.com/MichaelMure/go-term-markdown"
)

type ServerInspect struct {
	Server          nova.Server                `json:"server"`
	Interfaces      []nova.InterfaceAttachment `json:"interfaces"`
	Volumes         []nova.VolumeAttachment    `json:"volumes"`
	PowerState      string                     `json:"powerState"`
	InterfaceDetail map[string]neutron.Port    `json:"interfaceDetail"`
	VolumeDetail    map[string]cinder.Volume   `json:"volumeDetail"`
	Actions         []nova.InstanceAction      `json:"actions"`
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

func (client Openstack) ServerInspect(serverId string) (*ServerInspect, error) {
	server, err := client.NovaV2().Servers().Show(serverId)
	if err != nil {
		return nil, err
	}
	interfaceAttachmetns, err := client.NovaV2().Servers().ListInterfaces(serverId)
	if err != nil {
		return nil, err
	}
	volumeAttachments, err := client.NovaV2().Servers().ListVolumes(serverId)
	if err != nil {
		return nil, err
	}
	actions, err := client.NovaV2().Servers().ListActions(serverId)
	if err != nil {
		return nil, err
	}
	serverInspect := ServerInspect{
		Server:          *server,
		Interfaces:      interfaceAttachmetns,
		Volumes:         volumeAttachments,
		InterfaceDetail: map[string]neutron.Port{},
		VolumeDetail:    map[string]cinder.Volume{},
		Actions:         actions,
	}

	portQuery := url.Values{}
	ports, err := client.NeutronV2().Ports().List(portQuery)
	if err != nil {
		return nil, err
	}
	for _, port := range ports {
		serverInspect.InterfaceDetail[port.Id] = port
	}

	for _, volume := range serverInspect.Volumes {
		vol, err := client.CinderV2().Volumes().Show(volume.VolumeId)
		utility.LogError(err, "get volume failed", true)
		serverInspect.VolumeDetail[volume.VolumeId] = *vol
	}
	return &serverInspect, nil
}
