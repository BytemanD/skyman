package image

import (
	"encoding/json"
	"net/url"
)

type Image struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DiskFormat      string `json:"disk_format"`
	ContainerFormat string `json:"container_format"`
	Size            uint   `json:"size"`
	Status          string `json:"status"`
}

type ImagesBody struct {
	Images []Image `json:"images"`
}

func (client ImageClientV2) ImageList(query url.Values) []Image {

	resp, _ := client.AuthClient.Get(
		client.getUrl("images", ""), query, client.BaseHeaders)
	imagesBody := ImagesBody{}
	json.Unmarshal(resp.Body, &imagesBody)
	return imagesBody.Images
}
