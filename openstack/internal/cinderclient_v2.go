package internal

import (
	"fmt"
	"net/url"
	"time"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/utility"
	"github.com/samber/lo"
)

type CinderV2 struct{ *ServiceClient }

// volume api

func (c CinderV2) ListVolume(query url.Values, details ...bool) ([]cinder.Volume, error) {
	url := URL_VOLUMES
	if lo.FirstOrEmpty(details) {
		url = URL_VOLUMES_DETAIL
	}
	return QueryResource[cinder.Volume](c.ServiceClient, url.F(), query, "volumes")
}
func (c CinderV2) ListByName(name string, details ...bool) ([]cinder.Volume, error) {
	return c.ListVolume(utility.UrlValues(map[string]string{"name": name}), details...)
}
func (c CinderV2) GetVolume(id string) (*cinder.Volume, error) {
	return GetResource[cinder.Volume](c.ServiceClient, URL_VOLUME.F(id), "volume")
}

func (c CinderV2) FindVolume(idOrName string, allTenats ...bool) (*cinder.Volume, error) {
	return QueryByIdOrName(idOrName,
		c.GetVolume,
		func(query url.Values) ([]cinder.Volume, error) {
			if lo.CoalesceOrEmpty(allTenats...) {
				query.Set("all_tenants", "1")
			}
			return c.ListVolume(query, true)
		})
}

func (c CinderV2) CreateVolume(options map[string]any) (*cinder.Volume, error) {
	result := struct {
		Volume cinder.Volume `json:"volume"`
	}{Volume: cinder.Volume{}}

	_, err := c.R().SetBody(map[string]any{"volume": options}).
		SetResult(&result).Post(URL_VOLUMES.F())
	return &result.Volume, err
}

func (c CinderV2) DeleteVolume(id string, force bool, cascade bool) error {
	query := url.Values{}
	if force {
		query.Add("force", "false")
	}
	if cascade {
		query.Add("cascade", "true")
	}
	return DeleteResource(c.ServiceClient, URL_VOLUME.F(id))
}
func (c CinderV2) doVolumeAction(id string, data any) error {
	_, err := c.R().SetBody(data).Post(URL_VOLUME_ACTION.F(id))
	return err
}

func (c CinderV2) ExtendVolume(id string, size int) error {
	return c.doVolumeAction(id, ReqBody{
		"os-extend": {"new_size": size},
	})
}
func (c CinderV2) RetypeVolume(id string, newType string, migrationPolicy string) error {
	return c.doVolumeAction(id, ReqBody{
		"os-retype": {
			"new_type":         newType,
			"migration_policy": migrationPolicy,
		},
	})
}

func (c CinderV2) RevertVolume(id string, snapshotId string) error {
	return c.doVolumeAction(id, ReqBody{
		"revert": {"snapshot_id": snapshotId}})
}

// volume type api

func (c CinderV2) ListType(query url.Values) ([]cinder.VolumeType, error) {
	return QueryResource[cinder.VolumeType](c.ServiceClient, URL_VOLUME_TYPES.F(), query, "volume_types")
}
func (c CinderV2) GetType(id string) (*cinder.VolumeType, error) {
	return GetResource[cinder.VolumeType](c.ServiceClient, URL_VOLUME_TYPE.F(id), "volume_type")
}

func (c CinderV2) GetDefaultType() (*cinder.VolumeType, error) {
	return GetResource[cinder.VolumeType](c.ServiceClient, URL_VOLUME_TYPE_DEFAULT.F(), "volume_type")
}
func (c CinderV2) FindType(idOrName string) (*cinder.VolumeType, error) {
	return QueryByIdOrName(idOrName, c.GetType, c.ListType)
}
func (c CinderV2) CreateType(params map[string]any) (*cinder.VolumeType, error) {
	result := struct {
		VolumeType cinder.VolumeType `json:"volume_type"`
	}{}
	_, err := c.R().SetBody(ReqBody{"volume_type": params}).SetResult(&result).
		Post(URL_VOLUME_TYPES.F())
	if err != nil {
		return nil, err
	}
	return &result.VolumeType, err
}
func (c CinderV2) DeleteType(id string) error {
	return DeleteResource(c.ServiceClient, URL_VOLUME_TYPE.F(id))
}

