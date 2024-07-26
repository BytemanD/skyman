package openstack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/compare"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/BytemanD/skyman/utility/httpclient"
	"github.com/go-resty/resty/v2"
)

const (
	V2_1                    = "v2.1"
	URL_DETAIL              = "detail"
	URL_SERVER_VOLUMES_BOOT = "os-volumes_boot"
	URL_VOLUME_ATTACH       = "%s/os-volume_attachments"
	URL_VOLUME_DETACH       = "%s/os-volume_attachments/%s"
	URL_INTERFACE_ATTACH    = "%s/os-interface"
	URL_INTERFACE_DETACH    = "%s/os-interface/%s"
)

type microVersion struct {
	Version      int
	MicroVersion int
}

type NovaV2 struct {
	RestClient2
	currentVersion *model.ApiVersion
	MicroVersion   model.ApiVersion
	region         string
}

func (o *Openstack) NovaV2() *NovaV2 {
	if o.novaClient == nil {
		endpoint, err := o.AuthPlugin.GetServiceEndpoint("compute", "nova", "public")
		if err != nil {
			logging.Fatal("get nova endpoint falied: %v", err)
		}
		o.novaClient = &NovaV2{
			RestClient2: NewRestClient2(utility.VersionUrl(endpoint, V2_1), o.AuthPlugin),
			region:      o.AuthPlugin.Region(),
		}
		currentVersion, err := o.novaClient.GetCurrentVersion()
		if err != nil {
			logging.Warning("get current version failed: %v", err)
		} else {
			o.novaClient.MicroVersion = *currentVersion
		}
		o.novaClient.RestClient2.AddBaseHeader("Openstack-Api-Version", o.novaClient.MicroVersion.Version)
		o.novaClient.RestClient2.AddBaseHeader("X-Openstack-Nova-Api-Version", o.novaClient.MicroVersion.Version)
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
func (c *NovaV2) GetCurrentVersion() (*model.ApiVersion, error) {
	if c.currentVersion == nil {
		resp, err := c.Index()
		if err != nil {
			return nil, err
		}
		respBody := struct{ Versions []model.ApiVersion }{}
		if err := json.Unmarshal(resp.Body(), &respBody); err != nil {
			return nil, err
		}
		for _, version := range respBody.Versions {
			if version.Status == "CURRENT" {
				return &version, nil
			}
		}
	}
	if c.currentVersion != nil {
		return c.currentVersion, nil
	}
	return nil, fmt.Errorf("current version not found")
}

func (c *NovaV2) String() string {
	return fmt.Sprintf("<%s %s>", c.BaseUrl, c.region)
}

// server api
type serverApi struct {
	ResourceApi
}

func novaResourceApi(c NovaV2, baseUrl, singularKey, pluralKey string) ResourceApi {
	api := ResourceApi{
		Endpoint:     c.BaseUrl,
		BaseUrl:      baseUrl,
		Client:       c.session,
		MicroVersion: c.MicroVersion,
		SingularKey:  singularKey,
		PluralKey:    pluralKey,
	}
	api.SetHeaders(c.session.BaseHeaders)
	return api
}

func (c NovaV2) Server() serverApi {
	return serverApi{
		ResourceApi: novaResourceApi(c, "servers", "server", "servers"),
	}
}

func (c serverApi) List(query url.Values) ([]nova.Server, error) {
	body := struct{ Servers []nova.Server }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Servers, nil
}
func (c serverApi) ListByName(name string) ([]nova.Server, error) {
	query := url.Values{}
	query.Set("name", name)
	return c.List(query)
}
func (c serverApi) Detail(query url.Values) ([]nova.Server, error) {
	body := struct{ Servers []nova.Server }{}
	if _, err := c.AppendUrl(URL_DETAIL).SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Servers, nil

}
func (c serverApi) DetailByName(name string) ([]nova.Server, error) {
	return c.Detail(utility.UrlValues(map[string]string{"name": name}))
}

func (c serverApi) Show(id string) (*nova.Server, error) {
	body := struct{ Server nova.Server }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Server, nil
}
func (c serverApi) Found(idOrName string) (*nova.Server, error) {
	server, err := FoundResource[nova.Server](c.ResourceApi, idOrName)
	if err != nil {
		return nil, err
	}
	if server.Status == "" {
		server, err = c.Show(server.Id)
	}
	return server, err
}
func (c serverApi) Create(options nova.ServerOpt) (*nova.Server, error) {
	if options.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if options.Flavor == "" {
		return nil, fmt.Errorf("flavor is empty")
	}
	if options.Networks == nil {
		options.Networks = "none"
	}
	if options.MinCount == 0 {
		options.MinCount = 1
	}
	if options.MaxCount == 0 {
		options.MaxCount = 1
	}
	var err error
	body := struct{ Server nova.Server }{}
	if options.BlockDeviceMappingV2 != nil {
		_, err = c.SetUrl(URL_SERVER_VOLUMES_BOOT).
			SetBody(map[string]nova.ServerOpt{"server": options}).Post(&body)
	} else {
		_, err = c.SetBody(map[string]nova.ServerOpt{"server": options}).Post(&body)
	}
	if err != nil {
		return nil, err
	}
	return &body.Server, nil
}

func (c serverApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}
func (c serverApi) ListVolumes(id string) ([]nova.VolumeAttachment, error) {
	body := struct{ VolumeAttachments []nova.VolumeAttachment }{}
	if _, err := c.AppendUrlf(URL_VOLUME_ATTACH, id).Get(&body); err != nil {
		return nil, err
	}
	return body.VolumeAttachments, nil
}
func (c serverApi) AddVolume(id string, volumeId string) (*nova.VolumeAttachment, error) {
	body := struct{ VolumeAttachment nova.VolumeAttachment }{}
	_, err := c.AppendUrlf(URL_VOLUME_ATTACH, id).
		SetBody(map[string]map[string]string{"volumeAttachment": {"volumeId": volumeId}}).
		Post(&body)
	if err != nil {
		return nil, err
	}
	return &body.VolumeAttachment, nil
}
func (c serverApi) DeleteVolume(id string, volumeId string) error {
	_, err := c.AppendUrlf(URL_VOLUME_DETACH, id, volumeId).Delete(nil)
	return err
}
func (c serverApi) DeleteVolumeAndWait(id string, volumeId string, waitSeconds int) error {
	err := c.DeleteVolume(id, volumeId)
	if err != nil {
		return err
	}
	startTime := time.Now()
	for {
		detached := true
		attachedVolumes, err := c.ListVolumes(id)
		if err != nil {
			return fmt.Errorf("list server interfaces failed: %s", err)
		}
		for _, vol := range attachedVolumes {
			if vol.VolumeId == volumeId {
				detached = false
				break
			}
		}
		if detached {
			return nil
		}
		if time.Since(startTime) >= time.Second*time.Duration(waitSeconds) {
			return fmt.Errorf("interface %s is not detached after %d seconds", volumeId, waitSeconds)
		}
		time.Sleep(time.Second * 2)
	}
}
func (c serverApi) ListInterfaces(id string) ([]nova.InterfaceAttachment, error) {
	body := struct{ InterfaceAttachments []nova.InterfaceAttachment }{}
	if _, err := c.AppendUrlf(URL_INTERFACE_ATTACH, id).Get(&body); err != nil {
		return nil, err
	}
	return body.InterfaceAttachments, nil
}
func (c serverApi) AddInterface(id string, netId, portId string) (*nova.InterfaceAttachment, error) {
	params := map[string]string{}
	if netId == "" && portId == "" {
		return nil, errors.New("invalid params: portId or netId is required")
	}
	if netId != "" {
		params["net_id"] = netId
	}
	if portId != "" {
		params["port_id"] = portId
	}
	body := struct{ InterfaceAttachment nova.InterfaceAttachment }{}
	resp, err := c.AppendUrlf(URL_INTERFACE_ATTACH, id).
		SetBody(map[string]map[string]string{"interfaceAttachment": params}).
		Post(&body)
	if err != nil {
		return nil, err
	}
	body.InterfaceAttachment.SetRequestId(resp.Header().Get(X_OPENSTACK_REQUEST_ID))
	return &body.InterfaceAttachment, nil
}
func (c serverApi) DeleteInterface(id string, portId string) (*resty.Response, error) {
	return c.AppendUrlf(URL_INTERFACE_DETACH, id, portId).
		Delete(nil)
}
func (c serverApi) DeleteInterfaceAndWait(id string, portId string, timeout time.Duration) error {
	resp, err := c.DeleteInterface(id, portId)
	if err != nil {
		return err
	}
	reqId := resp.Header().Get(X_OPENSTACK_REQUEST_ID)
	logging.Info("[%s] detaching interface %s, request id: %s", id, portId, reqId)

	return utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:     timeout,
			IntervalMin: time.Second * 2},
		[]string{"ActionNotFinishedError"},
		func() error {
			action, err := c.ShowAction(id, reqId)
			if err != nil {
				return fmt.Errorf("get action events failed: %s", err)
			}
			if len(action.Events) == 0 || action.Events[0].FinishTime == "" {
				return utility.NewActionNotFinishedError(reqId)
			}
			logging.Info("[%s] action result: %s", id, action.Events[0].Result)
			if action.Events[0].Result == "Error" {
				return fmt.Errorf("request %s is error", reqId)
			} else {
				return nil
			}
		},
	)
}
func (c serverApi) doAction(action string, id string, params interface{}) (*resty.Response, error) {
	return c.SetBody(map[string]interface{}{action: params}).AppendUrl(id).AppendUrl("action").Post(nil)
}
func (c serverApi) Stop(id string) error {
	_, err := c.doAction("os-stop", id, nil)
	return err
}
func (c serverApi) Start(id string) error {
	_, err := c.doAction("os-start", id, nil)
	return err
}
func (c serverApi) Reboot(id string, hard bool) error {
	rebootTypes := map[bool]string{false: "soft", true: "hard"}
	body := map[string]string{"type": rebootTypes[hard]}
	_, err := c.doAction("reboot", id, body)
	return err
}
func (c serverApi) Pause(id string) error {
	_, err := c.doAction("pause", id, nil)
	return err
}
func (c serverApi) Unpause(id string) error {
	_, err := c.doAction("unpause", id, nil)
	return err
}
func (c serverApi) Shelve(id string) error {
	_, err := c.doAction("shelve", id, nil)
	return err
}
func (c serverApi) Unshelve(id string) error {
	_, err := c.doAction("unshelve", id, nil)
	return err
}
func (c serverApi) Suspend(id string) error {
	_, err := c.doAction("suspend", id, nil)
	return err
}
func (c serverApi) Resume(id string) error {
	_, err := c.doAction("resume", id, nil)
	return err
}
func (c serverApi) Resize(id string, flavorRef string) error {
	_, err := c.doAction("resize", id, map[string]string{"flavorRef": flavorRef})
	return err
}
func (c serverApi) ResizeConfirm(id string) error {
	_, err := c.doAction("confirmResize", id, nil)
	return err
}
func (c serverApi) ResizeRevert(id string) error {
	_, err := c.doAction("revertResize", id, nil)
	return err
}

