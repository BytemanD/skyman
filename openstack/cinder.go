package openstack

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/utility"
)

const V2 = "v2"

type CinderV2 struct {
	RestClient
}
type VolumeApi struct {
	CinderV2
}

type VolumeTypeApi struct {
	CinderV2
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
			NewRestClient(utility.VersionUrl(endpoint, V2), o.AuthPlugin),
		}
	}
	return o.cinderClient
}
func (c CinderV2) GetCurrentVersion() (*model.ApiVersion, error) {
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

func (c CinderV2) Volumes() VolumeApi {
	return VolumeApi{c}
}
func (c CinderV2) VolumeTypes() VolumeTypeApi {
	return VolumeTypeApi{c}
}
func (c VolumeApi) Detail(query url.Values) ([]cinder.Volume, error) {
	resp, err := c.CinderV2.Get("volumes/detail", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]cinder.Volume{"volumes": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["volumes"], nil
}
func (c VolumeApi) List(query url.Values) ([]cinder.Volume, error) {
	resp, err := c.CinderV2.Get("volumes", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]cinder.Volume{"volumes": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["volumes"], nil
}
func (c VolumeApi) ListByName(name string) ([]cinder.Volume, error) {
	return c.List(utility.UrlValues(map[string]string{
		"name": name,
	}))
}
func (c VolumeApi) DetailByName(name string) ([]cinder.Volume, error) {
	return c.Detail(utility.UrlValues(map[string]string{
		"name": name,
	}))
}
func (c VolumeApi) Show(id string) (*cinder.Volume, error) {
	resp, err := c.CinderV2.Get(utility.UrlJoin("volumes", id), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*cinder.Volume{"volumes": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["volumes"], nil
}

func (c VolumeApi) Found(idOrName string) (*cinder.Volume, error) {
	volume, err := c.Show(idOrName)
	if err == nil {
		return volume, nil
	}
	if httpError, ok := err.(*utility.HttpError); ok {
		if !httpError.IsNotFound() {
			return nil, err
		}
	}
	volumes, err := c.DetailByName(idOrName)
	if err == nil {
		if len(volumes) == 0 {
			return nil, fmt.Errorf("volume %s not found", idOrName)
		} else if len(volumes) == 1 {
			return &(volumes[0]), nil
		} else {
			return nil, fmt.Errorf("Found multi volumes named %s", idOrName)
		}
	}
	return nil, err
}
func (c VolumeApi) Create(options map[string]interface{}) (*cinder.Volume, error) {
	reqBody, err := json.Marshal(&options)
	if err != nil {
		return nil, err
	}
	resp, err := c.CinderV2.Post("volumes", reqBody, nil)
	if err != nil {
		return nil, err
	}
	image := cinder.Volume{}
	resp.BodyUnmarshal(&image)
	return &image, nil
}
func (c VolumeApi) Delete(id string) (err error) {
	_, err = c.CinderV2.Delete(utility.UrlJoin("volumes", id), nil)
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
	reqBody, _ := json.Marshal(data)
	_, err := c.CinderV2.Post(
		utility.UrlJoin("volumes", id), reqBody, nil,
	)
	if err != nil {
		return err
	}
	return nil
}
func (c VolumeTypeApi) List(query url.Values) ([]cinder.VolumeType, error) {
	resp, err := c.CinderV2.Get("types", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]cinder.VolumeType{"volume_types": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["volume_types"], nil
}
func (c VolumeTypeApi) Show(id string) (*cinder.VolumeType, error) {
	resp, err := c.CinderV2.Get(utility.UrlJoin("types", id), nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*cinder.VolumeType{"volume_type": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["volume_type"], nil
}
func (c VolumeTypeApi) ListByName(name string) ([]cinder.VolumeType, error) {
	volumeTypes, err := c.List(nil)
	if err != nil {
		return nil, err
	}
	foundVolumeTypes := []cinder.VolumeType{}
	for _, volumeType := range volumeTypes {
		if volumeType.Name == name {
			foundVolumeTypes = append(foundVolumeTypes, volumeType)
		}
	}
	return foundVolumeTypes, nil
}
func (c VolumeTypeApi) Default() (*cinder.VolumeType, error) {
	return c.Show("default")
}
func (c VolumeTypeApi) Found(idOrName string) (*cinder.VolumeType, error) {
	volumeType, err := c.Show(idOrName)
	if err == nil {
		return volumeType, nil
	}
	if httpError, ok := err.(*utility.HttpError); ok {
		if !httpError.IsNotFound() {
			return nil, err
		}
	}
	volumeTypes, err := c.ListByName(idOrName)
	if err == nil {
		if len(volumeTypes) == 0 {
			return nil, fmt.Errorf("volume type %s not found", idOrName)
		} else if len(volumeTypes) == 1 {
			return &(volumeTypes[0]), nil
		} else {
			return nil, fmt.Errorf("Found multi volume types named %s", idOrName)
		}
	}
	return nil, err
}
func (c VolumeTypeApi) Create(params map[string]interface{}) (*cinder.VolumeType, error) {
	data := map[string]interface{}{"volume_type": params}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	resp, err := c.CinderV2.Post("volumes", body, nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*cinder.VolumeType{"volume_type": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["volume_type"], nil
}
func (c VolumeTypeApi) Delete(id string) (err error) {
	_, err = c.CinderV2.Delete(utility.UrlJoin("types", id), nil)
	return err
}
func (c VolumeApi) Prune(query url.Values, yes bool) {
	if query == nil {
		query = url.Values{}
	}
	if len(query) == 0 {
		query.Add("status", "error")
	}
	logging.Info("查询卷: %v", query)
	volumes, err := c.List(query)
	if err != nil {
		logging.Error("get volumes failed, %s", err)
		return
	}
	logging.Info("需要清理的卷数量: %d\n", len(volumes))
	if len(volumes) == 0 {
		return
	}
	if !yes {
		fmt.Printf("即将清理 %d 个卷:\n", len(volumes))
		for _, server := range volumes {
			fmt.Printf("%s (%s)\n", server.Id, server.Name)
		}
		yes = stringutils.ScanfComfirm("是否删除?", []string{"yes", "y"}, []string{"no", "n"})
		if !yes {
			return
		}
	}
	logging.Info("开始清理")
	tg := syncutils.TaskGroup{
		Func: func(i interface{}) error {
			volume := i.(cinder.Volume)
			logging.Debug("delete volume %s(%s)", volume.Id, volume.Name)
			err := c.Delete(volume.Id)
			if err != nil {
				return fmt.Errorf("delete volume %s failed: %v", volume.Id, err)
			}
			return nil
		},
		Items:        volumes,
		ShowProgress: true,
	}
	err = tg.Start()
	if err != nil {
		logging.Error("清理失败: %v", err)
	} else {
		logging.Info("清理完成")
	}
}
