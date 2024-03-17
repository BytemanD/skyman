package openstack

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

const V2_1 = "v2.1"

type microVersion struct {
	Version      int
	MicroVersion int
}

func getMicroVersion(vertionStr string) microVersion {
	versionList := strings.Split(vertionStr, ".")
	v, _ := strconv.Atoi(versionList[0])
	micro, _ := strconv.Atoi(versionList[1])
	return microVersion{Version: v, MicroVersion: micro}
}

type NovaV2 struct {
	RestClient
	MicroVersion model.ApiVersion
}
type ServersApi struct {
	NovaV2
}
type FlavorApi struct {
	NovaV2
}
type ComputeServiceApi struct {
	NovaV2
}
type HypervisorApi struct {
	NovaV2
}
type KeypairApi struct {
	NovaV2
}
type MigrationApi struct {
	NovaV2
}
type AZApi struct {
	NovaV2
}
type ServerGroupApi struct {
	NovaV2
}
type AggregateApi struct {
	NovaV2
}

func (c NovaV2) Servers() ServersApi {
	return ServersApi{c}
}
func (c NovaV2) Flavors() FlavorApi {
	return FlavorApi{c}
}
func (c NovaV2) Services() ComputeServiceApi {
	return ComputeServiceApi{c}
}
func (c NovaV2) Hypervisors() HypervisorApi {
	return HypervisorApi{c}
}
func (c NovaV2) Keypairs() KeypairApi {
	return KeypairApi{c}
}
func (c NovaV2) Migrations() MigrationApi {
	return MigrationApi{c}
}
func (c NovaV2) AvailabilityZones() AZApi {
	return AZApi{c}
}
func (c NovaV2) ServerGroups() ServerGroupApi {
	return ServerGroupApi{c}
}
func (c NovaV2) Aggregates() AggregateApi {
	return AggregateApi{c}
}
func (o Openstack) NovaV2() *NovaV2 {
	if o.novaClient == nil {
		endpoint, err := o.AuthPlugin.GetServiceEndpoint("compute", "nova", "public")
		if err != nil {
			logging.Fatal("get compute endpoint falied: %v", err)
		}
		o.novaClient = &NovaV2{
			RestClient: RestClient{
				BaseUrl:    utility.VersionUrl(endpoint, V2_1),
				AuthPlugin: o.AuthPlugin,
				Headers:    map[string]string{},
			},
		}
		currentVersion, err := o.novaClient.GetCurrentVersion()
		if err != nil {
			logging.Warning("get current failed: %v", err)
		}
		o.novaClient.MicroVersion = *currentVersion
		o.novaClient.RestClient.Headers["Openstack-Api-Version"] = o.novaClient.MicroVersion.Version
		o.novaClient.RestClient.Headers["X-Openstack-Nova-Api-Version"] = o.novaClient.MicroVersion.Version
		logging.Debug("current nova version: %s", o.novaClient.MicroVersion)
	}
	return o.novaClient
}

func (c *NovaV2) MicroVersionLargeEqual(version string) bool {
	clientVersion := getMicroVersion(c.MicroVersion.Version)
	otherVersion := getMicroVersion(version)
	if clientVersion.Version > otherVersion.Version {
		return true
	} else if clientVersion.Version == otherVersion.Version {
		return clientVersion.MicroVersion >= otherVersion.MicroVersion
	} else {
		return false
	}
}
func (c NovaV2) GetCurrentVersion() (*model.ApiVersion, error) {
	resp, err := c.Index()
	if err != nil {
		return nil, err
	}
	apiVersions := struct{ Versions []model.ApiVersion }{}
	if err := resp.BodyUnmarshal(&apiVersions); err != nil {
		return nil, err
	}
	for _, version := range apiVersions.Versions {
		if version.Status == "CURRENT" {
			return &version, nil
		}
	}
	return nil, fmt.Errorf("current version not found")
}

