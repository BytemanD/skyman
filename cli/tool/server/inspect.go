package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"text/template"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/easygo/pkg/terminal"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/spf13/cobra"
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
	- **UUID**: {{.Server.Id}}
	- **Root BDM Type**: {{.Server.RootBdmType}}
	- **InstanceName**: {{.Server.InstanceName}}
	- **Created**: {{.Server.Created}}
	- **Updated**: {{.Server.Updated}}
	- **LaunchedAt**: {{.Server.LaunchedAt}}
	- **TerminatedAt**: {{.Server.TerminatedAt}}
	- **Description**: {{.Server.Description}}
	- **Image**: {{.Server.Image.Name}} ( {{.Server.Image.Id}} )
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
- **Ram**: {{ humanRam .Server.Flavor}}
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
  - **Binding Profile**: {{ getPortBindingProfile $interface.PortId }}
{{end}}

## 磁盘
| **Device** | **Volume Id** | **Type**  | **Size** | **Image** |
|---|---|---|---|---|
{{ range $index, $volume := .Volumes }} {{$volume.Device}} |{{$volume.VolumeId}} | {{ getVolumeType $volume.VolumeId }} |{{ getVolumeSize $volume.VolumeId }} |{{ getVolumeImage $volume.VolumeId }}|
{{end}}
  
## 操作记录

|  **Start Time** |  **Request Id** | **Action Name** | **Message** |
|---|---|---|---|
{{ range $index, $action := .Actions }} {{$action.StartTime}} |{{$action.RequestId}} | {{ $action.Action }} |{{ $action.Message }} |
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
				data, _ := json.Marshal(port.BindingDetails)
				return string(data)
			} else {
				return "-"
			}
		},
		"getPortBindingProfile": func(portId string) string {
			if port, ok := serverInspect.InterfaceDetail[portId]; ok {
				data, _ := json.Marshal(port.BindingProfile)
				return string(data)
			} else {
				return "-"
			}
		},
		"humanRam": func(flavor nova.Flavor) string {
			return flavor.HumanRam()
		},
	})
	tmpl, _ = tmpl.Parse(source)
	tmpl.Execute(bufferWriter, serverInspect)
	bufferWriter.Flush()

	width := 120
	if curTerm := terminal.CurTerminal(); curTerm != nil {
		width = curTerm.Columns
	}
	result := markdown.Render(templateBuffer.String(), width, 4)
	fmt.Println(string(result))
}

func inspect(client *openstack.Openstack, serverId string) (*ServerInspect, error) {
	server, err := client.NovaV2().Servers().Show(serverId)
	if err != nil {
		return nil, err
	}
	logging.Info("get server image")
	image, err := client.GlanceV2().Images().Show(server.Image.Id)
	if err != nil {
		return nil, err
	}
	server.Image.Name = image.Name
	interfaceAttachmetns, err := client.NovaV2().Servers().ListInterfaces(serverId)
	if err != nil {
		return nil, err
	}
	logging.Info("list server volumes")
	volumeAttachments, err := client.NovaV2().Servers().ListVolumes(serverId)
	if err != nil {
		return nil, err
	}
	logging.Info("list server ations")
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
	portQuery.Set("device_id", server.Id)
	logging.Info("list server ports details")
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

var serverInspect = &cobra.Command{
	Use:   "inspect <id>",
	Short: "inspect server ",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		serverId := args[0]
		format, _ := cmd.Flags().GetString("format")

		serverInspect, err := inspect(client, serverId)
		utility.LogError(err, "inspect sever faield", true)

		switch format {
		case "json":
			output, err := stringutils.JsonDumpsIndent(serverInspect)
			if err != nil {
				logging.Fatal("print json failed, %s", err)
			}
			fmt.Println(output)
		case "yaml":
			output, err := common.GetYaml(serverInspect)
			if err != nil {
				logging.Fatal("print json failed, %s", err)
			}
			fmt.Println(output)
		default:
			serverInspect.Print()
		}
	},
}

func init() {

	ServerCommand.AddCommand(serverInspect)
}
