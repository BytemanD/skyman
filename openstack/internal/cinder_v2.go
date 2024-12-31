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

// volume api
type VolumeApi struct{ ResourceApi }

func (c VolumeApi) Detail(query url.Values) ([]cinder.Volume, error) {
	result := struct{ Volumes []cinder.Volume }{}
	if _, err := c.Get("volumes/detail", query, &result); err != nil {
		return nil, err
	}
	return result.Volumes, nil
}
func (c VolumeApi) List(query url.Values) ([]cinder.Volume, error) {
	result := struct{ Volumes []cinder.Volume }{}
	if _, err := c.Get("volumes", query, &result); err != nil {
		return nil, err
	}
	return result.Volumes, nil
}
func (c VolumeApi) ListByName(name string) ([]cinder.Volume, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}
func (c VolumeApi) DetailByName(name string) ([]cinder.Volume, error) {
	return c.Detail(utility.UrlValues(map[string]string{"name": name}))
}
func (c VolumeApi) Show(id string) (*cinder.Volume, error) {
	result := struct{ Volume cinder.Volume }{}
	if _, err := c.Get("volumes/"+id, nil, &result); err != nil {
		return nil, err
	}
	return &result.Volume, nil
}

func (c VolumeApi) Find(idOrName string) (*cinder.Volume, error) {
	return FindResource(idOrName, c.Show, c.Detail)
}
func (c VolumeApi) Create(options map[string]interface{}) (*cinder.Volume, error) {
	result := struct {
		Volume cinder.Volume `json:"volume"`
	}{Volume: cinder.Volume{}}

	_, err := c.Post("volumes", map[string]interface{}{"volume": options}, &result)
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
	_, err := c.ResourceDelete(id, query)
	return err
}

func (c VolumeApi) Extend(id string, size int) error {
	repData := map[string]map[string]interface{}{
		"os-extend": {
			"new_size": size,
		},
	}
	return c.doAction(id, repData)
}
func (c VolumeApi) Retype(id string, newType string, migrationPolicy string) error {
	repData := map[string]map[string]interface{}{
		"os-retype": {
			"new_type":         newType,
			"migration_policy": migrationPolicy,
		},
	}
	return c.doAction(id, repData)
}

func (c VolumeApi) doAction(id string, data interface{}) error {
	_, err := c.Post(fmt.Sprintf("volumes/%s/action", id), data, nil)
	return err
}
func (c VolumeApi) Revert(id string, snapshotId string) error {
	data := struct {
		Revert map[string]string `json:"revert"`
	}{
		Revert: map[string]string{"snapshot_id": snapshotId},
	}
	return c.doAction(id, data)
}

// volume type api
type VolumeTypeApi struct{ ResourceApi }

func (c CinderV2) VolumeType() VolumeTypeApi {
	return VolumeTypeApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Url},
	}
}
func (c VolumeTypeApi) List(query url.Values) ([]cinder.VolumeType, error) {
	result := struct {
		VolumeTypes []cinder.VolumeType `json:"volume_types"`
	}{}

	if _, err := c.Get("types", query, &result); err != nil {
		return nil, err
	}
	return result.VolumeTypes, nil
}
func (c VolumeTypeApi) Show(id string) (*cinder.VolumeType, error) {
	result := struct {
		VolumeType cinder.VolumeType `json:"volume_type"`
	}{}

	if _, err := c.Get("types/"+id, nil, &result); err != nil {
		return nil, err
	}
	return &result.VolumeType, nil
}

