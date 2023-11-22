package image

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/common"
)

func min(numbers ...int) int {
	minNumber := numbers[0]
	for _, number := range numbers[1:] {
		if number < minNumber {
			minNumber = number
		}
	}
	return minNumber
}
func (client ImageClientV2) newRequest(resource string, id string, query url.Values, body []byte) common.RestfulRequest {
	return common.RestfulRequest{
		Endpoint: client.endpoint,
		Resource: resource, Id: id,
		Query:   query,
		Body:    body,
		Headers: client.BaseHeaders}
}

func (client ImageClientV2) ImageList(query url.Values, total int) ([]Image, error) {
	images := []Image{}
	limit := 0
	if query.Get("limit") != "" {
		limit, _ = strconv.Atoi(query.Get("limit"))
	}
	for {
		resp, err := client.Request(client.newRequest("images", "", query, nil))
		if err != nil {
			return nil, err
		}
		respBbody := ImagesResp{}
		resp.BodyUnmarshal(&respBbody)
		if len(respBbody.Images) == 0 {
			break
		}
		if total > 0 {
			images = append(images, respBbody.Images[:min(total, len(respBbody.Images))]...)
			if len(images) >= total {
				break
			}
		} else {
			images = append(images, respBbody.Images...)
		}
		if respBbody.Next == "" {
			break
		}
		logging.Info("next query url : %s", respBbody.Next)
		parsedUrl, _ := url.Parse(respBbody.Next)
		query = parsedUrl.Query()
		if limit > 0 && total > 0 && (total-len(images)) < limit {
			query.Set("limit", strconv.Itoa(total-len(images)))
		}
	}
	return images, nil
}

func (client ImageClientV2) ImageListByName(name string) (Images, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.ImageList(query, 0)
}
func (client ImageClientV2) ImageShow(id string) (*Image, error) {
	image := Image{}
	resp, err := client.Request(client.newRequest("images", id, nil, nil))
	if err != nil {
		return nil, err
	}
	resp.BodyUnmarshal(&image)
	return &image, nil
}
func (client ImageClientV2) ImageFound(idOrName string) (*Image, error) {
	var (
		image *Image
		err   error
	)
	image, err = client.ImageShow(idOrName)
	if err == nil {
		return image, nil
	}
	if httpError, ok := err.(common.HttpError); ok {
		if httpError.IsNotFound() {
			var images []Image
			if images, err = client.ImageListByName(idOrName); err != nil {
				return nil, err
			}
			if len(images) == 0 {
				return nil, fmt.Errorf("image %s not found", idOrName)
			}
			image, err = client.ImageShow(images[0].Id)
		}
	}
	return image, err
}
