package storage

import "net/url"

func (client StorageClientV2) VolumeList(query url.Values) []Volume {
	body := map[string][]Volume{"volumes": {}}
	client.List("volumes", query, nil, &body)
	return body["volumes"]
}
func (client StorageClientV2) VolumeListDetail(query url.Values) Volumes {
	body := map[string][]Volume{"volumes": {}}
	client.List("volumes/detail", query, nil, &body)
	return body["volumes"]
}
func (client StorageClientV2) VolumeListDetailByName(name string) Volumes {
	query := url.Values{}
	query.Set("name", name)
	return client.VolumeListDetail(query)
}
func (client StorageClientV2) VolumeShow(id string) (*Volume, error) {
	body := map[string]*Volume{"volume": {}}
	err := client.Show("volumes", id, client.BaseHeaders, &body)
	return body["volume"], err
}

// func (client CinderClientV2) VolumeListByName(name string) []Volumes {
// 	query := url.Values{}
// 	query.Set("name", name)
// 	return client.ImageList(query)
// }
// func (client CinderClientV2) ImageShow(id string) (*Volume, error) {
// 	volume := Volume{}
// 	err := client.Show("images", id, nil, &volume)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &volume, nil
// }
