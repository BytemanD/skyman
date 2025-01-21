package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/compare"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/BytemanD/skyman/utility"
)

const (
	URL_DETAIL              = "detail"
	URL_SERVER_VOLUMES_BOOT = "os-volumes_boot"
	URL_VOLUME_ATTACH       = "%s/os-volume_attachments"
	URL_VOLUME_DETACH       = "%s/os-volume_attachments/%s"
	URL_INTERFACE_ATTACH    = "servers/%s/os-interface"
	URL_INTERFACE_DETACH    = "servers/%s/os-interface/%s"
)

var COMPUTE_API_VERSION string

type microVersion struct {
	Version      int
	MicroVersion int
}

func (v microVersion) LargeEqual(other string) bool {
	otherMicroVersion := getMicroVersion(other)
	if v.Version > otherMicroVersion.Version {
		return true
	} else if v.Version == otherMicroVersion.Version {
		return v.MicroVersion >= otherMicroVersion.MicroVersion
	} else {
		return false
	}
}

type NovaV2 struct {
	*ServiceClient
	currentVersion *model.ApiVersion
	MicroVersion   *model.ApiVersion
}

func getMicroVersion(vertionStr string) microVersion {
	versionList := strings.Split(vertionStr, ".")
	v, _ := strconv.Atoi(versionList[0])
	micro, _ := strconv.Atoi(versionList[1])
	return microVersion{Version: v, MicroVersion: micro}
}

func (c *NovaV2) MicroVersionLargeEqual(version string) bool {
	clientVersion := getMicroVersion(c.MicroVersion.Version)
	otherVersion := getMicroVersion(version)
	if clientVersion.Version != otherVersion.Version {
		return clientVersion.Version > otherVersion.Version
	}
	return clientVersion.MicroVersion >= otherVersion.MicroVersion
}
func (c *NovaV2) GetCurrentVersion() (*model.ApiVersion, error) {
	if c.currentVersion == nil {
		result := struct{ Versions model.ApiVersions }{}
		if resp, err := c.Index(nil); err != nil {
			return nil, err
		} else {
			if err := resp.UnmarshalBody(&result); err != nil {
				return nil, err
			}
		}
		c.currentVersion = result.Versions.Current()

	}
	if c.currentVersion != nil {
		return c.currentVersion, nil
	}
	return nil, fmt.Errorf("current version not found")
}

func (c *NovaV2) String() string {
	return fmt.Sprintf("<Compute: %s>", c.Url)
}

type ServerApi struct{ ResourceApi }
type ComputeServiceApi struct{ ResourceApi }
type FlavorApi struct{ ResourceApi }
type AggregateApi struct{ ResourceApi }
type AZApi struct{ ResourceApi }
type Api struct{ ResourceApi }
type KeypairApi struct{ ResourceApi }
type HypervisorApi struct{ ResourceApi }
type MigrationApi struct{ ResourceApi }
type ServerGroupApi struct{ ResourceApi }
type ComputeQuotaApi struct{ ResourceApi }

func (c NovaV2) Server() ServerApi {
	return ServerApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "servers",
			SingularKey:  "server",
			PluralKey:    "servers",
		},
	}
}
func (c NovaV2) Flavor() FlavorApi {
	return FlavorApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "flavors",
			SingularKey:  "flavor",
			PluralKey:    "flavors",
		},
	}
}
func (c NovaV2) Service() ComputeServiceApi {
	return ComputeServiceApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "os-services",
			SingularKey:  "service",
			PluralKey:    "services",
		},
	}
}
func (c NovaV2) Keypair() KeypairApi {
	return KeypairApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "os-keypairs",
			SingularKey:  "keypair",
			PluralKey:    "keypairs",
		},
	}
}
func (c NovaV2) Hypervisor() HypervisorApi {
	return HypervisorApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "os-hypervisors",
			SingularKey:  "hypervisor",
			PluralKey:    "hypervisors",
		},
	}
}

func (c NovaV2) Aggregate() AggregateApi {
	return AggregateApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "os-aggregates",
			SingularKey:  "aggregate",
			PluralKey:    "aggregates",
		},
	}
}
func (c NovaV2) AZ() AZApi {
	return AZApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "os-availability-zone",
			PluralKey:    "availabilityZoneInfo",
		},
	}
}
func (c NovaV2) Migraion() MigrationApi {
	return MigrationApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "os-migrations",
			SingularKey:  "migration",
			PluralKey:    "migrations",
		},
	}
}

