package compute

import (
	"encoding/json"
	"fmt"
	"net/url"
	netUrl "net/url"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/utility"
)

func (client ComputeClientV2) newRequest(resource string, id string, query url.Values, body []byte) common.RestfulRequest {
	return common.RestfulRequest{
		Endpoint: client.endpoint,
		Resource: resource, Id: id,
		Query:   query,
		Body:    body,
		Headers: client.BaseHeaders}
}
func (client ComputeClientV2) newGetRequest(resource string, id string) common.RestfulRequest {
	req := client.newRequest(resource, id, nil, nil)
	req.Method = "GET"
	return req
}
func (client ComputeClientV2) newPostRequest(resource string, body []byte) common.RestfulRequest {
	req := client.newRequest(resource, "", nil, body)
	req.Method = "POST"
	return req
}
func (client ComputeClientV2) newDeleteRequest(resource string, id string) common.RestfulRequest {
	req := client.newRequest(resource, id, nil, nil)
	req.Method = "DELETE"
	return req
}
func (client ComputeClientV2) newPutRequest(resource string, id string, body []byte) common.RestfulRequest {
	req := client.newRequest(resource, id, nil, body)
	req.Method = "PUT"
	return req
}

func (client ComputeClientV2) ServerList(query netUrl.Values) ([]Server, error) {
	resp, err := client.Request(client.newRequest("servers", "", query, nil))
	if err != nil {
		return nil, err
	}
	body := map[string][]Server{"servers": {}}
	resp.BodyUnmarshal(&body)
	return body["servers"], nil
}
func (client ComputeClientV2) ServerListDetails(query netUrl.Values) ([]Server, error) {
	resp, err := client.Request(client.newRequest("servers/detail", "", query, nil))
	if err != nil {
		return nil, err
	}
	body := map[string][]Server{"servers": {}}
	resp.BodyUnmarshal(&body)
	return body["servers"], nil
}
func (client ComputeClientV2) ServerListByName(name string) ([]Server, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.ServerList(query)
}
func (client ComputeClientV2) ServerListDetailsByName(name string) ([]Server, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.ServerListDetails(query)
}
func (client ComputeClientV2) ServerShow(id string) (*Server, error) {
	resp, err := client.Request(client.newRequest("servers", id, nil, nil))
	if err != nil {
		return nil, err
	}
	body := map[string]*Server{"server": {}}
	resp.BodyUnmarshal(&body)
	return body["server"], err
}
func (client ComputeClientV2) ServerFound(idOrName string) (*Server, error) {
	var server *Server
	var err error
	server, err = client.ServerShow(idOrName)
	if err == nil {
		return server, nil
	}
	if httpError, ok := err.(*utility.HttpError); ok {
		if httpError.IsNotFound() {
			var servers []Server
			servers, err = client.ServerListByName(idOrName)
			if err != nil || len(servers) == 0 {
				return nil, fmt.Errorf("server %s not found", idOrName)
			}
			server, err = client.ServerShow(servers[0].Id)
		}
	}
	return server, err
}

