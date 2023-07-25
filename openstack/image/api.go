package image

import (
	"net/url"
)

func (client ImageClientV2) ImageList(query url.Values) []Image {
	body := map[string][]Image{"images": {}}
	client.List("images", query, nil, &body)
	return body["images"]
}
func (client ImageClientV2) ImageListByName(name string) Images {
	query := url.Values{}
	query.Set("name", name)
	return client.ImageList(query)
}
func (client ImageClientV2) ImageShow(id string) (*Image, error) {
	image := Image{}
	err := client.Show("images", id, client.BaseHeaders, &image)
	if err != nil {
		return nil, err
	}
	return &image, nil
}