func (c NovaV2) ServerGroup() ServerGroupApi {
	return ServerGroupApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "os-server-groups",
			SingularKey:  "server_group",
			PluralKey:    "server_groups",
		},
	}
}
func (c NovaV2) Quota() ComputeQuotaApi {
	return ComputeQuotaApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "os-quota-sets",
			SingularKey:  "quota_set",
			PluralKey:    "quota_sets",
		},
	}
}
func (c NovaV2) Migration() MigrationApi {
	return MigrationApi{
		ResourceApi{
			Client:       c.rawClient,
			BaseUrl:      c.Url,
			MicroVersion: c.MicroVersion,
			ResourceUrl:  "os-migrations",
			SingularKey:  "migration",
			PluralKey:    "migrations",
		},
	}
}
func (c ServerApi) List(query url.Values) ([]nova.Server, error) {
	return ListResource[nova.Server](c.ResourceApi, query)
}
func (c ServerApi) ListByName(name string) ([]nova.Server, error) {
	return c.List(url.Values{"name": []string{name}})
}
func (c ServerApi) Detail(query url.Values) ([]nova.Server, error) {
	return ListResource[nova.Server](c.ResourceApi, query, true)
}
func (c ServerApi) DetailByName(name string) ([]nova.Server, error) {
	return c.Detail(url.Values{"name": []string{name}})
}

func (c ServerApi) Show(id string) (*nova.Server, error) {
	return ShowResource[nova.Server](c.ResourceApi, id)
}
func (c ServerApi) Find(idOrName string) (*nova.Server, error) {
	return FindResource(idOrName, c.Show, c.Detail)
}
func (c ServerApi) Create(options nova.ServerOpt) (*nova.Server, error) {
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
		_, err = c.R().SetBody(map[string]nova.ServerOpt{"server": options}).
			SetResult(&body).ResetPath().Post(URL_SERVER_VOLUMES_BOOT)
	} else {
		_, err = c.R().SetBody(map[string]nova.ServerOpt{"server": options}).
			SetResult(&body).Post()
	}
	if err != nil {
		return nil, err
	}
	return &body.Server, nil
}

