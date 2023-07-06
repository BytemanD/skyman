package compute

import (
	"encoding/json"
	"fmt"
	netUrl "net/url"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

func (cmpCli ComputeClientV2) ServerList(query netUrl.Values) []Server {
	serversBody := ServersBody{}

	resp, _ := cmpCli.AuthClient.Get(
		cmpCli.getUrl("servers", ""), query, cmpCli.BaseHeaders)
	json.Unmarshal(resp.Body, &serversBody)
	return serversBody.Servers
}
func (cmpCli ComputeClientV2) ServerListDetails(query netUrl.Values) []Server {
	serversBody := ServersBody{}

	resp, _ := cmpCli.AuthClient.Get(
		cmpCli.getUrl("servers/detail", ""), query, cmpCli.BaseHeaders)
	json.Unmarshal(resp.Body, &serversBody)
	return serversBody.Servers
}
func (computeClient ComputeClientV2) ServerShow(id string) (Server, error) {
	resp, _ := computeClient.AuthClient.Get(
		computeClient.getUrl("servers", id), nil, computeClient.BaseHeaders)
	if err := resp.JudgeStatus(); err != nil {
		return Server{}, err
	}
	serverBody := ServerBody{}
	json.Unmarshal(resp.Body, &serverBody)
	return serverBody.Server, nil
}

func (computeClient ComputeClientV2) ServerDelete(id string) error {
	resp, err := computeClient.AuthClient.Delete(
		computeClient.getUrl("servers", id), computeClient.BaseHeaders)
	if err != nil {
		return err
	}
	if err2 := resp.JudgeStatus(); err2 != nil {
		return err2
	}
	return nil
}
func (computeClient ComputeClientV2) getBlockDeviceMappingV2(imageRef string) BlockDeviceMappingV2 {
	return BlockDeviceMappingV2{
		BootIndex:          0,
		UUID:               imageRef,
		VolumeSize:         10,
		SourceType:         "image",
		DestinationType:    "volume",
		DeleteOnTemination: true,
	}
}
func (computeClient ComputeClientV2) ServerCreate(options ServerOpt) (Server, error) {
	if options.Name == "" {
		options.Name = fmt.Sprintf("ecTools-server-%s", time.Now().Format("2006-01-02-15:04:05"))
	}
	body, _ := json.Marshal(ServeCreaterBody{Server: options})
	var url string
	if options.BlockDeviceMappingV2 != nil {
		url = computeClient.getUrl("os-volumes_boot", "")
	} else {
		url = computeClient.getUrl("servers", "")
	}
	resp, _ := computeClient.AuthClient.Post(url, body, computeClient.BaseHeaders)
	serverBody := ServerBody{}
	json.Unmarshal(resp.Body, &serverBody)
	return serverBody.Server, resp.JudgeStatus()
}
func (client ComputeClientV2) WaitServerCreate(options ServerOpt) (Server, error) {
	server, err := client.ServerCreate(options)
	if err != nil {
		return server, err
	}
	if server.Id == "" {
		return server, fmt.Errorf("create server failed")
	}
	return client.WaitServerStatusSecond(server.Id, "ACTIVE", 5)
}

func (client ComputeClientV2) WaitServerStatusSecond(serverId string, status string, second int) (Server, error) {
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

func (client ComputeClientV2) WaitServerStatus(serverId string, status string) (Server, error) {
	return client.WaitServerStatusSecond(serverId, status, 1)
}

func (client ComputeClientV2) WaitServerDeleted(id string) {
	client.ServerDelete(id)
	for {
		server, err := client.ServerShow(id)
		if server.Id == "" || err != nil {
			break
		}
		logging.Info("servers status is %s", server.Status)
		time.Sleep(time.Second * 2)
	}
}
