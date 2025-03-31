package internal

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/BytemanD/skyman/utility"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
)

const X_OPENSTACK_NOVA_API_VERSION = "X-Openstack-Nova-Api-Version"

type microVersion struct {
	Version      int
	MicroVersion int
}

func (v microVersion) Compare(other microVersion) int {
	if v.Version != other.Version {
		return v.Version - other.Version
	}
	return v.MicroVersion - other.MicroVersion
}

type NovaV2 struct {
	*ServiceClient
	MicroVersion *model.ApiVersion
}

func getMicroVersion(vertionStr string) microVersion {
	versionList := strings.Split(vertionStr, ".")
	v, _ := strconv.Atoi(versionList[0])
	micro, _ := strconv.Atoi(versionList[1])
	return microVersion{Version: v, MicroVersion: micro}
}

func (c *NovaV2) GetMicroVersion() string {
	return c.Header().Get(X_OPENSTACK_NOVA_API_VERSION)
}
func (c *NovaV2) MicroVersionLargeEqual(version string) bool {
	clientMicroVersoin := c.GetMicroVersion()
	return getMicroVersion(clientMicroVersoin).
		Compare(getMicroVersion(version)) >= 0
}
func (c *NovaV2) DiscoverMicroVersion() error {
	console.Debug("discorver micro version for nova")
	result := struct{ Versions model.ApiVersions }{}
	if _, err := c.Index(&result); err != nil {
		return err
	} else {
		current := result.Versions.Current()
		if current == nil {
			return fmt.Errorf("current version not found")
		}
		console.Debug("use micro version: %s", current.Version)
		c.SetHeader(X_OPENSTACK_NOVA_API_VERSION, current.Version)
	}
	return nil
}
func (c *NovaV2) String() string {
	return fmt.Sprintf("<Compute: %s>", c.BaserUrl())
}

// server api

func (c NovaV2) ListServer(query url.Values, details ...bool) ([]nova.Server, error) {
	u := URL_SERVERS
	if lo.CoalesceOrEmpty(details...) {
		u = URL_SERVERS_DETAIL
	}
	return QueryResource[nova.Server](c.ServiceClient, u.F(), query, SERVERS)
}
func (c NovaV2) GetServer(id string) (*nova.Server, error) {
	result := struct {
		Server nova.Server `json:"server"`
	}{}
	_, err := c.R().SetResult(&result).Get(URL_SERVER.F(id))
	return &result.Server, err
}

func (c NovaV2) FindServer(idOrName string, allTenants ...bool) (*nova.Server, error) {
	allTenant := lo.FirstOrEmpty(allTenants)
	return QueryByIdOrName(
		idOrName, c.GetServer,
		func(query url.Values) ([]nova.Server, error) {
			if allTenant {
				query.Set("all_tenants", "1")
			}
			return c.ListServer(query, true)
		})
}
func (c NovaV2) CreateServer(options nova.ServerOpt) (*nova.Server, error) {
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
			SetResult(&body).Post(URL_SERVER_VOLUMES_BOOT.F())
	} else {
		_, err = c.R().SetBody(map[string]nova.ServerOpt{"server": options}).
			SetResult(&body).Post(URL_SERVERS.F())
	}
	if err != nil {
		return nil, err
	}
	return &body.Server, nil
}
func (c NovaV2) DeleteServer(id string) error {
	return DeleteResource(c.ServiceClient, URL_SERVER.F(id))
}

// server volume api

func (c NovaV2) ListServerVolumes(id string) ([]nova.VolumeAttachment, error) {
	body := struct{ VolumeAttachments []nova.VolumeAttachment }{}
	_, err := c.R().SetResult(&body).Get(URL_SERVER_VOLUMES.F(id))
	return body.VolumeAttachments, err
}

func (c NovaV2) ServerAddVolume(id, volumeId string) (*nova.VolumeAttachment, error) {
	body := struct{ VolumeAttachment nova.VolumeAttachment }{}
	_, err := c.R().SetResult(&body).
		SetBody(ReqBody{"volumeAttachment": {"volumeId": volumeId}}).
		Post(string(URL_SERVER_VOLUMES.F(id)))
	if err != nil {
		return nil, err
	}
	return &body.VolumeAttachment, err
}