func (c ServerApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id)
	return err
}
func (c ServerApi) ListVolumes(id string) ([]nova.VolumeAttachment, error) {
	body := struct{ VolumeAttachments []nova.VolumeAttachment }{}
	if _, err := c.R().SetResult(&body).Get(id, "os-volume_attachments"); err != nil {
		return nil, err
	}
	return body.VolumeAttachments, nil
}
func (c ServerApi) AddVolume(id string, volumeId string) (*nova.VolumeAttachment, error) {
	body := struct{ VolumeAttachment nova.VolumeAttachment }{}
	_, err := c.R().SetResult(&body).
		SetBody(ReqBody{"volumeAttachment": {"volumeId": volumeId}}).
		Post(id, "os-volume_attachments")
	if err != nil {
		return nil, err
	}
	return &body.VolumeAttachment, nil
}
func (c ServerApi) DeleteVolume(id string, volumeId string) error {
	_, err := c.R().Delete(id, "os-volume_attachments", volumeId)
	return err
}
func (c ServerApi) DeleteVolumeAndWait(id string, volumeId string, waitSeconds int) error {
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
func (c ServerApi) ListInterfaces(id string) ([]nova.InterfaceAttachment, error) {
	body := struct{ InterfaceAttachments []nova.InterfaceAttachment }{}
	if _, err := c.R().SetResult(&body).Get(id, "os-interface"); err != nil {
		return nil, err
	}
	return body.InterfaceAttachments, nil
}
func (c ServerApi) AddInterface(id string, netId, portId string) (*nova.InterfaceAttachment, error) {
	params := map[string]interface{}{}
	if netId == "" && portId == "" {
		return nil, errors.New("invalid params: portId or netId is required")
	}
	if netId != "" {
		params["net_id"] = netId
	}
	if portId != "" {
		params["port_id"] = portId
	}
	result := struct{ InterfaceAttachment nova.InterfaceAttachment }{}
	resp, err := c.R().SetResult(&result).SetBody(ReqBody{"interfaceAttachment": params}).
		Post(id, "os-interface")
	if err != nil {
		return nil, err
	}
	result.InterfaceAttachment.SetRequestId(resp.RequestId())
	return &result.InterfaceAttachment, nil
}
func (c ServerApi) DeleteInterface(id string, portId string) (*session.Response, error) {
	return c.R().Delete(id, "os-interface", portId)
}
func (c ServerApi) DeleteInterfaceAndWait(id string, portId string, timeout time.Duration) error {
	resp, err := c.DeleteInterface(id, portId)
	if err != nil {
		return err
	}
	reqId := resp.RequestId()
	console.Info("[%s] detaching interface %s, request id: %s", id, portId, reqId)

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
			console.Info("[%s] action result: %s", id, action.Events[0].Result)
			if action.Events[0].Result == "Error" {
				return fmt.Errorf("request %s is error", reqId)
			} else {
				return nil
			}
		},
	)
}
func (c ServerApi) doAction(action string, id string, params interface{}, result ...interface{}) (*session.Response, error) {
	req := c.R().SetBody(map[string]interface{}{action: params})
	if len(result) > 0 {
		req.SetResult(result[0])
	}
	return req.Post(id, "action")
}
func (c ServerApi) Stop(id string) error {
	_, err := c.doAction("os-stop", id, nil)
	return err
}
func (c ServerApi) Start(id string) error {
	_, err := c.doAction("os-start", id, nil)
	return err
}
func (c ServerApi) Reboot(id string, hard bool) error {
	body := map[string]string{}
	if hard {
		body["type"] = "hard"
	} else {
		body["type"] = "soft"
	}
	_, err := c.doAction("reboot", id, body)
	return err
}
func (c ServerApi) Pause(id string) error {
	_, err := c.doAction("pause", id, nil)
	return err
}
func (c ServerApi) Unpause(id string) error {
	_, err := c.doAction("unpause", id, nil)
	return err
}
func (c ServerApi) Shelve(id string) error {
	_, err := c.doAction("shelve", id, nil)
	return err
}
func (c ServerApi) Unshelve(id string) error {
	_, err := c.doAction("unshelve", id, nil)
	return err
}
func (c ServerApi) Suspend(id string) error {
	_, err := c.doAction("suspend", id, nil)
	return err
}
func (c ServerApi) Resume(id string) error {
	_, err := c.doAction("resume", id, nil)
	return err
}
func (c ServerApi) Resize(id string, flavorRef string) error {
	_, err := c.doAction("resize", id, map[string]string{"flavorRef": flavorRef})
	return err
}
func (c ServerApi) ResizeConfirm(id string) error {
	_, err := c.doAction("confirmResize", id, nil)
	return err
}
func (c ServerApi) ResizeRevert(id string) error {
	_, err := c.doAction("revertResize", id, nil)
	return err
}

