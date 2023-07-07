package image

import (
	"net/url"
)

func (client ImageClientV2) ImageList(query url.Values) Images {
	imagesBody := ImagesBody{}
	client.List("images", query, &imagesBody)
	return imagesBody.Images
}
func (client ImageClientV2) ImageListByName(name string) Images {
	query := url.Values{}
	query.Set("name", name)
	return client.ImageList(query)
}
func (client ImageClientV2) ImageShow(id string) (*Image, error) {
	image := Image{}
	err := client.Show("images", id, nil, &image)
	if err != nil {
		return nil, err
	}
	return &image, nil
}