// TODO: more params
func (c serverApi) Rebuild(id string, options map[string]interface{}) error {
	_, err := c.doAction("rebuild", id, options)
	return err
}
func (c serverApi) Evacuate(id string, password string, host string, force bool) error {
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
func (c serverApi) SetPassword(id string, password, user string) error {
	data := map[string]interface{}{
		"adminPass": password,
	}
	if user != "" {
		data["userName"] = user
	}
	_, err := c.doAction("changePassword", id, data)
	return err
}
func (c serverApi) SetName(id string, name string) error {
	data := map[string]interface{}{"name": name}
	_, err := c.doAction("adminPass", id, data)
	return err
}
func (c serverApi) SetState(id string, active bool) error {
	data := map[string]interface{}{}
	if active {
		data["state"] = "active"
	} else {
		data["state"] = "error"
	}
	_, err := c.doAction("os-resetState", id, data)
	return err
}
func (c serverApi) ConsoleLog(id string, length uint) (*nova.ConsoleLog, error) {
	params := map[string]interface{}{}
	if length != 0 {
		params["length"] = length
	}
	resp, err := c.doAction("os-getConsoleOutput", id, params)
	if err != nil {
		return nil, err
	}
	body := nova.ConsoleLog{}
	json.Unmarshal(resp.Body(), &body)
	return &body, nil
}
func (c serverApi) getVNCConsole(id string, consoleType string) (*nova.Console, error) {
	params := map[string]interface{}{"type": consoleType}
	resp, err := c.doAction("os-getVNCConsole", id, params)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*nova.Console{"console": {}}
	json.Unmarshal(resp.Body(), &respBody)
	return respBody["console"], nil
}
func (c serverApi) getRemoteConsole(id string, protocol string, consoleType string) (*nova.Console, error) {
	params := map[string]interface{}{
		"remote_console": map[string]interface{}{
			"protocol": protocol,
			"type":     consoleType,
		},
	}
	respBody := struct {
		RemoteConsole nova.Console `json:"remote_console"`
	}{}
	_, err := c.AppendUrl(id).AppendUrl("remote-consoles").SetBody(params).Post(&respBody)
	if err != nil {
		return nil, err
	}
	return &respBody.RemoteConsole, nil
}

func (c serverApi) ConsoleUrl(id string, consoleType string) (*nova.Console, error) {
	if c.MicroVersionLargeEqual("2.6") {
		// TODO: do not set "vnc" directly
		return c.getRemoteConsole(id, "vnc", consoleType)
	}
	return c.getVNCConsole(id, consoleType)
}
func (c serverApi) Migrate(id string, host string) error {
	data := map[string]interface{}{}
	if host != "" {
		server, err := c.Show(id)
		if err != nil {
			return err
		}
		if server.Host == host {
			return fmt.Errorf("server %s host is %s", id, server.Host)
		}
		data["host"] = host
	} else {
		data["host"] = nil
	}
	_, err := c.doAction("migrate", id, data)
	return err
}
func (c serverApi) LiveMigrate(id string, blockMigrate interface{}, host string) error {
	data := map[string]interface{}{"block_migration": blockMigrate}
	if host != "" {
		server, err := c.Show(id)
		if err != nil {
			return err
		}
		if server.Host == host {
			return fmt.Errorf("server %s host is %s", id, server.Host)
		}
		data["host"] = host
	} else {
		data["host"] = nil
	}
	_, err := c.doAction("os-migrateLive", id, data)
	return err
}
func (c serverApi) ListActions(id string) ([]nova.InstanceAction, error) {
	body := struct{ InstanceActions []nova.InstanceAction }{}
	if _, err := c.AppendUrl(id).AppendUrl("os-instance-actions").Get(&body); err != nil {
		return nil, err
	}
	return body.InstanceActions, nil
}
func (c serverApi) ShowAction(id, requestId string) (*nova.InstanceAction, error) {
	body := struct{ InstanceAction nova.InstanceAction }{}
	if _, err := c.AppendUrl(id).AppendUrl("os-instance-actions").AppendUrl(requestId).Get(&body); err != nil {
		return nil, err
	}
	return &body.InstanceAction, nil
}
func (c serverApi) ListActionsWithEvents(id string, actionName string, requestId string, last int) ([]nova.InstanceAction, error) {
	serverActions, err := c.ListActions(id)
	if err != nil {
		return nil, err
	}
	filterActions := []nova.InstanceAction{}
	for _, action := range serverActions {
		if requestId != "" && action.RequestId != requestId {
			continue
		}
		if actionName != "" && action.Action != actionName {
			continue
		}
		filterActions = append(filterActions, action)
	}

	if last == 0 {
		last = len(filterActions)
	}
	start := max(len(filterActions)-last, 0)
	actionWithEvents := []nova.InstanceAction{}
	for _, action := range filterActions[start:] {
		serverAction, err := c.ShowAction(id, action.RequestId)
		if err != nil {
			logging.Error("get server action %s failed: %s", action.RequestId, err)
		}
		actionWithEvents = append(actionWithEvents, *serverAction)
	}
	return actionWithEvents, nil
}

func (c serverApi) ListMigrations(id string, query url.Values) ([]nova.Migration, error) {
	body := struct{ Migrations []nova.Migration }{}
	if _, err := c.AppendUrl(id).AppendUrl("os-migrations").Get(&body); err != nil {
		return nil, err
	}
	return body.Migrations, nil

}
func (c serverApi) RegionLiveMigrate(id string, destRegion string, blockMigrate bool, dryRun bool, destHost string) (*nova.RegionMigrateResp, error) {
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
	json.Unmarshal(resp.Body(), &respBody)
	return respBody, nil
}
func (c serverApi) WaitStatus(serverId string, status string, interval int) (*nova.Server, error) {
	var (
		server *nova.Server
		err    error
	)
	utility.Retry(
		utility.RetryCondition{
			Timeout: time.Second * 60 * 10, IntervalMin: time.Second * time.Duration(interval),
		},
		func() bool {
			server, err = c.Show(serverId)
			if err != nil {
				return false
			}
			logging.Info("[server: %s] status: %s, taskState: %s", server.Id, server.Status, server.TaskState)
			switch strings.ToUpper(server.Status) {
			case "ERROR":
				err = fmt.Errorf("server status is error, message: %s", server.Fault.Message)
				return false
			case strings.ToUpper(status):
				return false
			}
			return true

		},
	)
	return server, err
}
func (c serverApi) WaitBooted(id string) (*nova.Server, error) {
	for {
		server, err := c.Show(id)
		if err != nil {
			return server, err
		}
		logging.Info("[%s] %s", id, server.AllStatus())
		if server.IsError() {
			return server, fmt.Errorf("server %s is error", server.Id)
		}
		if server.IsActive() && server.Host != "" {
			return server, nil
		}
		time.Sleep(time.Second * 2)
	}
}
func (c serverApi) WaitDeleted(id string) error {
	var (
		server *nova.Server
		err    error
	)
	cxt := context.TODO()
	cxt.Done()

	utility.Retry(
		utility.RetryCondition{
			Timeout:     time.Second * 60 * 10,
			IntervalMin: time.Second * time.Duration(2)},
		func() bool {
			server, err = c.Show(id)
			if err == nil {
				logging.Info("[%s] %s", id, server.AllStatus())
				return true
			}
			if compare.IsType[httpclient.HttpError](err) {
				httpError, _ := err.(httpclient.HttpError)
				if httpError.IsNotFound() {
					logging.Info("[%s] deleted", id)
					err = nil
					return false
				}
			}
			return false
		},
	)
	return err
}
func (c serverApi) WaitTask(id string, taskState string) (*nova.Server, error) {
	for {
		server, err := c.Show(id)
		if err != nil {
			return nil, err
		}
		logging.Info("[%s] %s progress: %d", id, server.AllStatus(), int(server.Progress))

		if strings.ToUpper(server.Status) == "ERROR" {
			return nil, fmt.Errorf("server %s status is ERROR", id)
		}
		if strings.EqualFold(server.TaskState, strings.ToUpper(taskState)) {
			return server, nil
		}
		time.Sleep(time.Second * 2)
	}
}
func (c serverApi) WaitResized(id string, newFlavorName string) (*nova.Server, error) {
	server, err := c.WaitTask(id, "")
	if err != nil {
		return nil, err
	}
	if server.Flavor.OriginalName == newFlavorName {
		return server, err
	}
	return nil, fmt.Errorf("server %s not resized", id)
}
func (c serverApi) WaitRebooted(id string, newFlavorName string) (*nova.Server, error) {
	server, err := c.WaitTask(id, "")
	if err != nil {
		return nil, err
	}
	return c.WaitStatus(server.Id, "ACTIVE", 5)
}
func (c serverApi) CreateAndWait(options nova.ServerOpt) (*nova.Server, error) {
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
	return c.WaitBooted(server.Id)
}

// flavor api
type flavorApi struct{ ResourceApi }

func (c NovaV2) Flavor() flavorApi {
	return flavorApi{
		ResourceApi: novaResourceApi(c, "flavors", "flavor", "flavors"),
	}
}
func (c flavorApi) List(query url.Values) ([]nova.Flavor, error) {
	body := struct{ Flavors []nova.Flavor }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Flavors, nil
}
func (c flavorApi) Detail(query url.Values) ([]nova.Flavor, error) {
	body := struct{ Flavors []nova.Flavor }{}
	if _, err := c.AppendUrl("detail").SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Flavors, nil
}
func (c flavorApi) Show(id string) (*nova.Flavor, error) {
	body := struct{ Flavor nova.Flavor }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Flavor, nil
}
func (c flavorApi) ShowWithExtraSpecs(id string) (*nova.Flavor, error) {
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
func (c flavorApi) SetExtraSpecs(id string, extraSpecs map[string]string) (nova.ExtraSpecs, error) {
	respBody := struct {
		ExtraSpecs nova.ExtraSpecs `json:"extra_specs"`
	}{}
	_, err := c.AppendUrl(id).AppendUrl("os-extra_specs").
		SetBody(map[string]nova.ExtraSpecs{"extra_specs": extraSpecs}).
		Post(&respBody)
	return respBody.ExtraSpecs, err
}
func (c flavorApi) DeleteExtraSpec(id string, extraSpec string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}
func (c flavorApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}
func (c flavorApi) Found(idOrName string) (*nova.Flavor, error) {
	return FoundResource[nova.Flavor](c.ResourceApi, idOrName)
}
func (c flavorApi) Create(flavor nova.Flavor) (*nova.Flavor, error) {
	respBody := struct{ Flavor nova.Flavor }{}
	_, err := c.SetBody(map[string]nova.Flavor{"flavor": flavor}).Post(&respBody)
	return &respBody.Flavor, err
}
func (c flavorApi) ListExtraSpecs(id string) (nova.ExtraSpecs, error) {
	respBody := struct {
		ExtraSpecs nova.ExtraSpecs `json:"extra_specs"`
	}{}
	_, err := c.AppendUrl(id).AppendUrl("os-extra_specs").Get(&respBody)
	return respBody.ExtraSpecs, err
}
func (c flavorApi) Copy(id string, newName string, newId string,
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
type hypervisorApi struct{ ResourceApi }

func (c NovaV2) Hypervisor() hypervisorApi {
	return hypervisorApi{
		ResourceApi: novaResourceApi(c, "os-hypervisors", "hypervisor", "hypervisors"),
	}
}
func (c hypervisorApi) List(query url.Values) ([]nova.Hypervisor, error) {
	body := struct{ Hypervisors []nova.Hypervisor }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Hypervisors, nil
}
func (c hypervisorApi) Detail(query url.Values) ([]nova.Hypervisor, error) {
	body := struct{ Hypervisors []nova.Hypervisor }{}
	if _, err := c.SetQuery(query).AppendUrl("detail").Get(&body); err != nil {
		return nil, err
	}
	return body.Hypervisors, nil
}
func (c hypervisorApi) ListByName(hostname string) ([]nova.Hypervisor, error) {
	return c.List(utility.UrlValues(map[string]string{"hypervisor_hostname_pattern": hostname}))
}

func (c hypervisorApi) Show(id string) (*nova.Hypervisor, error) {
	body := struct{ Hypervisor nova.Hypervisor }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Hypervisor, nil
}
func (c hypervisorApi) ShowByHostname(hostname string) (*nova.Hypervisor, error) {
	hypervisors, err := c.ListByName(hostname)
	if err != nil {
		return nil, err
	}
	if len(hypervisors) == 0 {
		return nil, fmt.Errorf("hypervisor %s not found", hostname)
	}
	return c.Show(hypervisors[0].Id)
}
func (c hypervisorApi) Found(idOrHostName string) (*nova.Hypervisor, error) {
	if stringutils.IsUUID(idOrHostName) {
		hypervisor, err := c.Show(idOrHostName)
		if err == nil {
			return hypervisor, nil
		}
	}
	return c.ShowByHostname(idOrHostName)
}
func (c hypervisorApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}
func (c hypervisorApi) Uptime(id string) (*nova.Hypervisor, error) {
	body := struct{ Hypervisor *nova.Hypervisor }{}
	_, err := c.AppendUrl(id).AppendUrl("uptime").Get(&body)
	if err != nil {
		return nil, err
	}
	return body.Hypervisor, nil
}

// keypair api
type keypairApi struct{ ResourceApi }

func (c NovaV2) Keypair() keypairApi {
	return keypairApi{
		ResourceApi: novaResourceApi(c, "keypairs", "keypair", "keypairs"),
	}
}
func (c keypairApi) List(query url.Values) ([]nova.Keypair, error) {
	body := struct{ Keypairs []nova.Keypair }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Keypairs, nil
}

// service api
type computeServiceApi struct{ ResourceApi }

func (c NovaV2) Service() computeServiceApi {
	return computeServiceApi{
		ResourceApi: novaResourceApi(c, "os-services", "service", "services"),
	}
}
func (c computeServiceApi) List(query url.Values) ([]nova.Service, error) {
	body := struct{ Services []nova.Service }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Services, nil
}
func (c computeServiceApi) GetByHostBinary(host string, binary string) (*nova.Service, error) {
	services, err := c.List(utility.UrlValues(map[string]string{"host": host, "binary": binary}))
	if err != nil {
		return nil, err
	}
	switch len(services) {
	case 0:
		return nil, fmt.Errorf("service %s:%s not found", host, binary)
	case 1:
		return &services[0], nil
	default:
		return nil, fmt.Errorf("found multi service %s:%s not found", host, binary)
	}
}

func (c computeServiceApi) doAction(action string, params map[string]interface{}) error {
	respBody := struct{ Service nova.Service }{}
	_, err := c.AppendUrl(action).SetBody(params).Put(&respBody)
	return err
}
func (c computeServiceApi) update(id string, update map[string]interface{}) (*nova.Service, error) {
	respBody := struct{ Service nova.Service }{}
	_, err := c.SetBody(update).AppendUrl(id).Put(&respBody)
	if err != nil {
		return nil, err
	}
	return &respBody.Service, nil
}
func (c computeServiceApi) Up(host string, binary string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return c.update(service.Id, map[string]interface{}{"forced_down": false})
	}
	err := c.doAction("force-down", map[string]interface{}{
		"host": host, "binary": binary, "forced_down": false,
	})
	if err != nil {
		return nil, err
	} else {
		return c.GetByHostBinary(host, binary)
	}
}
func (c computeServiceApi) Down(host string, binary string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return c.update(service.Id, map[string]interface{}{"forced_down": true})
	}
	err := c.doAction("force-down", map[string]interface{}{
		"host": host, "binary": binary, "forced_down": true,
	})
	if err != nil {
		return nil, err
	} else {
		return c.GetByHostBinary(host, binary)
	}
}
func (c computeServiceApi) Enable(host string, binary string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return c.update(service.Id, map[string]interface{}{"status": "enabled"})
	}
	err := c.doAction("enable", map[string]interface{}{
		"host": host, "binary": binary, "status": "enabled",
	})
	if err != nil {
		return nil, err
	} else {
		return c.GetByHostBinary(host, binary)
	}
}
func (c computeServiceApi) Disable(host string, binary string, reason string) (*nova.Service, error) {
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
	err := c.doAction("disable", map[string]interface{}{
		"host": host, "binary": binary, "status": "disabled",
	})
	if err != nil {
		return nil, err
	} else {
		return c.GetByHostBinary(host, binary)
	}
}
func (c computeServiceApi) Delete(host string, binary string) error {
	service, err := c.GetByHostBinary(host, binary)
	if err != nil {
		return err
	}
	_, err = c.AppendUrl(service.Id).Delete(nil)
	return err
}

