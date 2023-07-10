package storage

import "net/url"

func (client StorageClientV2) VolumeList(query url.Values) Volumes {
	body := VolumesBody{}
	client.List("volumes", query, &body)
	return body.Volumes
}
func (client StorageClientV2) VolumeListDetail(query url.Values) Volumes {
	body := VolumesBody{}
	client.List("volumes/detail", query, &body)
	return body.Volumes
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
