package internal

import (
	"fmt"
	"net/url"
	"time"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/utility"
)

const (
	VOLUMES   = "volumes"
	SNAPSHOTS = "snapshots"
	BACKUPS   = "backups"
)

type VolumeApi struct{ ResourceApi }
type VolumeTypeApi struct{ ResourceApi }
type VolumeServiceApi struct{ ResourceApi }
type SnapshotApi struct{ ResourceApi }
type BackupApi struct{ ResourceApi }

func (c CinderV2) Volume() VolumeApi {
	return VolumeApi{
		ResourceApi: ResourceApi{
			Client: c.Client, BaseUrl: c.BaserUrl(),
			ResourceUrl: "volumes",
			SingularKey: "volume",
			PluralKey:   VOLUMES,
		},
	}
}
func (c CinderV2) Service() VolumeServiceApi {
	return VolumeServiceApi{
		ResourceApi: ResourceApi{
			Client: c.Client, BaseUrl: c.BaserUrl(),
			ResourceUrl: "os-services",
			SingularKey: "service",
			PluralKey:   "services",
		},
	}
}

func (c CinderV2) Snapshot() SnapshotApi {
	return SnapshotApi{
		ResourceApi: ResourceApi{
			Client: c.Client, BaseUrl: c.BaserUrl(),
			ResourceUrl: "snapshots",
			SingularKey: "snapshot",
			PluralKey:   SNAPSHOTS,
		},
	}
}
func (c CinderV2) Backup() BackupApi {
	return BackupApi{
		ResourceApi: ResourceApi{Client: c.Client, BaseUrl: c.BaserUrl(),
			ResourceUrl: "backups",
			SingularKey: "backup",
			PluralKey:   BACKUPS,
		},
	}
}

func (c CinderV2) VolumeType() VolumeTypeApi {
	return VolumeTypeApi{
		ResourceApi{
			Client:      c.Client,
			BaseUrl:     c.BaserUrl(),
			ResourceUrl: "types",
			SingularKey: "volume_type",
			PluralKey:   "volume_types",
		},
	}
}

type ReqBody map[string]map[string]interface{}

// volume api

func (c VolumeApi) Detail(query url.Values) ([]cinder.Volume, error) {
	return ListResource[cinder.Volume](c.ResourceApi, query, true)
}
func (c VolumeApi) List(query url.Values) ([]cinder.Volume, error) {
	return ListResource[cinder.Volume](c.ResourceApi, query)
}
func (c VolumeApi) ListByName(name string) ([]cinder.Volume, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}
func (c VolumeApi) DetailByName(name string) ([]cinder.Volume, error) {
	return c.Detail(utility.UrlValues(map[string]string{"name": name}))
}
func (c VolumeApi) Show(id string) (*cinder.Volume, error) {
	return ShowResource[cinder.Volume](c.ResourceApi, id)
}

func (c VolumeApi) Find(idOrName string) (*cinder.Volume, error) {
	return FindIdOrName(c, idOrName)
}
func (c VolumeApi) Create(options map[string]interface{}) (*cinder.Volume, error) {
	result := struct {
		Volume cinder.Volume `json:"volume"`
	}{Volume: cinder.Volume{}}

	_, err := c.R().SetBody(map[string]interface{}{"volume": options}).SetResult(&result).Post()
	if err != nil {
		return nil, err
	}
	return &result.Volume, nil
}
func (c VolumeApi) CreateAndWait(options map[string]interface{}, timeoutSeconds int) (*cinder.Volume, error) {
	volume, err := c.Create(options)
	if err != nil {
		return nil, err
	}
	startTime := time.Now()
	for {
		vol, err := c.Show(volume.Id)
		if err != nil {
			return volume, fmt.Errorf("show volume failed: %s", err)
		}
		if vol.IsError() {
			return volume, fmt.Errorf("status is error")
		}
		if vol.IsAvailable() {
			return volume, nil
		}
		if time.Since(startTime) >= time.Second*time.Duration(timeoutSeconds) {
			return volume, fmt.Errorf("create timeout")
		}
		time.Sleep(time.Second * 2)
	}
}
func (c VolumeApi) Delete(id string, force bool, cascade bool) error {
	query := url.Values{}
	if force {
		query.Add("force", "false")
	}
	if cascade {
		query.Add("cascade", "true")
	}
	_, err := DeleteResource(c.ResourceApi, id, query)
	return err
}
func (c VolumeApi) doAction(id string, data interface{}) error {
	_, err := c.R().SetBody(data).Post(id, "action")
	return err
}

