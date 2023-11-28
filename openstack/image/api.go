package image

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
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
		resp, err := client.Request(
			common.NewResourceListRequest(client.endpoint, "images", query, client.BaseHeaders),
		)
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
	resp, err := client.Request(
		common.NewResourceShowRequest(client.endpoint, "images", id, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	resp.BodyUnmarshal(&image)
	rawBody := map[string]interface{}{}
	resp.BodyUnmarshal(&rawBody)
	image.raw = rawBody
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
	if httpError, ok := err.(*common.HttpError); ok {
		if httpError.IsNotFound() {
			var images []Image
			if images, err = client.ImageListByName(idOrName); err != nil {
				return nil, err
			}
			if len(images) == 0 {
				return nil, fmt.Errorf("image %s not found", idOrName)
			}
			if len(images) > 1 {
				return nil, fmt.Errorf("found multi images named %s ", idOrName)
			}
			image, err = client.ImageShow(images[0].Id)
		}
	}
	return image, err
}

func (client ImageClientV2) ImageCreate(options Image) (*Image, error) {
	// params := map[string]interface{}{"name": name}
	reqBody, _ := json.Marshal(&options)
	resp, err := client.Request(
		common.NewResourceCreateRequest(client.endpoint, "images", reqBody, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	image := Image{}
	resp.BodyUnmarshal(&image)
	return &image, nil
}
func (client ImageClientV2) ImageUpload(id string, fileName string) error {
	headers := common.CloneHeaders(client.BaseHeaders,
		map[string]string{"Content-Type": "application/octet-stream"},
	)

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	fileStat, err := file.Stat()
	buffer := make([]byte, fileStat.Size())
	file.Read(buffer)

	req := common.NewResourcePutRequest(client.endpoint, "images", id+"/file", buffer, headers)
	req.ShowProcess = true
	_, err = client.Request(req)
	return err
}
func (client ImageClientV2) ImageDelete(id string) error {
	_, err := client.Request(
		common.NewResourceDeleteRequest(client.endpoint, "images", id, client.BaseHeaders),
	)
	return err
}

func (client ImageClientV2) ImageDownload(id string, fileName string, process bool) error {
	headers := common.CloneHeaders(client.BaseHeaders,
		map[string]string{"Content-Type": "application/octet-stream"},
	)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	req := common.NewResourceShowRequest(client.endpoint, "images", id+"/file", headers)
	req.ShowProcess = true
	resp, err := client.Request(req)
	if err != nil {
		return err
	}
	resp.SaveBody(file, process)
	return err
}