func (client ComputeClientV2) ServerDelete(id string) error {
	_, err := client.Request(client.newDeleteRequest("servers", id))
	return err
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

	if options.Networks == nil {
		options.Networks = "none"
	}
	repBody, _ := json.Marshal(map[string]ServerOpt{"server": options})

	var req common.RestfulRequest
	if options.BlockDeviceMappingV2 != nil {
		req = client.newPostRequest("os-volumes_boot", repBody)
	} else {
		req = client.newPostRequest("os-volumes_boot", repBody)
	}
	resp, err := client.Request(req)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Server{"server": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["server"], err
}
func (client ComputeClientV2) ServerCreateAndWait(options ServerOpt) (*Server, error) {
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
		logging.Info("server %s status: %s, taskState: %s", server.Id, server.Status, server.TaskState)
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
	for {
		server, err := client.ServerShow(id)
		if err != nil || server.Id == "" {
			break
		}
		logging.Debug("servers status is %s", server.Status)
		time.Sleep(time.Second * 2)
	}
}

// service api
func (client ComputeClientV2) ServiceList(query netUrl.Values) ([]Service, error) {
	resp, err := client.Request(
		common.NewResourceListRequest(client.endpoint, "os-services", query, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]Service{"services": {}}
	resp.BodyUnmarshal(&body)
	return body["services"], nil
}
func (client ComputeClientV2) ServiceAction(action string, host string, binary string) (*Service, error) {
	reqBody, _ := json.Marshal(Service{Binary: binary, Host: host})
	resp, err := client.Request(client.newPutRequest("os-services/"+action, "", reqBody))
	if err != nil {
		return nil, err
	}
	body := map[string]*Service{"service": {}}
	resp.BodyUnmarshal(&body)
	return body["service"], nil
}
func (client ComputeClientV2) ServiceUpdate(id string, update map[string]interface{}) (*Service, error) {
	reqBody, _ := json.Marshal(update)
	resp, err := client.Request(client.newPutRequest("os-services", id, reqBody))
	if err != nil {
		return nil, err
	}
	body := map[string]*Service{"service": {}}
	resp.BodyUnmarshal(&body)
	return body["service"], nil
}
func (client ComputeClientV2) ServiceGetByHostBinary(host string, binary string) (*Service, error) {
	query := netUrl.Values{"host": []string{host}, "binary": []string{binary}}
	services, err := client.ServiceList(query)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("service %s:%s not found", host, binary)
	}
	return &services[0], nil
}
func (client ComputeClientV2) ServiceUp(host string, binary string) (*Service, error) {
	if client.MicroVersionLargeEqual("2.53") {
		service, err := client.ServiceGetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return client.ServiceUpdate(service.Id,
			map[string]interface{}{"forced_down": false})
	}
	return client.ServiceAction("up", host, binary)
}
func (client ComputeClientV2) ServiceDown(host string, binary string) (*Service, error) {
	if client.MicroVersionLargeEqual("2.53") {
		service, err := client.ServiceGetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return client.ServiceUpdate(service.Id, map[string]interface{}{"forced_down": true})
	}
	return client.ServiceAction("down", host, binary)
}
func (client ComputeClientV2) ServiceEnable(host string, binary string) (*Service, error) {
	if client.MicroVersionLargeEqual("2.53") {
		service, err := client.ServiceGetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return client.ServiceUpdate(service.Id, map[string]interface{}{"status": "enabled"})
	}
	return client.ServiceAction("enable", host, binary)
}
func (client ComputeClientV2) ServiceDisable(host string, binary string,
	reason string,
) (*Service, error) {
	if client.MicroVersionLargeEqual("2.53") {
		service, err := client.ServiceGetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		body := map[string]interface{}{"status": "disabled"}
		if reason != "" {
			body["disabled_reason"] = reason
		}
		return client.ServiceUpdate(service.Id, body)
	}
	return client.ServiceAction("disable", host, binary)
}
func (client ComputeClientV2) ServiceDelete(host string, binary string) error {
	service, err := client.ServiceGetByHostBinary(host, binary)
	if err != nil {
		return err
	}
	_, err = client.Request(client.newDeleteRequest("os-services", service.Id))
	return err
}

// server api
func (client ComputeClientV2) ServerAction(action string, id string, params interface{}) (*utility.Response, error) {
	body, _ := json.Marshal(map[string]interface{}{action: params})
	return client.Request(
		client.newPostRequest(fmt.Sprintf("servers/%s/action", id), body),
	)
}
func (client ComputeClientV2) ServerStop(id string) error {
	_, err := client.ServerAction("os-stop", id, nil)
	return err
}
func (client ComputeClientV2) ServerStart(id string) error {
	_, err := client.ServerAction("os-start", id, nil)
	return err
}
func (client ComputeClientV2) ServerReboot(id string, hard bool) error {
	actionBody := map[string]string{}
	if hard {
		actionBody["type"] = "HARD"
	} else {
		actionBody["type"] = "SOFT"
	}
	_, err := client.ServerAction("reboot", id, actionBody)
	return err
}
func (client ComputeClientV2) ServerPause(id string) error {
	_, err := client.ServerAction("pause", id, nil)
	return err
}
func (client ComputeClientV2) ServerUnpause(id string) error {
	_, err := client.ServerAction("unpause", id, nil)
	return err
}
func (client ComputeClientV2) ServerShelve(id string) error {
	_, err := client.ServerAction("shelve", id, nil)
	return err
}
func (client ComputeClientV2) ServerUnshelve(id string) error {
	_, err := client.ServerAction("unshelve", id, nil)
	return err
}
func (client ComputeClientV2) ServerSuspend(id string) error {
	_, err := client.ServerAction("suspend", id, nil)
	return err
}
func (client ComputeClientV2) ServerResume(id string) error {
	_, err := client.ServerAction("resume", id, nil)
	return err
}
func (client ComputeClientV2) ServerResize(id string, flavorRef string) error {
	data := map[string]interface{}{
		"flavorRef": flavorRef,
	}
	_, err := client.ServerAction("resize", id, data)
	return err
}
func (client ComputeClientV2) ServerMigrate(id string) error {
	_, err := client.ServerAction("migrate", id, nil)
	return err
}
func (client ComputeClientV2) ServerMigrateTo(id string, host string) error {
	server, err := client.ServerShow(id)
	if err != nil {
		return nil
	}
	if server.Host == host {
		return fmt.Errorf("server %s host is %s, skip to migrate", id, server.Host)
	}
	data := map[string]interface{}{"host": host}
	_, err = client.ServerAction("migrate", id, data)
	return err
}

func (client ComputeClientV2) ServerLiveMigrate(id string, blockMigrate bool) error {
	data := map[string]interface{}{"block_migration": blockMigrate, "host": nil}
	_, err := client.ServerAction("os-migrateLive", id, data)
	return err
}
func (client ComputeClientV2) ServerLiveMigrateTo(id string, blockMigrate bool, host string) error {
	data := map[string]interface{}{"block_migration": blockMigrate, "host": host}
	server, err := client.ServerShow(id)
	if err != nil {
		return nil
	}
	if server.Host == host {
		return fmt.Errorf("server %s host is %s, skip to migrate", id, server.Host)
	}
	_, err = client.ServerAction("os-migrateLive", id, data)
	return err
}

// TODO: more params
func (client ComputeClientV2) ServerRebuild(id string) error {
	data := map[string]interface{}{}
	_, err := client.ServerAction("rebuild", id, data)
	return err
}

func (client ComputeClientV2) ServerEvacuate(id string, password string, host string, force bool) error {
	data := map[string]interface{}{}
	if password != "" {
		data["password"] = password
	}
	if host != "" {
		data["host"] = password
	}
	if force {
		data["force"] = force
	}
	_, err := client.ServerAction("evacuate", id, data)
	return err
}
func (client ComputeClientV2) ServerConsoleLog(id string, length uint) (*ConsoleLog, error) {
	params := map[string]interface{}{}
	if length != 0 {
		params["length"] = length
	}
	resp, err := client.ServerAction("os-getConsoleOutput", id, params)
	if err != nil {
		return nil, err
	}
	body := ConsoleLog{}
	resp.BodyUnmarshal(&body)
	return &body, nil
}
func (client ComputeClientV2) getVNCConsole(id string, consoleType string) (*Console, error) {
	params := map[string]interface{}{"type": consoleType}
	resp, err := client.ServerAction("os-getVNCConsole", id, params)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Console{"console": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["console"], nil
}
func (client ComputeClientV2) getRemoteConsole(id string, protocol string, consoleType string) (*Console, error) {
	params := map[string]interface{}{
		"remote_console": map[string]interface{}{
			"protocol": protocol,
			"type":     consoleType,
		},
	}
	reqBody, _ := json.Marshal(params)
	resp, err := client.Request(
		client.newPostRequest(fmt.Sprintf("servers/%s/remote-consoles", id), reqBody),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Console{"console": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["remote_console"], nil
}

func (client ComputeClientV2) ServerConsoleUrl(id string, consoleType string) (*Console, error) {
	if client.MicroVersionLargeEqual("2.6") {
		// TODO: do not set "vnc" directly
		return client.getRemoteConsole(id, "vnc", consoleType)
	}
	return client.getVNCConsole(id, consoleType)
}

// server action api
func (client ComputeClientV2) ServerActionList(id string) ([]InstanceAction, error) {
	resp, err := client.Request(
		client.newRequest(fmt.Sprintf("servers/%s/os-instance-actions", id), "", nil, nil),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]InstanceAction{"instanceActions": {}}
	resp.BodyUnmarshal(&body)
	return body["instanceActions"], nil
}
func (client ComputeClientV2) ServerActionShow(id string, requestId string) (
	*InstanceAction, error,
) {
	resp, err := client.Request(
		client.newGetRequest(
			utility.UrlJoin("servers", id, "os-instance-actions"), requestId,
		),
	)
	if err != nil {
		return nil, err
	}
	body := map[string]*InstanceAction{"instanceAction": {}}
	resp.BodyUnmarshal(&body)
	return body["instanceAction"], nil
}
func (client ComputeClientV2) serverRegionLiveMigrate(id string, destRegion string, blockMigrate bool, dryRun bool, destHost string) (*RegionMigrateResp, error) {
	data := map[string]interface{}{
		"region":          destRegion,
		"block_migration": blockMigrate,
		"dry_run":         dryRun,
	}
	if destHost != "" {
		data["host"] = destHost
	}
	resp, err := client.ServerAction("os-migrateLive-region", id, data)
	if err != nil {
		return nil, err
	}
	respBody := &RegionMigrateResp{}
	resp.BodyUnmarshal(respBody)
	return respBody, nil
}
func (client ComputeClientV2) ServerRegionLiveMigrate(id string, destRegion string, dryRun bool) (*RegionMigrateResp, error) {
	return client.serverRegionLiveMigrate(id, destRegion, true, dryRun, "")
}
func (client ComputeClientV2) ServerRegionLiveMigrateTo(id string, destRegion string, dryRun bool, destHost string) (*RegionMigrateResp, error) {
	return client.serverRegionLiveMigrate(id, destRegion, true, dryRun, destHost)
}
func (client ComputeClientV2) ServerMigrationList(id string, query url.Values) ([]Migration, error) {
	resp, err := client.Request(
		client.newGetRequest(utility.UrlJoin("servers", id, "migrations"), ""),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string][]Migration{"migrations": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["migrations"], nil
}

// flavor api
func (client ComputeClientV2) FlavorList(query netUrl.Values) ([]Flavor, error) {
	resp, err := client.Request(client.newRequest("flavors", "", query, nil))
	if err != nil {
		return nil, err
	}
	body := map[string][]Flavor{"flavors": {}}
	resp.BodyUnmarshal(body)
	return body["flavors"], nil
}

func (client ComputeClientV2) FlavorCreate(flavor Flavor) (*Flavor, error) {
	reqBody, _ := json.Marshal(map[string]Flavor{"flavor": flavor})
	resp, err := client.Request(client.newPostRequest("flavors", reqBody))
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Flavor{"flavor": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["flavor"], nil
}
func (client ComputeClientV2) FlavorListDetail(query netUrl.Values) (Flavors, error) {
	resp, err := client.Request(client.newRequest("flavors/detail", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string]Flavors{"flavors": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["flavors"], nil
}
func (client ComputeClientV2) FlavorExtraSpecsList(flavorId string) (ExtraSpecs, error) {
	resp, err := client.Request(
		client.newGetRequest(utility.UrlJoin("flavors", flavorId, "os-extra_specs"), ""),
	)
	if err != nil {
		return nil, err
	}
	body := map[string]ExtraSpecs{"extra_specs": {}}
	resp.BodyUnmarshal(&body)
	return body["extra_specs"], nil
}
func (client ComputeClientV2) FlavorShow(flavorId string) (*Flavor, error) {
	resp, err := client.Request(client.newGetRequest("flavors", flavorId))
	if err != nil {
		return nil, err
	}
	body := map[string]*Flavor{"flavor": {}}
	resp.BodyUnmarshal(&body)
	return body["flavor"], nil
}

func (client ComputeClientV2) FlavorShowWithExtraSpecs(flavorId string) (*Flavor, error) {
	flavor, err := client.FlavorFound(flavorId)
	if err != nil {
		return nil, err
	}
	extraSpecs, err := client.FlavorExtraSpecsList(flavor.Id)
	if err != nil {
		return nil, err
	}
	flavor.ExtraSpecs = extraSpecs
	return flavor, nil
}
func (client ComputeClientV2) FlavorExtraSpecsCreate(flavorId string, extraSpecs map[string]string) (ExtraSpecs, error) {
	reqBody, _ := json.Marshal(map[string]ExtraSpecs{"extra_specs": extraSpecs})
	resp, err := client.Request(
		client.newPostRequest(utility.UrlJoin("flavors", flavorId, "os-extra_specs"), reqBody),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]ExtraSpecs{"extra_specs": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["extra_specs"], nil
}
func (client ComputeClientV2) FlavorExtraSpecsDelete(flavorId string, extraSpec string) error {
	_, err := client.Request(
		client.newDeleteRequest(utility.UrlJoin("flavors", flavorId, "os-extra_specs"), extraSpec),
	)
	return err
}
func (client ComputeClientV2) FlavorDelete(flavorId string) error {
	_, err := client.Request(client.newDeleteRequest("flavors", flavorId))
	return err
}
func (client ComputeClientV2) FlavorFoundByName(name string) (*Flavor, error) {
	flavors, err := client.FlavorList(nil)
	if err != nil {
		return nil, err
	}
	for _, flavor := range flavors {
		if flavor.Name != name {
			continue
		}
		flavor, err := client.FlavorShow(flavor.Id)
		if err != nil {
			return nil, err
		} else {
			return flavor, nil
		}
	}
	return nil, fmt.Errorf("flavor %s not found", name)
}
func (client ComputeClientV2) FlavorFound(idOrName string) (*Flavor, error) {
	flavor, err := client.FlavorShow(idOrName)
	if err == nil {
		return flavor, nil
	}
	if httpError, ok := err.(*utility.HttpError); ok {
		if !httpError.IsNotFound() {
			return nil, err
		}
	}
	return client.FlavorFoundByName(idOrName)
}

// hypervisor api
func (client ComputeClientV2) HypervisorList(query netUrl.Values) (Hypervisors, error) {
	body := map[string]Hypervisors{"hypervisors": []Hypervisor{}}
	resp, err := client.Request(client.newRequest("os-hypervisors", "", query, nil))
	if err != nil {
		return nil, err
	}
	resp.BodyUnmarshal(&body)
	return body["hypervisors"], nil
}
func (client ComputeClientV2) HypervisorListDetail(query netUrl.Values) (Hypervisors, error) {
	body := map[string]Hypervisors{"hypervisors": {}}
	resp, err := client.Request(client.newRequest("os-hypervisors/detail", "", query, nil))
	if err != nil {
		return nil, err
	}
	resp.BodyUnmarshal(&body)
	return body["hypervisors"], nil
}
func (client ComputeClientV2) HypervisorShow(id string) (*Hypervisor, error) {
	resp, err := client.Request(client.newGetRequest("os-hypervisors", id))
	if err != nil {
		return nil, err
	}
	body := map[string]*Hypervisor{"hypervisor": {}}
	resp.BodyUnmarshal(&body)
	return body["hypervisor"], nil
}
func (client ComputeClientV2) HypervisorShowByHostname(hostname string) (*Hypervisor, error) {
	query := url.Values{"hypervisor_hostname_pattern": []string{hostname}}
	hypervisors, err := client.HypervisorList(query)
	if err != nil {
		return nil, err
	}
	if len(hypervisors) == 0 {
		return nil, &utility.HttpError{
			Status: 404, Reason: "NotFound",
			Message: fmt.Sprintf("hypervisor named %s not found", hostname),
		}
	}
	hypervisor, err := client.HypervisorShow(hypervisors[0].Id)
	if err != nil {
		return nil, err
	}
	return hypervisor, nil
}
func (client ComputeClientV2) HypervisorFound(idOrName string) (*Hypervisor, error) {
	if !utility.IsUUID(idOrName) {
		return client.HypervisorShowByHostname(idOrName)
	}
	hypervisor, err := client.HypervisorShow(idOrName)
	if httpError, ok := err.(*utility.HttpError); ok {
		if httpError.Status == 404 {
			hypervisor, err = client.HypervisorShowByHostname(idOrName)
		}
	}
	return hypervisor, err
}
func (client ComputeClientV2) HypervisorUptime(id string) (*Hypervisor, error) {

	resp, err := client.Request(
		common.NewResourceListRequest(
			client.endpoint, utility.UrlJoin("os-hypervisors", id, "uptime"),
			nil, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	body := map[string]*Hypervisor{"hypervisor": {}}
	resp.BodyUnmarshal(&body)
	return body["hypervisor"], nil
}

// keypair api
func (client ComputeClientV2) KeypairList(query netUrl.Values) ([]Keypair, error) {
	body := map[string][]Keypair{"keypairs": {}}
	resp, err := client.Request(client.newRequest("os-keypairs", "", query, nil))
	if err != nil {
		return nil, err
	}
	resp.BodyUnmarshal(&body)
	return body["keypairs"], nil
}

// server volumes api

func (client ComputeClientV2) ServerVolumeList(id string) ([]VolumeAttachment, error) {
	body := map[string][]VolumeAttachment{"volumeAttachments": {}}
	resp, err := client.Request(
		client.newGetRequest(utility.UrlJoin("servers", id, "os-volume_attachments"), ""),
	)
	if err != nil {
		return nil, err
	}
	resp.BodyUnmarshal(&body)
	return body["volumeAttachments"], nil
}
func (client ComputeClientV2) ServerVolumeAdd(id string, volumeId string) (*VolumeAttachment, error) {
	reqBody, _ := json.Marshal(
		map[string]map[string]string{"volumeAttachment": {"volumeId": volumeId}},
	)
	respBody := map[string]*VolumeAttachment{"volumeAttachment": {}}
	resp, err := client.Request(
		client.newPostRequest(utility.UrlJoin("servers", id, "os-volume_attachments"), reqBody),
	)
	if err != nil {
		return nil, err
	}
	resp.BodyUnmarshal(&respBody)
	return respBody["volumeAttachment"], nil
}
func (client ComputeClientV2) ServerVolumeDelete(id string, volumeId string) error {
	_, err := client.Request(
		client.newDeleteRequest(utility.UrlJoin("servers", id, "os-volume_attachments"), volumeId),
	)
	return err
}
func (client ComputeClientV2) ServerInterfaceList(id string) ([]InterfaceAttachment, error) {
	resp, err := client.Request(
		client.newGetRequest(utility.UrlJoin("servers", id, "os-interface"), ""),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]InterfaceAttachment{"interfaceAttachments": {}}
	resp.BodyUnmarshal(&body)
	return body["interfaceAttachments"], nil
}

func (client ComputeClientV2) ServerAddNet(id string, netId string) (*InterfaceAttachment, error) {
	data := map[string]string{"net_id": netId}
	reqBody, _ := json.Marshal(map[string]interface{}{"interfaceAttachment": data})
	resp, err := client.Request(
		client.newPostRequest(utility.UrlJoin("servers", id, "os-interface"), reqBody),
	)
	if err != nil {
		return nil, err
	}
	body := map[string]*InterfaceAttachment{"interfaceAttachment": {}}
	resp.BodyUnmarshal(&body)
	return body["interfaceAttachment"], nil
}
func (client ComputeClientV2) ServerAddPort(id string, portId string) (*InterfaceAttachment, error) {
	data := map[string]string{"port_id": portId}
	reqBody, _ := json.Marshal(map[string]interface{}{"interfaceAttachment": data})
	resp, err := client.Request(
		client.newPostRequest(utility.UrlJoin("servers", id, "os-interface"), reqBody),
	)
	if err != nil {
		return nil, err
	}
	body := map[string]*InterfaceAttachment{"interfaceAttachment": {}}
	resp.BodyUnmarshal(&body)
	return body["interfaceAttachment"], nil
}

func (client ComputeClientV2) ServerInterfaceDetach(id string, portId string) error {
	_, err := client.Request(
		client.newDeleteRequest(utility.UrlJoin("servers", id, "os-interface"), portId),
	)
	return err
}
func (client ComputeClientV2) ServerSetPassword(id string, password string, user string) error {
	data := map[string]interface{}{
		"adminPass": password,
	}
	if user != "" {
		data["userName"] = user
	}
	_, err := client.ServerAction("changePassword", id, data)
	return err
}
func (client ComputeClientV2) ServerSetName(id string, name string) error {
	data := map[string]interface{}{"name": name}
	_, err := client.ServerAction("rename", id, data)
	return err
}

// migration api

func (client ComputeClientV2) MigrationList(query netUrl.Values) ([]Migration, error) {
	resp, err := client.Request(client.newRequest("os-migrations", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]Migration{"migrations": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["migrations"], nil
}

// availability zone api
func (client ComputeClientV2) AZList(query netUrl.Values) (*AvailabilityZone, error) {
	resp, err := client.Request(client.newRequest("os-availability-zone", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string]*AvailabilityZone{"availabilityZoneInfo": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["availabilityZoneInfo"], nil
}
func (client ComputeClientV2) AZListDetail(query netUrl.Values) ([]AvailabilityZone, error) {
	resp, err := client.Request(client.newRequest("os-availability-zone/detail", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]AvailabilityZone{"availabilityZoneInfo": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["availabilityZoneInfo"], nil
}
func (client ComputeClientV2) AggregateList(query netUrl.Values) ([]Aggregate, error) {
	resp, err := client.Request(client.newRequest("os-aggregates", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]Aggregate{"aggregates": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["aggregates"], nil
}
func (client ComputeClientV2) AggregateShow(aggregate string) (*Aggregate, error) {
	resp, err := client.Request(client.newGetRequest("aggregates", aggregate))
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Aggregate{"aggregate": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["aggregate"], nil
}

func (client ComputeClientV2) ServerGroupList(query netUrl.Values) ([]ServerGroup, error) {
	resp, err := client.Request(client.newRequest("os-server-groups", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]ServerGroup{"server_groups": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["server_groups"], nil
}