func (c VolumeApi) Extend(id string, size int) error {
	return c.doAction(id, ReqBody{
		"os-extend": {
			"new_size": size,
		},
	})
}
func (c VolumeApi) Retype(id string, newType string, migrationPolicy string) error {
	return c.doAction(id, ReqBody{
		"os-retype": {
			"new_type":         newType,
			"migration_policy": migrationPolicy,
		},
	})
}

func (c VolumeApi) Revert(id string, snapshotId string) error {
	return c.doAction(id, ReqBody{
		"revert": {
			"snapshot_id": snapshotId,
		}})
}

// volume type api

func (c VolumeTypeApi) List(query url.Values) ([]cinder.VolumeType, error) {
	return ListResource[cinder.VolumeType](c.ResourceApi, query)
}
func (c VolumeTypeApi) Show(id string) (*cinder.VolumeType, error) {
	return ShowResource[cinder.VolumeType](c.ResourceApi, id)
}

func (c VolumeTypeApi) Default() (*cinder.VolumeType, error) {
	return c.Show("default")
}
func (c VolumeTypeApi) Find(idOrName string) (*cinder.VolumeType, error) {
	return FindIdOrName(c, idOrName)
}
func (c VolumeTypeApi) Create(params map[string]interface{}) (*cinder.VolumeType, error) {
	result := struct {
		VolumeType cinder.VolumeType `json:"volume_type"`
	}{}
	_, err := c.R().SetBody(ReqBody{"volume_type": params}).SetResult(&result).Post()
	if err != nil {
		return nil, err
	}
	return &result.VolumeType, err
}
func (c VolumeTypeApi) Delete(id string) error {
	_, err := c.ResourceDelete(id)
	return err
}

// volume service api

func (c VolumeServiceApi) List(query url.Values) ([]cinder.Service, error) {
	return ListResource[cinder.Service](c.ResourceApi, query)
}

// snapshot api

func (c SnapshotApi) List(query url.Values) ([]cinder.Snapshot, error) {
	return ListResource[cinder.Snapshot](c.ResourceApi, query)
}
func (c SnapshotApi) Detail(query url.Values) ([]cinder.Snapshot, error) {
	return ListResource[cinder.Snapshot](c.ResourceApi, query, true)
}
func (c SnapshotApi) Show(id string) (*cinder.Snapshot, error) {
	return ShowResource[cinder.Snapshot](c.ResourceApi, id)
}
func (c SnapshotApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id)
	return err
}
func (c SnapshotApi) Find(idOrName string) (*cinder.Snapshot, error) {
	return FindIdOrName(c, idOrName)
}
func (c SnapshotApi) Create(volumeId string, name string, force bool) (*cinder.Snapshot, error) {
	result := struct {
		Snapshot cinder.Snapshot
	}{}
	params := map[string]interface{}{
		"volume_id": volumeId,
		"name":      name,
	}
	if force {
		params["force"] = force
	}
	_, err := c.R().SetBody(ReqBody{"snapshot": params}).SetResult(&result).Post()
	if err != nil {
		return nil, err
	}
	return &result.Snapshot, err
}

// backup api

func (c BackupApi) List(query url.Values) ([]cinder.Backup, error) {
	return ListResource[cinder.Backup](c.ResourceApi, query)
}
func (c BackupApi) Detail(query url.Values) ([]cinder.Backup, error) {
	return ListResource[cinder.Backup](c.ResourceApi, query, true)
}
func (c BackupApi) Show(id string) (*cinder.Backup, error) {
	return ShowResource[cinder.Backup](c.ResourceApi, id)
}
func (c BackupApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id)
	return err
}
func (c BackupApi) Find(idOrName string) (*cinder.Backup, error) {
	return FindIdOrName(c, idOrName)
}
func (c BackupApi) Create(volumeId string, name string, force bool) (*cinder.Backup, error) {
	result := struct{ Backup cinder.Backup }{}
	params := map[string]interface{}{
		"volume_id": volumeId,
		"name":      name,
	}
	if force {
		params["force"] = force
	}
	_, err := c.R().SetBody(ReqBody{"backup": params}).SetResult(&result).Post()
	if err != nil {
		return nil, err
	}
	return &result.Backup, err
}

type CinderV2 struct{ *ServiceClient }

func (c CinderV2) GetCurrentVersion() (*model.ApiVersion, error) {
	result := struct{ Versions model.ApiVersions }{}

	if resp, err := c.Index(nil); err != nil {
		return nil, err
	} else if err := resp.UnmarshalBody(&result); err != nil {
		return nil, err
	}
	version := result.Versions.Current()
	if version != nil {
		return version, nil
	}
	return nil, fmt.Errorf("current version not found")
}
