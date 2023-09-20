package storage

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/openstack/common"
)

func (client StorageClientV2) VolumeList(query url.Values) []Volume {
	body := map[string][]Volume{"volumes": {}}
	client.List("volumes", query, nil, &body)
	return body["volumes"]
}
func (client StorageClientV2) VolumeListByName(name string) []Volume {
	query := url.Values{}
	query.Set("name", name)
	return client.VolumeList(query)
}
func (client StorageClientV2) VolumeListDetail(query url.Values) (Volumes, error) {
	body := map[string][]Volume{"volumes": {}}
	err := client.List("volumes/detail", query, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["volumes"], nil
}
func (client StorageClientV2) VolumeListDetailByName(name string) (Volumes, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.VolumeListDetail(query)
}

func (client StorageClientV2) VolumeShow(id string) (*Volume, error) {
	body := map[string]*Volume{"volume": {}}
	err := client.Show("volumes", id, client.BaseHeaders, &body)
	return body["volume"], err
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
	respBody := map[string]*Volume{"volume": {}}
	err := client.Create("volumes", body, client.BaseHeaders, &respBody)
	if err != nil {
		return nil, err
	}
	return respBody["volume"], nil
}
func (client StorageClientV2) VolumeDelete(id string) error {
	return client.Delete("volumes", id, client.BaseHeaders)
}

// func (client CinderClientV2) ImageShow(id string) (*Volume, error) {
// 	volume := Volume{}
// 	err := client.Show("images", id, nil, &volume)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &volume, nil
// }
