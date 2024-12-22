package internal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/utility"
	"github.com/BytemanD/skyman/utility/httpclient"
)

type ImageApi struct{ ResourceApi }

func (c ImageApi) List(query url.Values, total int) ([]glance.Image, error) {
	images := []glance.Image{}
	fixQuery := query
	if !fixQuery.Has("limit") {
		if total > 0 {
			fixQuery.Set("limit", strconv.Itoa(total))
		}
	} else if total > 0 {
		limit, _ := strconv.Atoi(query.Get("limit"))
		fixQuery.Set("limit", strconv.Itoa(min(limit, total)))
	}
	for {
		respBbody := glance.ImagesResp{}
		req := c.NewGetRequest("images", fixQuery, nil)
		logging.Info("query params: %s", req.QueryParam.Encode())
		_, err := req.SetResult(&respBbody).Send()
		if err != nil {
			return nil, err
		}
		if len(respBbody.Images) == 0 {
			break
		}
		images = append(images, respBbody.Images...)
		if len(images) >= total {
			break
		}
		if respBbody.Next == "" {
			break
		}
		parsedUrl, _ := url.Parse(respBbody.Next)
		fixQuery = parsedUrl.Query()
		if parsedUrl.Query().Has("limit") && total > 0 {
			limit, _ := strconv.Atoi(parsedUrl.Query().Get("limit"))
			limit = min(limit, total-len(images))
			fixQuery.Set("limit", strconv.Itoa(limit))
		}
	}
	return images, nil
}
func (c ImageApi) ListAll(query url.Values) ([]glance.Image, error) {
	return c.List(query, 0)
}
func (c ImageApi) ListByName(name string) ([]glance.Image, error) {
	return c.List(url.Values{"name": []string{name}}, 0)
}
func (c ImageApi) Show(id string) (*glance.Image, error) {
	result := glance.Image{}
	_, err := c.Get("images/"+id, nil, &result)
	return &result, err
}
func (c ImageApi) FoundByName(name string) (*glance.Image, error) {
	images, err := c.ListByName(name)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("image %s not found", name)
	}
	if len(images) > 1 {
		return nil, fmt.Errorf("found multi images named %s ", name)
	}
	return c.Show(images[0].Id)
}
func (c ImageApi) Find(idOrName string) (*glance.Image, error) {
	return FindResource(idOrName, c.Show, c.ListAll)
}
func (c ImageApi) Delete(id string) error {
	_, err := c.ResourceApi.Delete("images/" + id)
	return err
}
func (c ImageApi) Create(options glance.Image) (*glance.Image, error) {
	respBody := glance.Image{}
	_, err := c.Post("images", options, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil

}
func (c ImageApi) Set(id string, params map[string]interface{}) (*glance.Image, error) {
	attributies := []glance.AttributeOp{}
	for k, v := range params {
		attributies = append(attributies, glance.AttributeOp{
			Path:  fmt.Sprintf("/%s", k),
			Value: v,
			Op:    "replace",
		})
	}
	headers := map[string]string{
		"Content-Type": "application/openstack-images-v2.1-json-patch",
	}
	body := glance.Image{}
	resp, err := c.Patch("images/"+id, attributies, &body, headers)
	if err != nil {
		return nil, err
	}
	rawBody := map[string]interface{}{}
	if err := json.Unmarshal(resp.Body(), &rawBody); err != nil {
		return nil, err
	}

	body.SetRaw(rawBody)
	return &body, nil
}

func (c ImageApi) Upload(id string, file string) error {
	fileStat, err := os.Stat(file)
	if err != nil {
		return err
	}
	fileReader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	// TODO:
	reader := utility.NewProcessReader(fileReader, int(fileStat.Size()))
	_, err = c.NewPutRequest(utility.UrlJoin("images", id, "file"), reader, nil).
		SetHeader(httpclient.CONTENT_TYPE, httpclient.CONTENT_TYPE_STREAM).
		Send()

	// _, err := c.NewPutRequest(utility.UrlJoin("images", id, "file"), nil, nil).
	// 	SetHeader(httpclient.CONTENT_TYPE, httpclient.CONTENT_TYPE_STREAM).
	// 	SetFile("image", file).
	// 	Send()
	return err
}

func (c ImageApi) Download(id string, fileName string, process bool) error {
	image, err := c.Show(id)
	if err != nil {
		return err
	}
	req := c.NewGetRequest(utility.UrlJoin("images", id, "file"), nil, nil).
		SetHeader(httpclient.CONTENT_TYPE, "application/octet-stream").
		SetOutput(fileName)

	if utility.IsFileExists(fileName) {
		if err := os.Remove(fileName); err != nil {
			return err
		}
	}
	// TODO: 异常
	go req.Send()
	utility.WatchFileSize(fileName, int(image.Size))
	return err
}

type GlanceV2 struct{ *ServiceClient }

func (c GlanceV2) Images() ImageApi {
	return ImageApi{
		ResourceApi: ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint},
	}
}

func (c GlanceV2) GetCurrentVersion() (*model.ApiVersion, error) {
	respBody := struct{ Versions []model.ApiVersion }{}
	if _, err := c.Index(&respBody); err != nil {
		return nil, err
	}

	for _, version := range respBody.Versions {
		if version.Status == "CURRENT" {
			return &version, nil
		}
	}
	return nil, fmt.Errorf("current version not found")
}