func (c *NovaV2) String() string {
	return fmt.Sprintf("<%s %s>", c.BaseUrl, c.AuthPlugin.Region())
}

// server api

func (c ServersApi) List(query url.Values) ([]nova.Server, error) {
	resp, err := c.NovaV2.Get("servers", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.Server{"servers": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["servers"], nil
}
func (c ServersApi) ListByName(name string) ([]nova.Server, error) {
	query := url.Values{}
	query.Set("name", name)
	return c.List(query)
}
func (c ServersApi) Detail(query url.Values) ([]nova.Server, error) {
	resp, err := c.NovaV2.Get("servers/detail", query)
	if err != nil {
		return nil, err
	}
	body := struct{ Servers []nova.Server }{}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body.Servers, nil
}
func (c ServersApi) DetailsByName(name string) ([]nova.Server, error) {
	return c.Detail(utility.UrlValues(map[string]string{"name": name}))
}

func (c ServersApi) Show(id string) (*nova.Server, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("servers", id), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*nova.Server{"server": {}}
	resp.BodyUnmarshal(&body)
	return body["server"], err
}
func (c ServersApi) Found(idOrName string) (*nova.Server, error) {
	var (
		server *nova.Server
		err    error
	)
	server, err = c.Show(idOrName)
	if err == nil {
		return server, nil
	}
	if httpError, ok := err.(*utility.HttpError); ok {
		if httpError.IsNotFound() {
			var servers []nova.Server
			servers, err = c.Servers().ListByName(idOrName)
			if err != nil || len(servers) == 0 {
				return nil, fmt.Errorf("server %s not found", idOrName)
			}
			server, err = c.Servers().Show(servers[0].Id)
		}
	}
	return server, err
}
func (c ServersApi) Create(options nova.ServerOpt) (*nova.Server, error) {
	if options.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if options.Flavor == "" {
		return nil, fmt.Errorf("flavor is empty")
	}
	if options.Networks == nil {
		options.Networks = "none"
	}
	repBody, _ := json.Marshal(map[string]nova.ServerOpt{"server": options})
	var (
		resp *utility.Response
		err  error
	)

	if options.BlockDeviceMappingV2 != nil {
		resp, err = c.NovaV2.Post("os-volumes_boot", repBody, nil)
	} else {
		resp, err = c.NovaV2.Post("servers", repBody, nil)
	}
	if err != nil {
		return nil, err
	}
	respBody := map[string]*nova.Server{"server": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["server"], err
}

func (c ServersApi) Delete(id string) (err error) {
	_, err = c.NovaV2.Delete(utility.UrlJoin("servers", id), nil)
	return err
}
func (c ServersApi) ListVolumes(id string) ([]nova.VolumeAttachment, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("servers", id, "os-volume_attachments"), nil)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.VolumeAttachment{"volumeAttachments": {}}
	resp.BodyUnmarshal(&body)
	return body["volumeAttachments"], nil
}
func (c ServersApi) AddVolume(id string, volumeId string) (*nova.VolumeAttachment, error) {
	reqBody, _ := json.Marshal(
		map[string]map[string]string{"volumeAttachment": {"volumeId": volumeId}},
	)
	resp, err := c.NovaV2.Post(
		utility.UrlJoin("servers", id, "os-volume_attachments"), reqBody, nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*nova.VolumeAttachment{"volumeAttachments": {}}
	resp.BodyUnmarshal(&body)
	return body["volumeAttachments"], nil
}
func (c ServersApi) DeleteVolume(id string, volumeId string) error {
	_, err := c.NovaV2.Delete(utility.UrlJoin("servers", id, "os-volume_attachments", volumeId), nil)
	return err
}
func (c ServersApi) ListInterfaces(id string) ([]nova.InterfaceAttachment, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("servers", id, "os-interface"), nil)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.InterfaceAttachment{"interfaceAttachments": {}}
	resp.BodyUnmarshal(&body)
	return body["interfaceAttachments"], nil
}
func (c ServersApi) AddInterface(id string, netId, portId string) (*nova.InterfaceAttachment, error) {
	params := map[string]string{}
	if netId == "" && portId == "" {
		return nil, errors.New("invalid params: portId or netId is required")
	}
	if netId != "" {
		params["net_id"] = portId
	}
	if portId != "" {
		params["port_id"] = portId
	}
	reqBody, _ := json.Marshal(map[string]map[string]string{"interfaceAttachment": params})
	resp, err := c.NovaV2.Post(utility.UrlJoin("servers", id, "os-interface"), reqBody, nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*nova.InterfaceAttachment{"interfaceAttachment": {}}
	resp.BodyUnmarshal(&body)
	return body["interfaceAttachment"], nil
}
func (c ServersApi) DeleteInterface(id string, portId string) error {
	_, err := c.NovaV2.Delete(utility.UrlJoin("servers", id, "os-interface", portId), nil)
	return err
}
func (c ServersApi) doAction(action string, id string, params interface{}) (*utility.Response, error) {
	body, _ := json.Marshal(map[string]interface{}{action: params})
	return c.NovaV2.Post(utility.UrlJoin("servers", id, "action", id), body, nil)
}
func (c ServersApi) Stop(id string) error {
	_, err := c.doAction("os-stop", id, nil)
	return err
}
func (c ServersApi) Start(id string) error {
	_, err := c.doAction("os-start", id, nil)
	return err
}
func (c ServersApi) Reboot(id string, hard bool) error {
	rebootTypes := map[bool]string{}
	body := map[string]string{"type": rebootTypes[hard]}
	_, err := c.doAction("reboot", id, body)
	return err
}
func (c ServersApi) Pause(id string) error {
	_, err := c.doAction("pause", id, nil)
	return err
}
func (c ServersApi) Unpause(id string) error {
	_, err := c.doAction("unpause", id, nil)
	return err
}
func (c ServersApi) Shelve(id string) error {
	_, err := c.doAction("shelve", id, nil)
	return err
}
func (c ServersApi) Unshelve(id string) error {
	_, err := c.doAction("unshelve", id, nil)
	return err
}
func (c ServersApi) Suspend(id string) error {
	_, err := c.doAction("suspend", id, nil)
	return err
}
func (c ServersApi) Resume(id string) error {
	_, err := c.doAction("resume", id, nil)
	return err
}
func (c ServersApi) Resize(id string, flavorRef string) error {
	_, err := c.doAction("resize", id, map[string]string{"flavorRef": flavorRef})
	return err
}

// TODO: more params
func (c ServersApi) Rebuild(id string) error {
	_, err := c.doAction("rebuild", id, map[string]string{})
	return err
}
func (c ServersApi) Evacuate(id string, password string, host string, force bool) error {
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
	_, err := c.doAction("evacuate", id, data)
	return err
}
func (c ServersApi) SetPassword(id string, password, user string) error {
	data := map[string]interface{}{
		"adminPass": password,
	}
	if user != "" {
		data["userName"] = user
	}
	_, err := c.doAction("adminPass", id, data)
	return err
}
func (c ServersApi) SetName(id string, name string) error {
	data := map[string]interface{}{"name": name}
	_, err := c.doAction("adminPass", id, data)
	return err
}
func (c ServersApi) ConsoleLog(id string, length uint) (*nova.ConsoleLog, error) {
	params := map[string]interface{}{}
	if length != 0 {
		params["length"] = length
	}
	resp, err := c.doAction("os-getConsoleOutput", id, params)
	if err != nil {
		return nil, err
	}
	body := nova.ConsoleLog{}
	resp.BodyUnmarshal(&body)
	return &body, nil
}
func (c ServersApi) getVNCConsole(id string, consoleType string) (*nova.Console, error) {
	params := map[string]interface{}{"type": consoleType}
	resp, err := c.doAction("os-getVNCConsole", id, params)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*nova.Console{"console": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["console"], nil
}
func (c ServersApi) getRemoteConsole(id string, protocol string, consoleType string) (*nova.Console, error) {
	params := map[string]interface{}{
		"remote_console": map[string]interface{}{
			"protocol": protocol,
			"type":     consoleType,
		},
	}
	reqBody, _ := json.Marshal(params)
	resp, err := c.NovaV2.Post(utility.UrlJoin("servers", id, "remote-consoles"), reqBody, nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*nova.Console{"console": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["remote_console"], nil
}

func (c ServersApi) ConsoleUrl(id string, consoleType string) (*nova.Console, error) {
	if c.MicroVersionLargeEqual("2.6") {
		// TODO: do not set "vnc" directly
		return c.getRemoteConsole(id, "vnc", consoleType)
	}
	return c.getVNCConsole(id, consoleType)
}
func (c ServersApi) Migrate(id string, host string) error {
	data := map[string]interface{}{}
	if host != "" {
		server, err := c.Show(id)
		if err != nil {
			return nil
		}
		if server.Host == host {
			return fmt.Errorf("server %s host is %s", id, server.Host)
		}
		data["host"] = host
	}
	_, err := c.doAction("migrate", id, data)
	return err
}
func (c ServersApi) LiveMigrate(id string, blockMigrate bool, host string) error {
	data := map[string]interface{}{"block_migration": blockMigrate}
	if host != "" {
		server, err := c.Show(id)
		if err != nil {
			return nil
		}
		if server.Host == host {
			return fmt.Errorf("server %s host is %s", id, server.Host)
		}
		data["host"] = host
	}
	_, err := c.doAction("migrate", id, data)
	return err
}
func (c ServersApi) ListActions(id string) ([]nova.InstanceAction, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("servers", id, "os-instance-actions"), nil)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.InstanceAction{"instanceActions": {}}
	resp.BodyUnmarshal(&body)
	return body["instanceActions"], nil
}
func (c ServersApi) ShowAction(id, requestId string) (*nova.InstanceAction, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("servers", id, "os-instance-actions", requestId), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*nova.InstanceAction{"instanceAction": {}}
	resp.BodyUnmarshal(&body)
	return body["instanceAction"], nil
}
func (c ServersApi) ListActionsWithEvents(id string, actionName string, requestId string) ([]nova.InstanceActionAndEvents, error) {
	serverActions, err := c.ListActions(id)
	if err != nil {
		return nil, err
	}
	var actionEvents []nova.InstanceActionAndEvents
	for _, action := range serverActions {
		if requestId != "" && action.RequestId != requestId {
			continue
		}
		if actionName != "" && action.Action != actionName {
			continue
		}
		serverAction, err := c.ShowAction(id, action.RequestId)
		if err != nil {
			logging.Error("get server action %s failed: %s", action.RequestId, err)
		}
		actionEvents = append(actionEvents, nova.InstanceActionAndEvents{
			InstanceAction: action,
			RequestId:      action.RequestId,
			Events:         serverAction.Events,
		})
	}
	return actionEvents, nil
}

func (c ServersApi) ListMigrations(id string, query url.Values) ([]nova.Migration, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("servers", id, "migrations"), query)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.Migration{"migrations": {}}
	resp.BodyUnmarshal(&body)
	return body["migrations"], nil
}
func (c ServersApi) RegionLiveMigrate(id string, destRegion string, blockMigrate bool, dryRun bool, destHost string) (*nova.RegionMigrateResp, error) {
	data := map[string]interface{}{
		"region":          destRegion,
		"block_migration": blockMigrate,
		"dry_run":         dryRun,
	}
	if destHost != "" {
		data["host"] = destHost
	}
	resp, err := c.doAction("os-migrateLive-region", id, data)
	if err != nil {
		return nil, err
	}
	respBody := &nova.RegionMigrateResp{}
	resp.BodyUnmarshal(respBody)
	return respBody, nil
}

// flavor api

func (c FlavorApi) List(query url.Values) ([]nova.Flavor, error) {
	resp, err := c.NovaV2.Get("flavors", query)
	if err != nil {
		return nil, err
	}
	body := struct{ Flavors []nova.Flavor }{}
	// fmt.Println("flavors: ")
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body.Flavors, nil
}
func (c FlavorApi) Detail(query url.Values) ([]nova.Flavor, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("flavors", "detail"), query)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.Flavor{"flavors": {}}

	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["flavors"], nil
}
func (c FlavorApi) Show(id string) (*nova.Flavor, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("flavors", id), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*nova.Flavor{"flavor": {}}
	resp.BodyUnmarshal(&body)
	return body["flavor"], err
}
func (c FlavorApi) ShowWithExtraSpecs(id string) (*nova.Flavor, error) {
	flavor, err := c.Found(id)
	if err != nil {
		return nil, err
	}
	extraSpecs, err := c.ListExtraSpecs(flavor.Id)
	if err != nil {
		return nil, err
	}
	flavor.ExtraSpecs = extraSpecs
	return flavor, err
}
func (c FlavorApi) SetExtraSpecs(id string, extraSpecs map[string]string) (nova.ExtraSpecs, error) {
	reqBody, _ := json.Marshal(map[string]nova.ExtraSpecs{"extra_specs": extraSpecs})
	resp, err := c.NovaV2.Post(
		utility.UrlJoin("flavors", id, "os-extra_specs"),
		reqBody, nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]nova.ExtraSpecs{"extra_specs": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["extra_specs"], nil
}
func (c FlavorApi) DeleteExtraSpecs(id string, extraSpec string) error {
	_, err := c.NovaV2.Post(utility.UrlJoin("flavors", id, "os-extra_specs", extraSpec), nil, nil)
	return err
}
func (c FlavorApi) Delete(id string) (err error) {
	_, err = c.NovaV2.Delete(utility.UrlJoin("flavors", id), nil)
	return err
}
func (c FlavorApi) Found(idOrName string) (*nova.Flavor, error) {
	var (
		flavor *nova.Flavor
		err    error
	)
	flavor, err = c.Show(idOrName)
	if err == nil {
		return flavor, nil
	}
	if httpError, ok := err.(*utility.HttpError); !ok || !httpError.IsNotFound() {
		return nil, err
	}

	flavors, err := c.List(nil)
	if err != nil || len(flavors) == 0 {
		return nil, fmt.Errorf("flavor %s not found", idOrName)
	}
	for _, flavor := range flavors {
		if flavor.Name != idOrName {
			continue
		}
		flavor, err := c.Show(flavor.Id)
		if err != nil {
			return nil, err
		} else {
			return flavor, nil
		}
	}
	return nil, fmt.Errorf("flavor %s not found", idOrName)
}
func (c FlavorApi) Create(flavor nova.Flavor) (*nova.Flavor, error) {
	reqBody, _ := json.Marshal(map[string]nova.Flavor{"flavor": flavor})
	resp, err := c.NovaV2.Post("flavors", reqBody, nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*nova.Flavor{"flavor": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["flavor"], nil
}
func (c FlavorApi) ListExtraSpecs(id string) (nova.ExtraSpecs, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("flavors", id, "os-extra_specs"), nil)

	if err != nil {
		return nil, err
	}
	body := map[string]nova.ExtraSpecs{"extra_specs": {}}
	resp.BodyUnmarshal(&body)
	return body["extra_specs"], err
}
func (c FlavorApi) Copy(id string, newName string, newId string,
	newVcpus int, newRam int, newDisk int, newSwap int,
	newEphemeral int, newRxtxFactor float32, setProperties map[string]string,
	unsetProperties []string,
) (*nova.Flavor, error) {

	logging.Info("Show flavor")
	flavor, err := c.Show(id)
	if err != nil {
		return nil, err
	}
	flavor.Name = newName
	flavor.Id = newId
	if newVcpus != 0 {
		flavor.Vcpus = newVcpus
	}
	if newRam != 0 {
		flavor.Ram = int(newRam)
	}
	if newDisk != 0 {
		flavor.Disk = newDisk
	}
	if newSwap != 0 {
		flavor.Swap = newSwap
	}
	if newEphemeral != 0 {
		flavor.Ephemeral = newEphemeral
	}
	if newRxtxFactor != 0 {
		flavor.RXTXFactor = newRxtxFactor
	}
	logging.Info("Show flavor extra specs")
	extraSpecs, err := c.ListExtraSpecs(id)
	if err != nil {
		return nil, err
	}
	for k, v := range setProperties {
		extraSpecs[k] = v
	}
	for _, k := range unsetProperties {
		delete(extraSpecs, k)
	}
	logging.Info("Create new flavor")
	newFlavor, err := c.Create(*flavor)
	if err != nil {
		return nil, fmt.Errorf("create flavor failed, %v", err)
	}
	if len(extraSpecs) != 0 {
		logging.Info("Set new flavor extra specs")
		_, err = c.SetExtraSpecs(newFlavor.Id, extraSpecs)
		if err != nil {
			return nil, fmt.Errorf("set flavor extra specs failed, %v", err)
		}
		newFlavor.ExtraSpecs = extraSpecs
	}
	return newFlavor, nil
}

// hypervisor api

func (c HypervisorApi) List(query url.Values) ([]nova.Hypervisor, error) {
	resp, err := c.NovaV2.Get("os-hypervisors", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.Hypervisor{"hypervisors": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["hypervisors"], nil
}
func (c HypervisorApi) Details(query url.Values) ([]nova.Hypervisor, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("os-hypervisors", "detail"), query)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.Hypervisor{"flavors": {}}

	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["hypervisors"], nil
}
func (c HypervisorApi) ListByName(name string) ([]nova.Hypervisor, error) {
	query := url.Values{}
	query.Set("name", name)
	return c.List(query)
}

func (c HypervisorApi) Show(id string) (*nova.Hypervisor, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("os-hypervisors", id), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*nova.Hypervisor{"flavor": {}}
	resp.BodyUnmarshal(&body)
	return body["flavor"], err
}
func (c HypervisorApi) ShowByHostname(hostname string) (*nova.Hypervisor, error) {
	hypervisors, err := c.List(utility.UrlValues(map[string]string{
		"hypervisor_hostname_pattern": hostname,
	}))
	if err != nil {
		return nil, err
	}
	if len(hypervisors) == 0 {
		return nil, &utility.HttpError{
			Status: 404, Reason: "NotFound",
			Message: fmt.Sprintf("hypervisor %s not found", hostname),
		}
	}
	return c.Show(hypervisors[0].Id)
}
func (c HypervisorApi) Found(idOrHostName string) (*nova.Hypervisor, error) {
	if !utility.IsUUID(idOrHostName) {
		return c.ShowByHostname(idOrHostName)
	}
	hypervisors, err := c.ListByName(idOrHostName)
	if err != nil {
		return nil, err
	}
	if len(hypervisors) == 0 {
		return nil, fmt.Errorf("hypervisor %s not found", idOrHostName)
	}
	return c.Show(hypervisors[0].Id)
}
func (c HypervisorApi) Delete(id string) (err error) {
	_, err = c.NovaV2.Delete(utility.UrlJoin("os-hypervisors", id), nil)
	return err
}
func (c HypervisorApi) Uptime(id string) (*nova.Hypervisor, error) {
	resp, err := c.NovaV2.Get(utility.UrlJoin("os-hypervisors", id, "uptime"), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*nova.Hypervisor{"hypervisor": {}}
	resp.BodyUnmarshal(&body)
	return body["hypervisor"], nil
}
func (c KeypairApi) List(query url.Values) ([]nova.Keypair, error) {
	resp, err := c.NovaV2.Get("keypairs", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.Keypair{"keypairs": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["keypairs"], nil
}

// service api

func (c ComputeServiceApi) List(query url.Values) ([]nova.Service, error) {
	resp, err := c.NovaV2.Get("os-services", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.Service{"services": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["services"], nil
}
func (c ComputeServiceApi) GetByHostBinary(host string, binary string) (*nova.Service, error) {
	services, err := c.List(utility.UrlValues(map[string]string{
		"host": host, "binary": binary,
	}))
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("service %s:%s not found", host, binary)
	}
	return &services[0], nil
}

func (c ComputeServiceApi) doAction(action string, host string, binary string) (*nova.Service, error) {
	reqBody, _ := json.Marshal(nova.Service{Binary: binary, Host: host})
	resp, err := c.NovaV2.Put(utility.UrlJoin("os-services", action), reqBody, nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*nova.Service{"service": {}}
	resp.BodyUnmarshal(&body)
	return body["service"], nil
}
func (c ComputeServiceApi) update(id string, update map[string]interface{}) (*nova.Service, error) {
	reqBody, _ := json.Marshal(update)
	resp, err := c.NovaV2.Put(utility.UrlJoin("os-services", id), reqBody, nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*nova.Service{"service": {}}
	resp.BodyUnmarshal(&body)
	return body["service"], nil
}
func (c ComputeServiceApi) Up(host string, binary string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return c.update(service.Id, map[string]interface{}{"forced_down": false})
	}
	return c.doAction("up", host, binary)
}
func (c ComputeServiceApi) Down(host string, binary string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return c.update(service.Id, map[string]interface{}{"forced_down": true})
	}
	return c.doAction("down", host, binary)
}
func (c ComputeServiceApi) Enable(host string, binary string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return c.update(service.Id, map[string]interface{}{"status": "enabled"})
	}
	return c.doAction("enable", host, binary)
}
func (c ComputeServiceApi) Disable(host string, binary string, reason string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		body := map[string]interface{}{"status": "disabled"}
		if reason != "" {
			body["disabled_reason"] = reason
		}
		return c.update(service.Id, body)
	}
	return c.doAction("disable", host, binary)
}
func (c ComputeServiceApi) Delete(host string, binary string) error {
	service, err := c.GetByHostBinary(host, binary)
	if err != nil {
		return err
	}
	_, err = c.NovaV2.Delete(utility.UrlJoin("os-services", service.Id), nil)
	return err
}

// migration api

func (c MigrationApi) List(query url.Values) ([]nova.Migration, error) {
	resp, err := c.NovaV2.Get("os-migrations", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]nova.Migration{"migrations": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["migrations"], nil
}

// avaliable zone api
func (c AZApi) List(query url.Values) ([]nova.AvailabilityZone, error) {
	resp, err := c.NovaV2.Get("os-availability-zone", query)
	if err != nil {
		return nil, err
	}
	respBody := map[string][]nova.AvailabilityZone{"availabilityZoneInfo": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["availabilityZoneInfo"], nil
}
func (c AZApi) Detail(query url.Values) ([]nova.AvailabilityZone, error) {
	resp, err := c.NovaV2.Get("os-availability-zone/detail", query)
	if err != nil {
		return nil, err
	}
	respBody := map[string][]nova.AvailabilityZone{"availabilityZoneInfo": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["availabilityZoneInfo"], nil
}
func (c AggregateApi) List(query url.Values) ([]nova.Aggregate, error) {
	resp, err := c.NovaV2.Get("os-aggregates", query)
	if err != nil {
		return nil, err
	}
	respBody := map[string][]nova.Aggregate{"aggregates": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["aggregates"], nil
}
func (c AggregateApi) Show(aggregate string) (*nova.Aggregate, error) {
	resp, err := c.NovaV2.Get("aggregates", nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*nova.Aggregate{"aggregate": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["aggregate"], nil
}

// server group api

func (c ServerGroupApi) List(query url.Values) ([]nova.ServerGroup, error) {
	resp, err := c.NovaV2.Get("os-server-groups", query)
	if err != nil {
		return nil, err
	}
	respBody := map[string][]nova.ServerGroup{"server_groups": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["server_groups"], nil
}

func (c ServersApi) WaitStatus(serverId string, status string, interval int) (*nova.Server, error) {
	for {
		server, err := c.Show(serverId)
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
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func (c ServersApi) WaitDeleted(id string) error {
	for {
		server, err := c.Show(id)
		if err != nil {
			if httpError, ok := err.(*utility.HttpError); ok {
				if httpError.Status == 404 {
					return nil
				}
			}
			return err
		}
		logging.Debug("servers status is %s", server.Status)
		time.Sleep(time.Second * 2)
	}
}
func (c ServersApi) WaitTask(id string, taskState string) (*nova.Server, error) {
	for {
		server, err := c.Show(id)
		if err != nil {
			return nil, err
		}
		logging.Info("server %s status: %s", id, server.AllStatus())

		if strings.ToUpper(server.Status) == "ERROR" {
			return nil, fmt.Errorf("server %s status is ERROR", id)
		}
		if strings.EqualFold(server.TaskState, strings.ToUpper(taskState)) {
			return server, nil
		}
		time.Sleep(time.Second * 2)
	}
}
func (c ServersApi) WaitResized(id string, newFlavorName string) (*nova.Server, error) {
	server, err := c.WaitTask(id, "")
	if err != nil {
		return nil, err
	}
	if server.Flavor.OriginalName == newFlavorName {
		return server, err
	}
	return nil, fmt.Errorf("server %s not resized", id)
}
func (c ServersApi) WaitRebooted(id string, newFlavorName string) (*nova.Server, error) {
	server, err := c.WaitTask(id, "")
	if err != nil {
		return nil, err
	}
	return c.WaitStatus(server.Id, "ACTIVE", 5)
}
func (c ServersApi) CreateAndWait(options nova.ServerOpt) (*nova.Server, error) {
	server, err := c.Create(options)
	if err != nil {
		return server, err
	}
	if server.Id == "" {
		return server, fmt.Errorf("create server failed")
	}
	server, err = c.WaitTask(server.Id, "")
	if err != nil {
		return server, err
	}
	return c.WaitStatus(server.Id, "ACTIVE", 5)
}

func (c ServersApi) Prune(query url.Values, yes bool, waitDeleted bool) {
	if len(query) == 0 {
		query.Set("status", "error")
	}
	logging.Info("查询虚拟机: %v", query.Encode())
	servers, err := c.Servers().List(query)
	utility.LogError(err, "query servers failed", true)
	logging.Info("需要清理的虚拟机数量: %d\n", len(servers))
	if len(servers) == 0 {
		return
	}
	if !yes {
		fmt.Printf("即将删除 %d 个虚拟机:\n", len(servers))
		for _, server := range servers {
			fmt.Printf("    %s(%s)\n", server.Id, server.Name)
		}
		yes = utility.ScanfComfirm("是否删除", []string{"yes", "y"}, []string{"no", "n"})
	}
	if !yes {
		return
	}
	logging.Info("开始删除虚拟机")
	tg := utility.TaskGroup{
		Items: servers,
		Func: func(o interface{}) error {
			s := o.(nova.Server)
			logging.Info("delete %s", s.Id)
			err := c.Servers().Delete(s.Id)
			if err != nil {
				return fmt.Errorf("delete %s failed: %v", s.Id, err)
			}
			if waitDeleted {
				c.Servers().WaitDeleted(s.Id)
			}
			return nil
		},
		ShowProgress: true,
	}
	tg.Start()
}