// volume service api

func (c CinderV2) ListService(query url.Values) ([]cinder.Service, error) {
	return QueryResource[cinder.Service](c.ServiceClient, URL_VOLUME_SERVICES.F(), query, "services")
}

// snapshot api

func (c CinderV2) ListSnapshot(query url.Values, details ...bool) ([]cinder.Snapshot, error) {
	url := URL_SNAPSHOTS
	if lo.FirstOrEmpty(details) {
		url = URL_SNAPSHOTS
	}
	return QueryResource[cinder.Snapshot](c.ServiceClient, url.F(), query, "snapshots")
}
func (c CinderV2) GetSnapshot(id string) (*cinder.Snapshot, error) {
	return GetResource[cinder.Snapshot](c.ServiceClient, URL_SNAPSHOT.F(id), "snapshot")
}
func (c CinderV2) DeleteSnapshot(id string) error {
	return DeleteResource(c.ServiceClient, URL_SNAPSHOT.F(id))
}
func (c CinderV2) FindSnapshot(idOrName string, allTenats ...bool) (*cinder.Snapshot, error) {
	return QueryByIdOrName(idOrName, c.GetSnapshot,
		func(query url.Values) ([]cinder.Snapshot, error) {
			if lo.CoalesceOrEmpty(allTenats...) {
				query.Set("all_tenants", "1")
			}
			return c.ListSnapshot(query, true)
		})
}
func (c CinderV2) CreateSnapshot(volumeId string, name string, force bool) (*cinder.Snapshot, error) {
	result := struct {
		Snapshot cinder.Snapshot
	}{}
	params := map[string]any{
		"volume_id": volumeId,
		"name":      name,
	}
	if force {
		params["force"] = force
	}
	_, err := c.R().SetBody(ReqBody{"snapshot": params}).SetResult(&result).
		Post(URL_SNAPSHOTS.F())
	if err != nil {
		return nil, err
	}
	return &result.Snapshot, err
}

// backup api

func (c CinderV2) ListBackup(query url.Values, details ...bool) ([]cinder.Backup, error) {
	url := URL_BACKUPS
	if lo.FirstOrEmpty(details) {
		url = URL_BACKUPS_DETAIL
	}
	return QueryResource[cinder.Backup](c.ServiceClient, url.F(), query, "backups")

}
func (c CinderV2) GetBackup(id string) (*cinder.Backup, error) {
	return GetResource[cinder.Backup](c.ServiceClient, URL_VOLUME.F(id), "backup")
}
func (c CinderV2) DeleteBackup(id string) error {
	return DeleteResource(c.ServiceClient, URL_BACKUP.F(id))
}
func (c CinderV2) FindBackup(idOrName string, allTenats ...bool) (*cinder.Backup, error) {
	return QueryByIdOrName(idOrName, c.GetBackup, func(query url.Values) ([]cinder.Backup, error) {
		if lo.CoalesceOrEmpty(allTenats...) {
			query.Set("all_tenants", "1")
		}
		return c.ListBackup(query, true)
	})
}
func (c CinderV2) CreateBackup(volumeId string, name string, force bool) (*cinder.Backup, error) {
	result := struct{ Backup cinder.Backup }{}
	params := map[string]any{
		"volume_id": volumeId,
		"name":      name,
	}
	if force {
		params["force"] = force
	}
	_, err := c.R().SetBody(ReqBody{"backup": params}).SetResult(&result).
		Post(URL_BACKUPS.F())
	if err != nil {
		return nil, err
	}
	return &result.Backup, err
}

func (c CinderV2) GetCurrentVersion() (*model.ApiVersion, error) {
	result := struct{ Versions model.ApiVersions }{}
	if _, err := c.Index(&result); err != nil {
		return nil, err
	}
	version := result.Versions.Current()
	if version != nil {
		return version, nil
	}
	return nil, fmt.Errorf("current version not found")
}

// 扩展的方法

func (c CinderV2) CreateVolumeAndWait(options map[string]any, timeoutSeconds int) (*cinder.Volume, error) {
	volume, err := c.CreateVolume(options)
	if err != nil {
		return nil, err
	}
	startTime := time.Now()
	for {
		vol, err := c.GetVolume(volume.Id)
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