func (c NovaV2) ServerDeleteVolume(id, volumeId string) error {
	return DeleteResource(c.ServiceClient, URL_SERVER_VOLUMES.F(id))
}

// server interface api

func (c NovaV2) ListServerInterfaces(id string) ([]nova.InterfaceAttachment, error) {
	body := struct{ InterfaceAttachments []nova.InterfaceAttachment }{}
	_, err := c.R().SetResult(&body).Get(URL_SERVER_INTERFACES.F(id))
	return body.InterfaceAttachments, err
}

func (c NovaV2) ServerAddInterface(id, netId, portId string) (*nova.InterfaceAttachment, error) {
	params := map[string]any{}
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
	resp, err := c.R().SetResult(&result).
		SetBody(ReqBody{"interfaceAttachment": params}).
		Post(URL_SERVER_INTERFACES.F(id))
	if err != nil {
		return nil, err
	}
	result.InterfaceAttachment.SetRequestId(
		resp.Header().Get(session.HEADER_REQUEST_ID))
	return &result.InterfaceAttachment, nil
}
func (c NovaV2) ServerDeleteInterface(id, portId string) (*resty.Response, error) {
	return DeleteResourceWithResp(c.ServiceClient, URL_SERVER_INTERFACE.F(id, portId))
}

// server actions

func (c NovaV2) serverDoAction(action, id string, params any, result ...any) (*resty.Response, error) {
	req := c.R().SetBody(map[string]any{action: params})
	if len(result) > 0 {
		req.SetResult(result[0])
	}
	return req.Post(URL_SERVER_ACTION.F(id))
}

func (c NovaV2) StopServer(id string) error {
	_, err := c.serverDoAction("os-stop", id, nil)
	return err
}

func (c NovaV2) StartServer(id string) error {
	_, err := c.serverDoAction("os-start", id, nil)
	return err
}
func (c NovaV2) RebootServer(id string, hard bool) error {
	body := map[string]string{}
	if hard {
		body["type"] = "hard"
	} else {
		body["type"] = "soft"
	}
	_, err := c.serverDoAction("reboot", id, body)
	return err
}
func (c NovaV2) PauseServer(id string) error {
	_, err := c.serverDoAction("pause", id, nil)
	return err
}
func (c NovaV2) UnpauseServer(id string) error {
	_, err := c.serverDoAction("unpause", id, nil)
	return err
}
func (c NovaV2) ShelveServer(id string) error {
	_, err := c.serverDoAction("shelve", id, nil)
	return err
}
func (c NovaV2) UnshelveServer(id string) error {
	_, err := c.serverDoAction("unshelve", id, nil)
	return err
}
func (c NovaV2) SuspendServer(id string) error {
	_, err := c.serverDoAction("suspend", id, nil)
	return err
}
func (c NovaV2) ResumeServer(id string) error {
	_, err := c.serverDoAction("resume", id, nil)
	return err
}
func (c NovaV2) ResizeServer(id string, flavorRef string) error {
	_, err := c.serverDoAction("resize", id, map[string]string{"flavorRef": flavorRef})
	return err
}
func (c NovaV2) ResizeConfirm(id string) error {
	_, err := c.serverDoAction("confirmResize", id, nil)
	return err
}
func (c NovaV2) ResizeRevert(id string) error {
	_, err := c.serverDoAction("revertResize", id, nil)
	return err
}

func (c NovaV2) RebuildServer(id string, opt nova.RebuilOpt) error {
	options := map[string]any{}
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
	_, err := c.serverDoAction("rebuild", id, options)
	return err
}
func (c NovaV2) EvacuateServer(id string, password string, host string, force bool) error {
	data := map[string]any{}
	if password != "" {
		data["password"] = password
	}
	if host != "" {
		data["host"] = password
	}
	if force {
		data["force"] = force
	}
	_, err := c.serverDoAction("evacuate", id, data)
	return err
}
func (c NovaV2) SetServerPassword(id string, password, user string) error {
	data := map[string]any{"adminPass": password}
	if user != "" {
		data["userName"] = user
	}
	_, err := c.serverDoAction("changePassword", id, data)
	return err
}
func (c NovaV2) SetServerState(id string, active bool) error {
	data := map[string]any{}
	if active {
		data["state"] = "active"
	} else {
		data["state"] = "error"
	}
	_, err := c.serverDoAction("os-resetState", id, data)
	return err
}