// @param	opt nova.RebuilOpt
//
//	UserData=nil: 删除user data
//	UserData="" : 不指定 user data 参数
//	UserData=非空: 指定 user data 参数
func (c ServerApi) Rebuild(id string, opt nova.RebuilOpt) error {
	options := map[string]interface{}{}
	if opt.ImageId != "" {
		options["imageRef"] = opt.ImageId
	}
	if opt.Password != "" {
		options["adminPass"] = opt.Password
	}
	if opt.Name != "" {
		options["name"] = opt.Name
	}
	if opt.UserData == nil {
		options["user_data"] = nil
	} else if opt.UserData != "" {
		options["user_data"] = opt.UserData
	}
	_, err := c.doAction("rebuild", id, options)
	return err
}
func (c ServerApi) Evacuate(id string, password string, host string, force bool) error {
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
func (c ServerApi) SetPassword(id string, password, user string) error {
	data := map[string]interface{}{
		"adminPass": password,
	}
	if user != "" {
		data["userName"] = user
	}
	_, err := c.doAction("changePassword", id, data)
	return err
}
func (c ServerApi) Set(id string, params map[string]interface{}) error {
	_, err := c.R().SetBody(ReqBody{"server": params}).Put(id)
	return err
}
func (c ServerApi) SetName(id string, name string) error {
	return c.Set(id, map[string]interface{}{"name": name})
}
func (c ServerApi) SetState(id string, active bool) error {
	data := map[string]interface{}{}
	if active {
		data["state"] = "active"
	} else {
		data["state"] = "error"
	}
	_, err := c.doAction("os-resetState", id, data)
	return err
}
func (c ServerApi) ConsoleLog(id string, length uint) (*nova.ConsoleLog, error) {
	params := map[string]interface{}{}
	if length != 0 {
		params["length"] = length
	}
	result := nova.ConsoleLog{}
	_, err := c.doAction("os-getConsoleOutput", id, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func (c ServerApi) getVNCConsole(id string, consoleType string) (*nova.Console, error) {
	params := map[string]interface{}{"type": consoleType}
	result := map[string]*nova.Console{"console": {}}
	_, err := c.doAction("os-getVNCConsole", id, params, &result)
	if err != nil {
		return nil, err
	}

	return result["console"], nil
}
func (c ServerApi) getRemoteConsole(id string, protocol string, consoleType string) (*nova.Console, error) {
	params := map[string]interface{}{
		"remote_console": map[string]interface{}{
			"protocol": protocol,
			"type":     consoleType,
		},
	}
	result := struct {
		RemoteConsole nova.Console `json:"remote_console"`
	}{}
	_, err := c.R().SetBody(params).SetResult(&result).Post(id, "remote-consoles")
	if err != nil {
		return nil, err
	}
	return &result.RemoteConsole, nil
}

func (c ServerApi) ConsoleUrl(id string, consoleType string) (*nova.Console, error) {
	if c.MicroVersionLargeEqual("2.6") {
		// TODO: do not set "vnc" directly
		return c.getRemoteConsole(id, "vnc", consoleType)
	}
	return c.getVNCConsole(id, consoleType)
}
func (c ServerApi) Migrate(id string, host string) error {
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
func (c ServerApi) LiveMigrate(id string, blockMigrate interface{}, host string) error {
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
func (c ServerApi) ListActions(id string) ([]nova.InstanceAction, error) {
	result := struct{ InstanceActions []nova.InstanceAction }{}
	if _, err := c.R().SetResult(&result).Get(id, "os-instance-actions"); err != nil {
		return nil, err
	}
	return result.InstanceActions, nil
}
func (c ServerApi) ShowAction(id, requestId string) (*nova.InstanceAction, error) {
	result := struct{ InstanceAction nova.InstanceAction }{}
	if _, err := c.R().SetResult(&result).Get(id, "os-instance-actions", requestId); err != nil {
		return nil, err
	}
	return &result.InstanceAction, nil
}
func (c ServerApi) ListActionsWithEvents(id string, actionName string, requestId string, last int) ([]nova.InstanceAction, error) {
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
			console.Error("get server action %s failed: %s", action.RequestId, err)
		}
		actionWithEvents = append(actionWithEvents, *serverAction)
	}
	return actionWithEvents, nil
}

func (c ServerApi) ListMigrations(id string, query url.Values) ([]nova.Migration, error) {
	result := struct{ Migrations []nova.Migration }{}
	if _, err := c.R().SetResult(&result).Get(id, "migrations"); err != nil {
		return nil, err
	}
	return result.Migrations, nil

}
func (c ServerApi) RegionLiveMigrate(id string, destRegion string, blockMigrate bool, dryRun bool, destHost string) (*nova.RegionMigrateResp, error) {
	data := map[string]interface{}{
		"region":          destRegion,
		"block_migration": blockMigrate,
		"dry_run":         dryRun,
	}
	if destHost != "" {
		data["host"] = destHost
	}
	result := nova.RegionMigrateResp{}
	_, err := c.doAction("os-migrateLive-region", id, data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func (c ServerApi) CreateImage(id string, imagName string, metadata map[string]string) (string, error) {
	data := map[string]interface{}{
		"name": imagName,
	}
	if len(metadata) > 0 {
		data["metadata"] = metadata
	}
	result := struct {
		ImageId string `json:"image_id"`
	}{}
	resp, err := c.doAction("createImage", id, data, &result)
	if err != nil {
		return "", err
	}
	return result.ImageId, json.Unmarshal(resp.Body(), &result)
}
func (c ServerApi) WaitStatus(serverId string, status string, interval int) (*nova.Server, error) {
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
			console.Info("[server: %s] status: %s, taskState: %s", server.Id, server.Status, server.TaskState)
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

func (c ServerApi) Rename(id string, name string) error {
	_, err := c.doAction("rename", id, map[string]interface{}{"name": name})
	return err
}

func (c ServerApi) WaitBooted(id string) (*nova.Server, error) {
	for {
		server, err := c.Show(id)
		if err != nil {
			return server, err
		}
		console.Info("[%s] %s", id, server.AllStatus())
		if server.IsError() {
			return server, fmt.Errorf("server %s is error", server.Id)
		}
		if server.IsActive() && server.Host != "" {
			return server, nil
		}
		time.Sleep(time.Second * 2)
	}
}
func (c ServerApi) WaitDeleted(id string) error {
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
				console.Info("[%s] %s", id, server.AllStatus())
				return true
			}
			if compare.IsType[session.HttpError](err) {
				httpError, _ := err.(session.HttpError)
				if httpError.IsNotFound() {
					console.Info("[%s] deleted", id)
					err = nil
					return false
				}
			}
			return false
		},
	)
	return err
}
func (c ServerApi) WaitTask(id string, taskState string) (*nova.Server, error) {
	for {
		server, err := c.Show(id)
		if err != nil {
			return nil, err
		}
		console.Info("[%s] %s progress: %d", id, server.AllStatus(), int(server.Progress))

		if strings.ToUpper(server.Status) == "ERROR" {
			return nil, fmt.Errorf("server %s status is ERROR", id)
		}
		if strings.EqualFold(server.TaskState, strings.ToUpper(taskState)) {
			return server, nil
		}
		time.Sleep(time.Second * 2)
	}
}
func (c ServerApi) WaitResized(id string, newFlavorName string) (*nova.Server, error) {
	server, err := c.WaitTask(id, "")
	if err != nil {
		return nil, err
	}
	if server.Flavor.OriginalName == newFlavorName {
		return server, nil
	} else {
		return nil, fmt.Errorf("flavor is not %s", newFlavorName)
	}
}
func (c ServerApi) StopAndWait(id string) error {
	if err := c.Stop(id); err != nil {
		return err
	}
	return utility.RetryWithErrors(
		utility.RetryCondition{
			Timeout:     time.Minute * 30,
			IntervalMin: time.Second * 2},
		[]string{"ServerNotStopped"},
		func() error {
			server, err := c.Show(id)
			if err != nil {
				return fmt.Errorf("get server %s failed: %s", id, err)
			}
			if server.IsStopped() {
				return nil
			}
			return utility.NewServerNotStopped(id)
		},
	)
}
func (c ServerApi) WaitRebooted(id string, newFlavorName string) (*nova.Server, error) {
	server, err := c.WaitTask(id, "")
	if err != nil {
		return nil, err
	}
	return c.WaitStatus(server.Id, "ACTIVE", 5)
}
func (c ServerApi) CreateAndWait(options nova.ServerOpt) (*nova.Server, error) {
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

func (c FlavorApi) List(query url.Values) ([]nova.Flavor, error) {
	return ListResource[nova.Flavor](c.ResourceApi, query)
}
func (c FlavorApi) Detail(query url.Values) ([]nova.Flavor, error) {
	return ListResource[nova.Flavor](c.ResourceApi, query, true)
}
func (c FlavorApi) Show(id string) (*nova.Flavor, error) {
	return ShowResource[nova.Flavor](c.ResourceApi, id)
}
func (c FlavorApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id, nil)
	return err
}

func (c FlavorApi) Find(idOrName string, withExtraSpecs bool) (*nova.Flavor, error) {
	return FindResource(idOrName, c.Show, c.List)
}

func (c FlavorApi) Create(flavor nova.Flavor) (*nova.Flavor, error) {
	result := struct{ Flavor nova.Flavor }{}
	_, err := c.R().SetBody(map[string]nova.Flavor{"flavor": flavor}).SetResult(&result).Post()
	return &result.Flavor, err
}

func (c FlavorApi) ListExtraSpecs(id string) (nova.ExtraSpecs, error) {
	result := struct {
		ExtraSpecs nova.ExtraSpecs `json:"extra_specs"`
	}{}
	_, err := c.R().SetResult(&result).Get(id, "os-extra_specs")
	return result.ExtraSpecs, err
}
func (c FlavorApi) ShowWithExtraSpecs(id string) (*nova.Flavor, error) {
	flavor, err := c.Show(id)
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
	result := struct {
		ExtraSpecs nova.ExtraSpecs `json:"extra_specs"`
	}{}
	_, err := c.R().SetBody(map[string]nova.ExtraSpecs{"extra_specs": extraSpecs}).
		SetResult(&result).Post(id, "os-extra_specs")
	return result.ExtraSpecs, err
}
func (c FlavorApi) DeleteExtraSpec(id string, extraSpec string) error {
	_, err := c.R().Delete(id, "os-extra_specs", extraSpec)
	return err
}
func (c FlavorApi) Copy(id string, newName string, newId string,
	newVcpus int, newRam int, newDisk int, newSwap int,
	newEphemeral int, newRxtxFactor float32, setProperties map[string]string,
	unsetProperties []string,
) (*nova.Flavor, error) {
	console.Info("Show flavor")
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
	console.Info("Show flavor extra specs")
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
	console.Info("Create new flavor")
	newFlavor, err := c.Create(*flavor)
	if err != nil {
		return nil, fmt.Errorf("create flavor failed, %v", err)
	}
	if len(extraSpecs) != 0 {
		console.Info("Set new flavor extra specs")
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
	return ListResource[nova.Hypervisor](c.ResourceApi, query)
}
func (c HypervisorApi) Detail(query url.Values) ([]nova.Hypervisor, error) {
	return ListResource[nova.Hypervisor](c.ResourceApi, query, true)
}
func (c HypervisorApi) ListByName(hostname string) ([]nova.Hypervisor, error) {
	return c.List(url.Values{"hypervisor_hostname_pattern": []string{hostname}})
}

func (c HypervisorApi) Show(id string) (*nova.Hypervisor, error) {
	result := struct {
		Hypervisor nova.Hypervisor `json:"hypervisor"`
	}{}
	resp, err := c.R().SetResult(&result).Get(id)
	if err != nil {
		return nil, err
	}
	result.Hypervisor.SetNumaNodes(resp.Body())
	return &result.Hypervisor, err

}
func (c HypervisorApi) ShowByHostname(hostname string) (*nova.Hypervisor, error) {
	hypervisors, err := c.ListByName(hostname)
	if err != nil {
		return nil, err
	}
	if len(hypervisors) == 0 {
		return nil, fmt.Errorf("hypervisor %s not found", hostname)
	}
	return c.Show(hypervisors[0].Id)
}
func (c HypervisorApi) Find(idOrHostName string) (*nova.Hypervisor, error) {
	if stringutils.IsUUID(idOrHostName) {
		hypervisor, err := c.Show(idOrHostName)
		if err == nil {
			return hypervisor, nil
		}
	}
	return c.ShowByHostname(idOrHostName)
}
func (c HypervisorApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id)
	return err
}
func (c HypervisorApi) Uptime(id string) (*nova.Hypervisor, error) {
	result := struct{ Hypervisor *nova.Hypervisor }{}
	_, err := c.R().SetResult(&result).Get("hypervisors", id, "uptime")
	if err != nil {
		return nil, err
	}
	return result.Hypervisor, nil
}
func (c HypervisorApi) FlavorCapacities(query url.Values) (*nova.FlavorCapacities, error) {
	body := nova.FlavorCapacities{}
	_, err := c.R().SetResult(&body).SetQuery(query).Get("statistics", "flavor-capacities")
	if err != nil {
		return nil, err
	}
	return &body, nil
}

// keypair api
func (c KeypairApi) List(query url.Values) ([]nova.Keypair, error) {
	return ListResource[nova.Keypair](c.ResourceApi, query)
}

// service api
func (c ComputeServiceApi) List(query url.Values) ([]nova.Service, error) {
	return ListResource[nova.Service](c.ResourceApi, query)
}
func (c ComputeServiceApi) ListCompute() ([]nova.Service, error) {
	return c.List(url.Values{"binary": {"nova-compute"}})
}
func (c ComputeServiceApi) GetByHostBinary(host string, binary string) (*nova.Service, error) {
	services, err := c.List(url.Values{"host": {host}, "binary": {binary}})
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

func (c ComputeServiceApi) doAction(action string, params map[string]interface{}) error {
	result := struct{ Service nova.Service }{}
	_, err := c.R().SetResult(&result).SetBody(params).Put(action)
	return err
}
func (c ComputeServiceApi) update(id string, update map[string]interface{}) (*nova.Service, error) {
	result := struct{ Service nova.Service }{}
	_, err := c.R().SetBody(update).SetResult(&result).Put(id)
	if err != nil {
		return nil, err
	}
	return &result.Service, nil
}
func (c ComputeServiceApi) Up(host string, binary string) (*nova.Service, error) {
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
func (c ComputeServiceApi) Down(host string, binary string) (*nova.Service, error) {
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
func (c ComputeServiceApi) Enable(host string, binary string) (*nova.Service, error) {
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
		fmt.Println(body)
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
func (c ComputeServiceApi) Delete(host string, binary string) error {
	service, err := c.GetByHostBinary(host, binary)
	if err != nil {
		return err
	}
	_, err = DeleteResource(c.ResourceApi, service.Id)
	return err
}

// migration api
func (c MigrationApi) List(query url.Values) ([]nova.Migration, error) {
	return ListResource[nova.Migration](c.ResourceApi, query)
}

// avaliable zone api
func (c AZApi) List(query url.Values) ([]nova.AvailabilityZone, error) {
	result := struct{ AvailabilityZoneInfo []nova.AvailabilityZone }{}
	if _, err := c.R().SetQuery(query).SetResult(&result).Get(); err != nil {
		return nil, err
	}
	return result.AvailabilityZoneInfo, nil
}
func (c AZApi) Detail(query url.Values) ([]nova.AvailabilityZone, error) {
	result := struct{ AvailabilityZoneInfo []nova.AvailabilityZone }{}
	if _, err := c.R().SetQuery(query).SetResult(&result).Get("detail"); err != nil {
		return nil, err
	}
	return result.AvailabilityZoneInfo, nil
}

// aggregate api

func (c AggregateApi) List(query url.Values) ([]nova.Aggregate, error) {
	return ListResource[nova.Aggregate](c.ResourceApi, query)
}
func (c AggregateApi) Show(id string) (*nova.Aggregate, error) {
	return ShowResource[nova.Aggregate](c.ResourceApi, id)
}
func (c AggregateApi) Create(agg nova.Aggregate) (*nova.Aggregate, error) {
	result := struct {
		Aggregate nova.Aggregate `json:"aggregate"`
	}{}
	if _, err := c.R().SetBody(map[string]nova.Aggregate{"aggregate": agg}).
		SetResult(&result).Post(); err != nil {
		return nil, err
	} else {
		return &result.Aggregate, nil
	}
}

func (c AggregateApi) Delete(id int) error {
	if _, err := DeleteResource(c.ResourceApi, strconv.Itoa(id)); err != nil {
		return err
	}
	return nil
}
func (c AggregateApi) Find(idOrName string) (*nova.Aggregate, error) {
	agg, err := c.Show(idOrName)
	if err == nil {
		return agg, nil
	}
	if !compare.IsType[session.HttpError](err) {
		return nil, err
	}
	httpError, _ := err.(session.HttpError)
	if !httpError.IsNotFound() {
		return nil, err
	}
	aggs, err := c.List(nil)
	if err != nil {
		return nil, err
	}
	aggs = utility.Filter(aggs, func(x nova.Aggregate) bool {
		return x.Name == idOrName
	})
	switch len(aggs) {
	case 0:
		return nil, fmt.Errorf("aggregate %s not found", idOrName)
	case 1:
		return &aggs[0], nil
	default:
		return nil, fmt.Errorf("found multi aggregates with name %s", idOrName)
	}
}
func (c AggregateApi) AddHost(id int, host string) (*nova.Aggregate, error) {
	body := struct {
		AddHost map[string]string `json:"add_host"`
	}{
		AddHost: map[string]string{"host": host},
	}
	result := struct{ Aggregate nova.Aggregate }{}
	if _, err := c.R().SetResult(&result).SetBody(body).
		Post(strconv.Itoa(id), "action"); err != nil {
		return nil, err
	}
	return &result.Aggregate, nil
}
func (c AggregateApi) RemoveHost(id int, host string) (*nova.Aggregate, error) {
	body := struct {
		RemoveHost map[string]string `json:"remove_host"`
	}{
		RemoveHost: map[string]string{"host": host},
	}
	result := struct{ Aggregate nova.Aggregate }{}
	if _, err := c.R().SetResult(&result).SetBody(body).
		Post(strconv.Itoa(id), "action"); err != nil {
		return nil, err
	}
	return &result.Aggregate, nil
}

// server group api

func (c ServerGroupApi) List(query url.Values) ([]nova.ServerGroup, error) {
	return ListResource[nova.ServerGroup](c.ResourceApi, query)
}

// quota api

func (c ComputeQuotaApi) Show(projectId string) (*nova.QuotaSet, error) {
	result := struct {
		QuotaSet nova.QuotaSet `json:"quota_set"`
	}{}
	if _, err := c.R().SetResult(&result).Get(projectId); err != nil {
		return nil, err
	}
	return &result.QuotaSet, nil
}
