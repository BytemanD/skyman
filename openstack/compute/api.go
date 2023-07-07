package compute

import (
	"encoding/json"
	"fmt"
	"net/url"
	netUrl "net/url"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

func (client ComputeClientV2) ServerList(query netUrl.Values) []Server {
	serversBody := ServersBody{}
	client.List("servers", query, &serversBody)
	return serversBody.Servers
}
func (client ComputeClientV2) ServerListDetails(query netUrl.Values) []Server {
	serversBody := ServersBody{}
	client.Show("servers", "detail", query, &serversBody)
	return serversBody.Servers
}
func (client ComputeClientV2) ServerListDetailsByName(name string) []Server {
	query := url.Values{}
	query.Set("name", name)
	return client.ServerListDetails(query)
}
func (client ComputeClientV2) ServerShow(id string) (*Server, error) {
	serverBody := ServerBody{}
	err := client.Show("servers", id, nil, &serverBody)
	return serverBody.Server, err
}

func (client ComputeClientV2) ServerDelete(id string) error {
	return client.Delete("servers", id)
}
func (client ComputeClientV2) getBlockDeviceMappingV2(imageRef string) BlockDeviceMappingV2 {
	return BlockDeviceMappingV2{
		BootIndex:          0,
		UUID:               imageRef,
		VolumeSize:         10,
		SourceType:         "image",
		DestinationType:    "volume",
		DeleteOnTemination: true,
	}
}
func (client ComputeClientV2) ServerCreate(options ServerOpt) (*Server, error) {
	if options.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if options.Flavor == "" {
		return nil, fmt.Errorf("flavor is empty")
	}
	if options.Image == "" || (len(options.BlockDeviceMappingV2) > 0 && options.BlockDeviceMappingV2[0].UUID == "") {
		return nil, fmt.Errorf("image is empty")
	}

	if options.Networks == nil {
		options.Networks = "none"
	}

	body, _ := json.Marshal(ServeCreaterBody{Server: options})
	serverBody := ServerBody{}
	var createErr error
	if options.BlockDeviceMappingV2 != nil {
		createErr = client.Create("os-volumes_boot", body, &serverBody)
	} else {
		createErr = client.Create("servers", body, &serverBody)
	}
	return serverBody.Server, createErr
}
func (client ComputeClientV2) WaitServerCreate(options ServerOpt) (*Server, error) {
	server, err := client.ServerCreate(options)
	if err != nil {
		return server, err
	}
	if server.Id == "" {
		return server, fmt.Errorf("create server failed")
	}
	return client.WaitServerStatusSecond(server.Id, "ACTIVE", 5)
}

func (client ComputeClientV2) WaitServerStatusSecond(serverId string, status string, second int) (*Server, error) {
	// var server Server
	for {
		server, err := client.ServerShow(serverId)
		if err != nil {
			return server, err
		}
		logging.Debug("server stauts is %s", server.Status)
		switch strings.ToUpper(server.Status) {
		case "ERROR":
			return server, fmt.Errorf("server status is error, message: %s", server.Fault.Message)
		case strings.ToUpper(status):
			return server, nil
		}
		time.Sleep(time.Second * time.Duration(second))
	}
}

func (client ComputeClientV2) WaitServerStatus(serverId string, status string) (*Server, error) {
	return client.WaitServerStatusSecond(serverId, status, 1)
}

func (client ComputeClientV2) WaitServerDeleted(id string) {
	client.ServerDelete(id)
	for {
		server, err := client.ServerShow(id)
		if server.Id == "" || err != nil {
			break
		}
		logging.Debug("servers status is %s", server.Status)
		time.Sleep(time.Second * 2)
	}
}
