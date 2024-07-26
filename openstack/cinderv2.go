package openstack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/utility"
)

const V2 = "v2"

type CinderV2 struct {
	RestClient2
}

const (
	POLICY_NEVER     = "never"
	POLICY_ON_DEMAND = "on-demand"
)

var MIGRATION_POLICYS = []string{POLICY_NEVER, POLICY_ON_DEMAND}

func InvalidMIgrationPoicy(policy string) error {
	if !stringutils.ContainsString(MIGRATION_POLICYS, policy) {
		return fmt.Errorf("invalid migration policy: %s, supported: %s", policy, MIGRATION_POLICYS)
	}
	return nil
}

func (o *Openstack) CinderV2() *CinderV2 {
	if o.cinderClient == nil {
		var (
			endpoint string
			err      error
		)
		endpoint, err = o.AuthPlugin.GetServiceEndpoint("volumev2", "cinderv2", "public")
		if err != nil {
			logging.Warning("get endpoint falied: %v", err)
		}
		o.cinderClient = &CinderV2{
			NewRestClient2(utility.VersionUrl(endpoint, V2), o.AuthPlugin),
		}
	}
	return o.cinderClient
}
func (c CinderV2) GetCurrentVersion() (*model.ApiVersion, error) {
	resp, err := c.Index()
	if err != nil {
		return nil, err
	}
	body := struct{ Versions []model.ApiVersion }{}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	for _, version := range body.Versions {
		if version.Status == "CURRENT" {
			return &version, nil
		}
	}
	return nil, fmt.Errorf("current version not found")
}

func (c CinderV2) Volume() volumeApi {
	return volumeApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "volumes",
			Client:      c.session,
			SingularKey: "volume",
			PluralKey:   "volumes",
		},
	}
}

// volume api
type volumeApi struct {
	ResourceApi
}

func (c volumeApi) Detail(query url.Values) ([]cinder.Volume, error) {
	body := struct{ Volumes []cinder.Volume }{}
	if _, err := c.AppendUrl("detail").SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Volumes, nil
}
func (c volumeApi) List(query url.Values) ([]cinder.Volume, error) {
	body := struct{ Volumes []cinder.Volume }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Volumes, nil
}
func (c volumeApi) ListByName(name string) ([]cinder.Volume, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}
func (c volumeApi) DetailByName(name string) ([]cinder.Volume, error) {
	return c.Detail(utility.UrlValues(map[string]string{"name": name}))
}
func (c volumeApi) Show(id string) (*cinder.Volume, error) {
	body := struct{ Volume cinder.Volume }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Volume, nil
}

func (c volumeApi) Found(idOrName string) (*cinder.Volume, error) {
	volume, err := FoundResource[cinder.Volume](c.ResourceApi, idOrName)
	if err != nil {
		return nil, err
	}
	if volume.Status == "" {
		return c.Show(volume.Id)
	} else {
		return volume, err
	}
}
func (c volumeApi) Create(options map[string]interface{}) (*cinder.Volume, error) {
	body := struct {
		Volume cinder.Volume `json:"volume"`
	}{Volume: cinder.Volume{}}
	_, err := c.SetBody(map[string]interface{}{"volume": options}).Post(&body)
	if err != nil {
		return nil, err
	}
	return &body.Volume, nil

}
func (c volumeApi) CreateAndWait(options map[string]interface{}, timeoutSeconds int) (*cinder.Volume, error) {
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
func (c volumeApi) Delete(id string, force bool, cascade bool) error {
	query := url.Values{}
	if force {
		query.Add("force", "false")
		query.Add("cascade", "true")
	}
	if cascade {
		query.Add("cascade", "true")
	}
	_, err := c.AppendUrl(id).SetQuery(query).Delete(nil)
	return err
}

func (c volumeApi) Extend(id string, size int) error {
	repData := map[string]map[string]interface{}{
		"os-extend": {
			"new_size": size,
		},
	}
	return c.doAction(id, repData)
}
func (c volumeApi) Retype(id string, newType string, migrationPolicy string) error {
	repData := map[string]map[string]interface{}{
		"os-retype": {
			"new_type":         newType,
			"migration_policy": migrationPolicy,
		},
	}
	return c.doAction(id, repData)
}

func (c volumeApi) doAction(id string, data interface{}) error {
	reqBody, _ := json.Marshal(data)
	_, err := c.AppendUrl(id).AppendUrl("action").
		SetBody(reqBody).Post(nil)
	return err
}

// volume type api
type volumeTypeApi struct {
	ResourceApi
}

func (c CinderV2) VolumeType() volumeTypeApi {
	return volumeTypeApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "types",
			Client:      c.session,
			SingularKey: "volume_type",
			PluralKey:   "volume_types",
		},
	}
}
func (c volumeTypeApi) List(query url.Values) ([]cinder.VolumeType, error) {
	body := struct {
		VolumeTypes []cinder.VolumeType `json:"volume_types"`
	}{}

	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.VolumeTypes, nil
}
func (c volumeTypeApi) Show(id string) (*cinder.VolumeType, error) {
	body := struct {
		VolumeType cinder.VolumeType `json:"volume_type"`
	}{}

	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.VolumeType, nil
}

func (c volumeTypeApi) Default() (*cinder.VolumeType, error) {
	return c.Show("default")
}
func (c volumeTypeApi) Found(idOrName string) (*cinder.VolumeType, error) {
	return FoundResource[cinder.VolumeType](c.ResourceApi, idOrName)
}
func (c volumeTypeApi) Create(params map[string]interface{}) (*cinder.VolumeType, error) {
	body := struct {
		VolumeType cinder.VolumeType `json:"volume_type"`
	}{}
	_, err := c.SetBody(map[string]interface{}{"volume_type": params}).Post(&body)
	if err != nil {
		return nil, err
	}
	return &body.VolumeType, err
}
func (c volumeTypeApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}

// volume service api
type volumeServiceApi struct {
	ResourceApi
}

func (c CinderV2) Service() volumeServiceApi {
	return volumeServiceApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "os-services",
			Client:      c.session,
			SingularKey: "service",
			PluralKey:   "services",
		},
	}
}

func (c volumeServiceApi) List(query url.Values) ([]cinder.Service, error) {
	body := struct{ Services []cinder.Service }{}
	_, err := c.SetQuery(query).Get(&body)
	if err != nil {
		return nil, err
	}
	return body.Services, nil
}

// snapshot api
type snapshotApi struct {
	ResourceApi
}

func (c CinderV2) Snapshot() snapshotApi {
	return snapshotApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "snapshots",
			Client:      c.session,
			SingularKey: "snapshot",
			PluralKey:   "snapshots",
		},
	}
}
func (c snapshotApi) List(query url.Values) ([]cinder.Snapshot, error) {
	body := struct{ Snapshots []cinder.Snapshot }{}
	_, err := c.SetQuery(query).Get(&body)
	if err != nil {
		return nil, err
	}
	return body.Snapshots, nil
}
func (c snapshotApi) Show(id string) (*cinder.Snapshot, error) {
	body := struct{ Snapshot cinder.Snapshot }{}
	_, err := c.AppendUrl(id).Get(&body)
	if err != nil {
		return nil, err
	}
	return &body.Snapshot, nil
}
func (c snapshotApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}
func (c snapshotApi) Found(idOrName string) (*cinder.Snapshot, error) {
	return FoundResource[cinder.Snapshot](c.ResourceApi, idOrName)
}
