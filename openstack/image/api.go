package image

import (
	"net/url"
	"strconv"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

func min(num1 int, num2 int) int {
	if num1 <= num2 {
		return num1
	} else {
		return num2
	}
}

func (client ImageClientV2) ImageList(query url.Values, total int) []Image {
	// body := map[string]interface{}{"images": []Image{}, "next": ""}
	images := []Image{}
	limit := 0
	if query.Get("limit") != "" {
		limit, _ = strconv.Atoi(query.Get("limit"))
	}
	for {
		body := ImagesResp{}
		client.List("images", query, nil, &body)
		if len(body.Images) == 0 {
			break
		}
		if total > 0 {
			images = append(images, body.Images[:min(total, len(body.Images))]...)
			if len(images) >= total {
				break
			}
		} else {
			images = append(images, body.Images...)
		}
		if body.Next == "" {
			break
		}
		logging.Info("next query url : %s", body.Next)
		parsedUrl, _ := url.Parse(body.Next)
		query = parsedUrl.Query()
		if limit > 0 && total > 0 && (total-len(images)) < limit {
			query.Set("limit", strconv.Itoa(total-len(images)))
		}
	}
	return images
}

func (client ImageClientV2) ImageListByName(name string) Images {
	query := url.Values{}
	query.Set("name", name)
	return client.ImageList(query, 0)
}
func (client ImageClientV2) ImageShow(id string) (*Image, error) {
	image := Image{}
	err := client.Show("images", id, client.BaseHeaders, &image)
	if err != nil {
		return nil, err
	}
	return &image, nil
}