func (c VolumeTypeApi) Default() (*cinder.VolumeType, error) {
	return c.Show("default")
}
func (c VolumeTypeApi) Find(idOrName string) (*cinder.VolumeType, error) {
	return FindResource(idOrName, c.Show, c.List)
}
func (c VolumeTypeApi) Create(params map[string]interface{}) (*cinder.VolumeType, error) {
	result := struct {
		VolumeType cinder.VolumeType `json:"volume_type"`
	}{}
	_, err := c.Post("types", map[string]interface{}{"volume_type": params}, &result)
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
type VolumeServiceApi struct{ ResourceApi }

func (c VolumeServiceApi) List(query url.Values) ([]cinder.Service, error) {
	result := struct{ Services []cinder.Service }{}
	_, err := c.Get("os-services", query, &result)
	if err != nil {
		return nil, err
	}
	return result.Services, nil
}

// snapshot api
type SnapshotApi struct{ ResourceApi }

func (c SnapshotApi) List(query url.Values) ([]cinder.Snapshot, error) {
	result := struct{ Snapshots []cinder.Snapshot }{}
	_, err := c.Get(SNAPSHOTS, query, &result)
	if err != nil {
		return nil, err
	}
	return result.Snapshots, nil
}
func (c SnapshotApi) Detail(query url.Values) ([]cinder.Snapshot, error) {
	result := struct{ Snapshots []cinder.Snapshot }{}
	_, err := c.Get(SNAPSHOTS+"/detail", query, &result)
	if err != nil {
		return nil, err
	}
	return result.Snapshots, nil
}
func (c SnapshotApi) Show(id string) (*cinder.Snapshot, error) {
	result := struct{ Snapshot cinder.Snapshot }{}
	_, err := c.Get("snapshots/"+id, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result.Snapshot, nil
}
func (c SnapshotApi) Delete(id string) error {
	_, err := c.ResourceApi.Delete("snapshots/" + id)
	return err
}
func (c SnapshotApi) Find(idOrName string) (*cinder.Snapshot, error) {
	return FindResource(idOrName, c.Show, c.List)
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
	_, err := c.Post(SNAPSHOTS, map[string]interface{}{"snapshot": params}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Snapshot, err
}

// snapshot api
type BackupApi struct{ ResourceApi }

func (c BackupApi) List(query url.Values) ([]cinder.Backup, error) {
	result := struct{ Backups []cinder.Backup }{}
	_, err := c.Get(BACKUPS, query, &result)
	if err != nil {
		return nil, err
	}
	return result.Backups, nil
}
func (c BackupApi) Detail(query url.Values) ([]cinder.Backup, error) {
	result := struct{ Backups []cinder.Backup }{}
	_, err := c.Get(BACKUPS+"/detail", query, &result)
	if err != nil {
		return nil, err
	}
	return result.Backups, nil
}
func (c BackupApi) Show(id string) (*cinder.Backup, error) {
	result := struct{ Backup cinder.Backup }{}
	_, err := c.Get("backups/"+id, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result.Backup, nil
}
func (c BackupApi) Delete(id string) error {
	_, err := c.ResourceDelete(id)
	return err
}
func (c BackupApi) Find(idOrName string) (*cinder.Backup, error) {
	return FindResource(idOrName, c.Show, c.List)
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
	_, err := c.Post(BACKUPS, map[string]interface{}{"backup": params}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Backup, err
}

type CinderV2 struct{ *ServiceClient }

func (c CinderV2) GetCurrentVersion() (*model.ApiVersion, error) {
	result := struct{ Versions []model.ApiVersion }{}
	_, err := c.Index(&result)
	if err != nil {
		return nil, err
	}
	for _, version := range result.Versions {
		if version.Status == "CURRENT" {
			return &version, nil
		}
	}
	return nil, fmt.Errorf("current version not found")
}

func (c CinderV2) Volume() VolumeApi {
	return VolumeApi{
		ResourceApi: ResourceApi{Client: c.rawClient, BaseUrl: c.Url,
			ResourceUrl: "volumes"},
	}
}
func (c CinderV2) Service() VolumeServiceApi {
	return VolumeServiceApi{
		ResourceApi: ResourceApi{Client: c.rawClient, BaseUrl: c.Url},
	}
}

func (c CinderV2) Snapshot() SnapshotApi {
	return SnapshotApi{
		ResourceApi: ResourceApi{Client: c.rawClient, BaseUrl: c.Url},
	}
}
func (c CinderV2) Backup() BackupApi {
	return BackupApi{
		ResourceApi: ResourceApi{Client: c.rawClient, BaseUrl: c.Url,
			ResourceUrl: "backups"},
	}
}
