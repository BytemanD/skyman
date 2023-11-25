package storage

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/openstack/common"
)

const (
	POLICY_NEVER     = "never"
	POLICY_ON_DEMAND = "on-demand"
)

var MIGRATION_POLICYS = []string{POLICY_NEVER, POLICY_ON_DEMAND}

func InvalidMIgrationPoicy(policy string) error {
	if !common.ContainsString(MIGRATION_POLICYS, policy) {
		return fmt.Errorf("Invalid migration policy: %s, supported: %s", policy, MIGRATION_POLICYS)
	}
	return nil
}

func (client StorageClientV2) VolumeList(query url.Values) ([]Volume, error) {
	resp, err := client.Request(
		common.NewResourceListRequest(client.endpoint, "volumes", query, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]Volume{"volumes": {}}
	resp.BodyUnmarshal(&body)
	return body["volumes"], nil
}
func (client StorageClientV2) VolumeListByName(name string) ([]Volume, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.VolumeList(query)
}
func (client StorageClientV2) VolumeListDetail(query url.Values) ([]Volume, error) {
	resp, err := client.Request(
		common.NewResourceListRequest(client.endpoint, "volumes/detail", query, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]Volume{"volumes": {}}
	resp.BodyUnmarshal(&body)
	return body["volumes"], nil
}
func (client StorageClientV2) VolumeListDetailByName(name string) (Volumes, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.VolumeListDetail(query)
}

func (client StorageClientV2) VolumeShow(id string) (*Volume, error) {
	resp, err := client.Request(
		common.NewResourceShowRequest(client.endpoint, "volumes", id, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Volume{"volume": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["volume"], nil
}
func (client StorageClientV2) VolumeFound(idOrName string) (*Volume, error) {
	volume, err := client.VolumeShow(idOrName)
	if err == nil {
		return volume, nil
	}
	if httpError, ok := err.(*common.HttpError); ok {
		if !httpError.IsNotFound() {
			return nil, err
		}
	}
	volumes, err := client.VolumeListDetailByName(idOrName)
	if err != nil {
		return nil, err
	}
	if len(volumes) == 0 {
		return nil, fmt.Errorf("volume %s not found", idOrName)
	} else if len(volumes) == 1 {
		return &(volumes[0]), nil
	} else {
		return nil, fmt.Errorf("Found multi volumes named %s", idOrName)
	}
}
func (client StorageClientV2) VolumeCreate(params map[string]interface{}) (*Volume, error) {
	data := map[string]interface{}{"volume": params}
	body, _ := json.Marshal(data)
	resp, err := client.Request(
		common.NewResourceCreateRequest(client.endpoint, "volumes", body, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Volume{"volume": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["volume"], nil
}
func (client StorageClientV2) doAction(id, data interface{}, resp interface{}) error {
	reqBody, _ := json.Marshal(data)
	_, err := client.Request(
		common.NewResourceCreateRequest(client.endpoint, "volumes", reqBody, client.BaseHeaders),
	)
	if err != nil {
		return err
	}
	return nil
}
func (client StorageClientV2) VolumeExtend(id string, size int) error {
	repData := map[string]map[string]interface{}{
		"os-extend": {
			"new_size": size,
		},
	}
	return client.doAction(id, repData, nil)
}
func (client StorageClientV2) VolumeRetype(id string, newType string, migrationPolicy string) error {
	repData := map[string]map[string]interface{}{
		"os-retype": {
			"new_type":         newType,
			"migration_policy": migrationPolicy,
		},
	}
	return client.doAction(id, repData, nil)
}

func (client StorageClientV2) VolumeDelete(id string) error {
	_, err := client.Request(
		common.NewResourceDeleteRequest(client.endpoint, "volumes", id, client.BaseHeaders),
	)
	return err
}

// volume type api
func (client StorageClientV2) VolumeTypeList(query url.Values) ([]VolumeType, error) {
	resp, err := client.Request(
		common.NewResourceListRequest(client.endpoint, "types", query, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]VolumeType{"volume_types": {}}
	resp.BodyUnmarshal(&body)
	return body["volume_types"], nil
}
func (client StorageClientV2) VolumeTypeListByName(name string) ([]VolumeType, error) {
	volumeTypes, err := client.VolumeTypeList(nil)
	if err != nil {
		return nil, err
	}
	foundVolumeTypes := []VolumeType{}
	for _, volumeType := range volumeTypes {
		if volumeType.Name == name {
			foundVolumeTypes = append(foundVolumeTypes, volumeType)
		}
	}
	return foundVolumeTypes, nil
}
func (client StorageClientV2) VolumeTypeDefaultGet() (*VolumeType, error) {
	resp, err := client.Request(
		common.NewResourceListRequest(client.endpoint, "types/default", nil, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*VolumeType{"volume_type": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["volume_type"], nil
}
func (client StorageClientV2) VolumeTypeShow(id string) (*VolumeType, error) {
	resp, err := client.Request(
		common.NewResourceShowRequest(client.endpoint, "types", id, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*VolumeType{"volume_type": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["volume_type"], nil
}
func (client StorageClientV2) VolumeTypeFound(idOrName string) (*VolumeType, error) {
	volumeType, err := client.VolumeTypeShow(idOrName)
	if err == nil {
		return volumeType, nil
	}
	if httpError, ok := err.(*common.HttpError); ok {
		if !httpError.IsNotFound() {
			return nil, err
		}
	}
	volumeTypes, err := client.VolumeTypeListByName(idOrName)
	if err != nil {
		return nil, err
	}
	if len(volumeTypes) == 0 {
		return nil, fmt.Errorf("volume type %s not found", idOrName)
	} else if len(volumeTypes) == 1 {
		return &(volumeTypes[0]), nil
	} else {
		return nil, fmt.Errorf("Found multi volume types named %s", idOrName)
	}
}
func (client StorageClientV2) VolumeTypeCreate(params map[string]interface{}) (*VolumeType, error) {
	data := map[string]interface{}{"volume_type": params}
	body, _ := json.Marshal(data)
	resp, err := client.Request(
		common.NewResourceCreateRequest(client.endpoint, "volume_types", body, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*VolumeType{"volume_type": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["volume_type"], nil
}
func (client StorageClientV2) VolumeTypeDelete(id string) error {
	_, err := client.Request(
		common.NewResourceDeleteRequest(client.endpoint, "volume_types", id, client.BaseHeaders),
	)
	return err
}