func (c NovaV2) GetServerConsoleLog(id string, length uint) (*nova.ConsoleLog, error) {
	params := map[string]any{}
	if length != 0 {
		params["length"] = length
	}
	result := nova.ConsoleLog{}
	_, err := c.serverDoAction("os-getConsoleOutput", id, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func (c NovaV2) getVNCConsole(id string, consoleType string) (*nova.Console, error) {
	params := map[string]any{"type": consoleType}
	result := map[string]*nova.Console{"console": {}}
	_, err := c.serverDoAction("os-getVNCConsole", id, params, &result)
	if err != nil {
		return nil, err
	}
	return result["console"], nil
}
func (c NovaV2) getRemoteConsole(id string, protocol string, consoleType string) (*nova.Console, error) {
	params := map[string]any{
		"remote_console": map[string]any{
			"protocol": protocol,
			"type":     consoleType,
		},
	}
	result := struct {
		RemoteConsole nova.Console `json:"remote_console"`
	}{}
	_, err := c.R().SetBody(params).SetResult(&result).
		Post(URL_SERVERS_REMOTE_CONSOLES.F(id))
	if err != nil {
		return nil, err
	}
	return &result.RemoteConsole, nil
}
func (c NovaV2) GetServerConsoleUrl(id string, consoleType string) (*nova.Console, error) {
	if c.MicroVersionLargeEqual("2.6") {
		// TODO: do not set "vnc" directly
		return c.getRemoteConsole(id, "vnc", consoleType)
	}
	return c.getVNCConsole(id, consoleType)
}

func (c NovaV2) ServerMigrate(id string, host string) error {
	data := map[string]any{}
	if host != "" {
		server, err := c.GetServer(id)
		if err != nil {
			return err
		}
		if server.Host == host {
			return fmt.Errorf("server %s is on host %s now", id, server.Host)
		}
		data["host"] = host
	} else {
		data["host"] = nil
	}
	_, err := c.serverDoAction("migrate", id, data)
	return err
}
func (c NovaV2) ServerLiveMigrate(id string, blockMigrate any, host string) error {
	data := map[string]any{"block_migration": blockMigrate}
	if host != "" {
		server, err := c.GetServer(id)
		if err != nil {
			return err
		}
		if server.Host == host {
			return fmt.Errorf("server %s is on host %s now", id, server.Host)
		}
		data["host"] = host
	} else {
		data["host"] = nil
	}
	_, err := c.serverDoAction("os-migrateLive", id, data)
	return err
}
func (c NovaV2) ServerRegionLiveMigrate(id string, destRegion string, blockMigrate bool, dryRun bool, destHost string) (*nova.RegionMigrateResp, error) {
	data := map[string]any{
		"region":          destRegion,
		"block_migration": blockMigrate,
		"dry_run":         dryRun,
	}
	if destHost != "" {
		data["host"] = destHost
	}
	result := nova.RegionMigrateResp{}
	_, err := c.serverDoAction("os-migrateLive-region", id, data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func (c NovaV2) ServerCreateImage(id string, imagName string, metadata map[string]string) (string, error) {
	data := map[string]any{
		"name": imagName,
	}
	if len(metadata) > 0 {
		data["metadata"] = metadata
	}
	result := struct {
		ImageId string `json:"image_id"`
	}{}
	_, err := c.serverDoAction("createImage", id, data, &result)
	if err != nil {
		return "", err
	}
	return result.ImageId, nil
}
func (c NovaV2) ServerRename(id string, name string) error {
	_, err := c.serverDoAction("rename", id, map[string]any{"name": name})
	return err
}

// server update api

func (c NovaV2) SetServer(id string, params map[string]any) error {
	_, err := c.R().SetBody(ReqBody{"server": params}).Put(URL_SERVER.F(id))
	return err
}
func (c NovaV2) ServerSetName(id string, name string) error {
	return c.SetServer(id, map[string]any{"name": name})
}

// server actions api

func (c NovaV2) ListServerActions(id string) ([]nova.InstanceAction, error) {
	result := struct{ InstanceActions []nova.InstanceAction }{}
	if _, err := c.R().SetResult(&result).
		Get(URL_SERVER_INSTANCE_ACTIONS.F(id)); err != nil {
		return nil, err
	}
	return result.InstanceActions, nil
}
func (c NovaV2) GetServerAction(id, requestId string) (*nova.InstanceAction, error) {
	result := struct{ InstanceAction nova.InstanceAction }{}
	if _, err := c.R().SetResult(&result).
		Get(URL_SERVER_INSTANCE_ACTION.F(id, requestId)); err != nil {
		return nil, err
	}
	return &result.InstanceAction, nil
}
func (c NovaV2) ListServerActionsWithEvents(id string, actionName string, requestId string, last int) ([]nova.InstanceAction, error) {
	serverActions, err := c.ListServerActions(id)
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
		serverAction, err := c.GetServerAction(id, action.RequestId)
		if err != nil {
			console.Error("get server action %s failed: %s", action.RequestId, err)
		}
		actionWithEvents = append(actionWithEvents, *serverAction)
	}
	return actionWithEvents, nil
}

// server migration api

func (c NovaV2) ListServerMigrations(id string, query url.Values) ([]nova.Migration, error) {
	result := struct{ Migrations []nova.Migration }{}
	if _, err := c.R().SetResult(&result).
		Get(URL_SERVER_MIGRATIONS.F(id)); err != nil {
		return nil, err
	}
	return result.Migrations, nil
}

// flavor api

func (c NovaV2) ListFlavors(query url.Values, details ...bool) ([]nova.Flavor, error) {
	u := URL_FLAVORS
	if lo.CoalesceOrEmpty(details...) {
		u = URL_FLAVORS_DETAIL
	}
	return QueryResource[nova.Flavor](c.ServiceClient, u.F(), query, "flavors")
}
func (c NovaV2) GetFlavor(id string) (*nova.Flavor, error) {
	return GetResource[nova.Flavor](c.ServiceClient, URL_FLAVOR.F(id), "flavor")
}
func (c NovaV2) GetFlavorExtraSpecs(id string) (nova.ExtraSpecs, error) {
	result := struct {
		ExtraSpecs nova.ExtraSpecs `json:"extra_specs"`
	}{}
	_, err := c.R().SetResult(&result).Get(URL_FLAVOR_EXTRA_SPECS.F(id))
	return result.ExtraSpecs, err
}
func (c NovaV2) GetFlavorWithExtraSpecs(id string) (*nova.Flavor, error) {
	flavor, err := c.GetFlavor(id)
	if err != nil {
		return nil, err
	}
	extraSpecs, err := c.GetFlavorExtraSpecs(flavor.Id)
	if err != nil {
		return nil, err
	}
	flavor.ExtraSpecs = extraSpecs
	return flavor, err
}
func (c NovaV2) FindFlavor(idOrName string) (*nova.Flavor, error) {
	return QueryByIdOrName(
		idOrName, c.GetFlavorWithExtraSpecs,
		func(query url.Values) ([]nova.Flavor, error) {
			return c.ListFlavors(query, true)
		})
}
func (c NovaV2) DeleteFlavor(id string) error {
	return DeleteResource(c.ServiceClient, URL_FLAVOR.F(id))
}
func (c NovaV2) CreateFlavor(flavor nova.Flavor) (*nova.Flavor, error) {
	result := struct{ Flavor nova.Flavor }{}
	_, err := c.R().SetBody(map[string]nova.Flavor{"flavor": flavor}).
		SetResult(&result).Post(URL_FLAVORS.F())
	return &result.Flavor, err
}
func (c NovaV2) SetFlavorExtraSpecs(id string, extraSpecs map[string]string) (nova.ExtraSpecs, error) {
	result := struct {
		ExtraSpecs nova.ExtraSpecs `json:"extra_specs"`
	}{}
	_, err := c.R().SetBody(map[string]nova.ExtraSpecs{"extra_specs": extraSpecs}).
		SetResult(&result).Post(URL_FLAVOR_EXTRA_SPECS.F(id))
	return result.ExtraSpecs, err
}
func (c NovaV2) DeleteFlavorExtraSpec(id string, extraSpec string) error {
	err := DeleteResource(c.ServiceClient, URL_FLAVOR_EXTRA_SPEC.F(id, extraSpec))
	if errors.Is(err, session.ErrHTTP404) {
		console.Warn("flavor %s dosen't has extra spec: %s", id, extraSpec)
		return nil
	}
	return err
}

// hypervisor api

func (c NovaV2) ListHypervisor(query url.Values, details ...bool) ([]nova.Hypervisor, error) {
	url := URL_HYPERVISORS
	if lo.FirstOrEmpty(details) {
		url = URL_HYPERVISORS_DETAIL
	}
	return QueryResource[nova.Hypervisor](c.ServiceClient, url.F(), query, "hypervisors")
}
func (c NovaV2) Detail(query url.Values) ([]nova.Hypervisor, error) {
	return QueryResource[nova.Hypervisor](
		c.ServiceClient, URL_HYPERVISORS_DETAIL.F(), query, "hypervisors")
}
func (c NovaV2) ListByName(hostname string) ([]nova.Hypervisor, error) {
	return c.ListHypervisor(url.Values{"hypervisor_hostname_pattern": []string{hostname}})
}
func (c NovaV2) GetHypervisor(id string) (*nova.Hypervisor, error) {
	result := struct {
		Hypervisor nova.Hypervisor `json:"hypervisor"`
	}{}
	resp, err := c.R().SetResult(&result).Get(URL_HYPERVISOR.F(id))
	if err != nil {
		return nil, err
	}
	result.Hypervisor.SetNumaNodes(resp.Body())
	return &result.Hypervisor, err

}
func (c NovaV2) GetHypervisorByHostname(hostname string) (*nova.Hypervisor, error) {
	hypervisors, err := c.ListByName(hostname)
	if err != nil {
		return nil, err
	}
	if len(hypervisors) == 0 {
		return nil, fmt.Errorf("hypervisor %s not found", hostname)
	}
	return c.GetHypervisor(hypervisors[0].Id)
}
func (c NovaV2) FindHypervisor(idOrHostName string) (*nova.Hypervisor, error) {
	if stringutils.IsUUID(idOrHostName) {
		hypervisor, err := c.GetHypervisor(idOrHostName)
		if err == nil {
			return hypervisor, nil
		}
	}
	return c.GetHypervisorByHostname(idOrHostName)
}
func (c NovaV2) Delete(id string) error {
	return DeleteResource(c.ServiceClient, URL_HYPERVISOR.F(id))
}
func (c NovaV2) GetHypervisorUptime(id string) (*nova.Hypervisor, error) {
	result := struct{ Hypervisor *nova.Hypervisor }{}
	_, err := c.R().SetResult(&result).Get(URL_HYPERVISOR_UPTIME.F(id))
	if err != nil {
		return nil, err
	}
	return result.Hypervisor, nil
}
func (c NovaV2) GetHypervisorFlavorCapacities(query url.Values) (*nova.FlavorCapacities, error) {
	body := nova.FlavorCapacities{}
	_, err := c.R().SetResult(&body).SetQueryParamsFromValues(query).
		Get(URL_HYPERVISOR_CAPACITIES.F())
	if err != nil {
		return nil, err
	}
	return &body, nil
}

// keypair api

func (c NovaV2) ListKeypair(query url.Values) ([]nova.Keypair, error) {
	return QueryResource[nova.Keypair](
		c.ServiceClient, URL_KEYPAIRS.F(), query, "keyapirs")
}
func (c NovaV2) GetKeypair(name string) (*nova.Keypair, error) {
	return GetResource[nova.Keypair](
		c.ServiceClient, URL_KEYPAIR.F(), "keyapirs",
	)
}
func (c NovaV2) CreateKeypair(name string, keyType string, opt nova.KeypairOpt) (*nova.Keypair, error) {
	result := nova.Keypair{}

	keypairOption := map[string]string{"name": name, "type": keyType}
	if opt.PublicKey != "" {
		keypairOption["public_key"] = opt.PublicKey
	}
	if opt.UserId != "" {
		keypairOption["user_id"] = opt.UserId
	}
	if _, err := c.R().SetBody(map[string]map[string]string{"keypair": keypairOption}).
		SetResult(&result).Post(URL_KEYPAIRS.F()); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}
func (c NovaV2) DeleteKeypair(name string) error {
	return DeleteResource(c.ServiceClient, URL_KEYPAIR.F(name))
}

// service api
func (c NovaV2) ListService(query url.Values) ([]nova.Service, error) {
	return QueryResource[nova.Service](
		c.ServiceClient, URL_COMPUTE_SERVICES.F(), query, "services")
}
func (c NovaV2) ListComputeService() ([]nova.Service, error) {
	return c.ListService(url.Values{"binary": {"nova-compute"}})
}
func (c NovaV2) GetByHostBinary(host string, binary string) (*nova.Service, error) {
	services, err := c.ListService(url.Values{"host": {host}, "binary": {binary}})
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

func (c NovaV2) doServiceAction(action string, params map[string]any) error {
	result := struct{ Service nova.Service }{}
	_, err := c.R().SetResult(&result).SetBody(params).Put(URL_COMPUTE_SERVICE_ACTION.F(action))
	return err
}
func (c NovaV2) update(id string, update map[string]any) (*nova.Service, error) {
	result := struct{ Service nova.Service }{}
	_, err := c.R().SetBody(update).SetResult(&result).Put(URL_COMPUTE_SERVICE.F(id))
	if err != nil {
		return nil, err
	}
	return &result.Service, nil
}
func (c NovaV2) UpService(host string, binary string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return c.update(service.Id, map[string]any{"forced_down": false})
	}
	err := c.doServiceAction("force-down", map[string]any{
		"host": host, "binary": binary, "forced_down": false,
	})
	if err != nil {
		return nil, err
	} else {
		return c.GetByHostBinary(host, binary)
	}
}
func (c NovaV2) DownService(host string, binary string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return c.update(service.Id, map[string]any{"forced_down": true})
	}
	err := c.doServiceAction("force-down", map[string]any{
		"host": host, "binary": binary, "forced_down": true,
	})
	if err != nil {
		return nil, err
	} else {
		return c.GetByHostBinary(host, binary)
	}
}
func (c NovaV2) EnableService(host string, binary string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		return c.update(service.Id, map[string]any{"status": "enabled"})
	}
	err := c.doServiceAction("enable", map[string]any{
		"host": host, "binary": binary, "status": "enabled",
	})
	if err != nil {
		return nil, err
	} else {
		return c.GetByHostBinary(host, binary)
	}
}
func (c NovaV2) DisableService(host string, binary string, reason string) (*nova.Service, error) {
	if c.MicroVersionLargeEqual("2.53") {
		service, err := c.GetByHostBinary(host, binary)
		if err != nil {
			return nil, err
		}
		body := map[string]any{"status": "disabled"}
		if reason != "" {
			body["disabled_reason"] = reason
		}
		return c.update(service.Id, body)
	}
	err := c.doServiceAction("disable", map[string]any{
		"host": host, "binary": binary, "status": "disabled",
	})
	if err != nil {
		return nil, err
	} else {
		return c.GetByHostBinary(host, binary)
	}
}
func (c NovaV2) DeleteService(host string, binary string) error {
	service, err := c.GetByHostBinary(host, binary)
	if err != nil {
		return err
	}
	return DeleteResource(c.ServiceClient, URL_COMPUTE_SERVICE.F(service.Id))
}

// migration api

func (c NovaV2) ListMigration(query url.Values) ([]nova.Migration, error) {
	return QueryResource[nova.Migration](
		c.ServiceClient, URL_MIGRATIONS_LIST.F(), query, "migrations")
}

// avaliable zone api

func (c NovaV2) ListAZ(query url.Values, detail ...bool) ([]nova.AvailabilityZone, error) {
	url := URL_AVAILABILITY_ZONES
	if lo.CoalesceOrEmpty(detail...) {
		url = URL_AVAILABILITY_ZONES_DETAIL
	}
	return QueryResource[nova.AvailabilityZone](
		c.ServiceClient, url.F(), c.Client.QueryParam, "availability-zone",
	)
}

// aggregate api

func (c NovaV2) ListAgg(query url.Values) ([]nova.Aggregate, error) {
	return QueryResource[nova.Aggregate](
		c.ServiceClient, URL_AGGREGATES.F(), query, "aggregates",
	)
}
func (c NovaV2) GetAgg(id string) (*nova.Aggregate, error) {
	return GetResource[nova.Aggregate](c.ServiceClient, URL_AGGREGATE.F(), "aggregate")
}
func (c NovaV2) CreateAgg(agg nova.Aggregate) (*nova.Aggregate, error) {
	result := struct {
		Aggregate nova.Aggregate `json:"aggregate"`
	}{}
	if _, err := c.R().SetBody(map[string]nova.Aggregate{"aggregate": agg}).
		SetResult(&result).Post(URL_AGGREGATES.F()); err != nil {
		return nil, err
	} else {
		return &result.Aggregate, nil
	}
}
func (c NovaV2) DeleteAgg(id int) error {
	return DeleteResource(c.ServiceClient, URL_AGGREGATE.F(id))
}
func (c NovaV2) FindAgg(idOrName string) (*nova.Aggregate, error) {
	agg, err := c.GetAgg(idOrName)
	if err == nil {
		return agg, nil
	}
	if !errors.Is(err, session.ErrHTTP404) {
		return nil, err
	}
	aggs, err := c.ListAgg(nil)
	if err != nil {
		return nil, err
	}
	aggs = lo.Filter(aggs, func(x nova.Aggregate, index int) bool {
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

func (c NovaV2) AggAddHost(id int, host string) (*nova.Aggregate, error) {
	body := struct {
		AddHost map[string]string `json:"add_host"`
	}{
		AddHost: map[string]string{"host": host},
	}
	result := struct{ Aggregate nova.Aggregate }{}
	if _, err := c.R().SetResult(&result).SetBody(body).
		Post(URL_AGGREGATE_ACTION.F(id)); err != nil {
		return nil, err
	}
	return &result.Aggregate, nil
}
func (c NovaV2) AggRemoveHost(id int, host string) (*nova.Aggregate, error) {
	body := struct {
		RemoveHost map[string]string `json:"remove_host"`
	}{
		RemoveHost: map[string]string{"host": host},
	}
	result := struct{ Aggregate nova.Aggregate }{}
	if _, err := c.R().SetResult(&result).SetBody(body).
		Post(URL_AGGREGATE_ACTION.F(id)); err != nil {
		return nil, err
	}
	return &result.Aggregate, nil
}

// server group api

func (c NovaV2) ListServerGroup(query url.Values) ([]nova.ServerGroup, error) {
	return QueryResource[nova.ServerGroup](
		c.ServiceClient, URL_SERVER_GROUPS.F(), query, "server_groups")
}

// quota api

func (c NovaV2) GetQuotaSet(projectId string) (*nova.QuotaSet, error) {
	return GetResource[nova.QuotaSet](
		c.ServiceClient, URL_QUOTA_DETAIL.F(projectId), "quota_set",
	)
}

// 扩展的方法

func (c NovaV2) WaitServerStatus(serverId string, status string, interval int) (*nova.Server, error) {
	var server *nova.Server
	err := utility.RetryWithError(
		utility.RetryCondition{
			Timeout:     time.Second * 60 * 10,
			IntervalMin: time.Second * time.Duration(interval),
		},
		ErrServerIsError,
		func() error {
			server, err := c.GetServer(serverId)
			if err != nil {
				return nil
			}
			console.Info("[server: %s] status: %s, taskState: %s", server.Id, server.Status, server.TaskState)
			if strings.EqualFold(server.Status, "ERROR") {
				return fmt.Errorf("%w, message: %s", ErrServerIsError, server.Fault.Message)
			}
			if strings.EqualFold(server.Status, status) {
				return nil
			}
			return fmt.Errorf("server status is %s, but want: %s", server.Status, status)
		},
	)
	return server, err
}

func (c NovaV2) WaitServerBooted(id string) (*nova.Server, error) {
	for {
		server, err := c.GetServer(id)
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
func (c NovaV2) WaitServerDeleted(id string) error {
	return utility.RetryWithError(
		utility.RetryCondition{
			Timeout:     time.Second * 60 * 10,
			IntervalMin: time.Second * time.Duration(2)},
		ErrServerIsNotDeleted,
		func() error {
			server, err := c.GetServer(id)
			if err == nil {
				console.Info("[%s] %s", id, server.AllStatus())
				return ErrServerIsNotDeleted
			}
			if errors.Is(err, session.ErrHTTP404) {
				console.Info("[%s] deleted", id)
				return nil
			} else {
				return err
			}
		},
	)
}
func (c NovaV2) WaitServerTask(id string, taskState string) (*nova.Server, error) {
	for {
		server, err := c.GetServer(id)
		if err != nil {
			return nil, err
		}
		console.Info("[%s] %s progress: %d", id, server.AllStatus(), int(server.Progress))

		if strings.EqualFold(server.Status, "ERROR") {
			return nil, fmt.Errorf("server %s status is ERROR", id)
		}
		if strings.EqualFold(server.TaskState, taskState) {
			return server, nil
		}
		time.Sleep(time.Second * 2)
	}
}
func (c NovaV2) WaitServerResized(id string, newFlavorName string) (*nova.Server, error) {
	server, err := c.WaitServerTask(id, "")
	if err != nil {
		return nil, err
	}
	if server.Flavor.OriginalName == newFlavorName {
		return server, nil
	} else {
		return nil, fmt.Errorf("flavor is not %s", newFlavorName)
	}
}
func (c NovaV2) StopServerAndWait(id string) error {
	if err := c.StopServer(id); err != nil {
		return err
	}
	return utility.RetryWithError(
		utility.RetryCondition{
			Timeout:     time.Minute * 30,
			IntervalMin: time.Second * 2},
		ErrServerNotStopped,
		func() error {
			server, err := c.GetServer(id)
			if err != nil {
				return fmt.Errorf("get server %s failed: %s", id, err)
			}
			if server.IsStopped() {
				return nil
			}
			return fmt.Errorf("%w: %s", ErrServerNotStopped, id)
		},
	)
}
func (c NovaV2) WaitServerRebooted(id string, newFlavorName string) (*nova.Server, error) {
	server, err := c.WaitServerTask(id, "")
	if err != nil {
		return nil, err
	}
	return c.WaitServerStatus(server.Id, "ACTIVE", 5)
}
func (c NovaV2) CreateServerAndWait(options nova.ServerOpt) (*nova.Server, error) {
	server, err := c.CreateServer(options)
	if err != nil {
		return server, err
	}
	if server.Id == "" {
		return server, fmt.Errorf("create server failed")
	}
	server, err = c.WaitServerTask(server.Id, "")
	if err != nil {
		return server, err
	}
	return c.WaitServerBooted(server.Id)
}
func (c NovaV2) DeleteServerInterfaceAndWait(id string, portId string, timeout time.Duration) error {
	resp, err := c.ServerDeleteInterface(id, portId)
	if err != nil {
		return err
	}
	reqId := resp.Header().Get(session.HEADER_REQUEST_ID)
	console.Info("[%s] detaching interface %s, request id: %s", id, portId, reqId)
	return utility.RetryWithError(
		utility.RetryCondition{
			Timeout:     timeout,
			IntervalMin: time.Second * 2},
		ErrActionNotFinish,
		func() error {
			action, err := c.GetServerAction(id, reqId)
			if err != nil {
				return fmt.Errorf("get action events failed: %s", err)
			}
			if len(action.Events) == 0 || action.Events[0].FinishTime == "" {
				return fmt.Errorf("%w: %s", ErrActionNotFinish, reqId)
			}
			console.Info("[%s] action result: %s", id, action.Events[0].Result)
			if action.Events[0].Result == "Error" {
				return fmt.Errorf("%w: request id is %s", ErrActionFailed, reqId)
			} else {
				return nil
			}
		},
	)
}

func (c NovaV2) DeleteServerVolumeAndWait(id string, volumeId string, waitSeconds int) error {
	err := c.ServerDeleteVolume(id, volumeId)
	if err != nil {
		return err
	}
	startTime := time.Now()
	for {
		detached := true
		attachedVolumes, err := c.ListServerVolumes(id)
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

func (c NovaV2) CopyFlavor(id string, newName string, newId string,
	newVcpus int, newRam int, newDisk int, newSwap int,
	newEphemeral int, newRxtxFactor float32, setProperties map[string]string,
	unsetProperties []string,
) (*nova.Flavor, error) {
	console.Info("Show flavor")
	flavor, err := c.GetFlavor(id)
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
	extraSpecs, err := c.GetFlavorExtraSpecs(id)
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
	newFlavor, err := c.CreateFlavor(*flavor)
	if err != nil {
		return nil, fmt.Errorf("create flavor failed, %v", err)
	}
	if len(extraSpecs) != 0 {
		console.Info("Set new flavor extra specs")
		_, err = c.SetFlavorExtraSpecs(newFlavor.Id, extraSpecs)
		if err != nil {
			return nil, fmt.Errorf("set flavor extra specs failed, %v", err)
		}
		newFlavor.ExtraSpecs = extraSpecs
	}
	return newFlavor, nil
}