// migration api
type migrationApi struct{ ResourceApi }

func (c NovaV2) Migration() migrationApi {
	return migrationApi{
		ResourceApi: novaResourceApi(c, "os-migrations", "migration", "migrations"),
	}
}
func (c migrationApi) List(query url.Values) ([]nova.Migration, error) {
	body := struct{ Migrations []nova.Migration }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Migrations, nil
}

// avaliable zone api
type azApi struct{ ResourceApi }

func (c NovaV2) AZ() azApi {
	return azApi{
		ResourceApi: novaResourceApi(c, "os-availability-zone", "", "availabilityZoneInfo"),
	}
}

func (c azApi) List(query url.Values) ([]nova.AvailabilityZone, error) {
	body := struct{ AvailabilityZoneInfo []nova.AvailabilityZone }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.AvailabilityZoneInfo, nil
}
func (c azApi) Detail(query url.Values) ([]nova.AvailabilityZone, error) {
	body := struct{ AvailabilityZoneInfo []nova.AvailabilityZone }{}
	if _, err := c.SetQuery(query).AppendUrl("detail").Get(&body); err != nil {
		return nil, err
	}
	return body.AvailabilityZoneInfo, nil
}

// aggregate api
type aggregateApi struct{ ResourceApi }

func (c NovaV2) Aggregate() aggregateApi {
	return aggregateApi{
		ResourceApi: novaResourceApi(c, "os-aggregates", "aggregate", "aggregates"),
	}
}
func (c aggregateApi) List(query url.Values) ([]nova.Aggregate, error) {
	body := struct{ Aggregates []nova.Aggregate }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Aggregates, nil
}
func (c aggregateApi) Show(id string) (*nova.Aggregate, error) {
	body := struct{ Aggregate nova.Aggregate }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Aggregate, nil
}

// server group api
type serverGroupApi struct{ ResourceApi }

func (c NovaV2) ServerGroup() serverGroupApi {
	return serverGroupApi{
		ResourceApi: novaResourceApi(c, "os-server-groups", "server_group", "server_groups"),
	}
}
func (c serverGroupApi) List(query url.Values) ([]nova.ServerGroup, error) {
	body := struct{ ServerGroups []nova.ServerGroup }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.ServerGroups, nil
}
