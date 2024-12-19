package openstack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/BytemanD/easygo/pkg/compare"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/utility"
	"github.com/BytemanD/skyman/utility/httpclient"
	"github.com/go-resty/resty/v2"
)

type Glance struct {
	RestClient2
}
type ImageApi struct {
	Glance
}

func (o *Openstack) GlanceV2() *Glance {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.glanceClient == nil {
		endpoint, err := o.AuthPlugin.GetServiceEndpoint("image", "glance", "public")
		if err != nil {
			logging.Fatal("get glance endpoint falied: %v", err)
		}
		o.glanceClient = &Glance{
			NewRestClient2(utility.VersionUrl(endpoint, V2), o.AuthPlugin),
		}
	}
	return o.glanceClient
}
func (c Glance) GetCurrentVersion() (*model.ApiVersion, error) {
	resp, err := c.Index()
	if err != nil {
		return nil, err
	}
	body := struct{ Versions []model.ApiVersion }{}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	for _, version := range body.Versions {
		if version.Status == "CURRENT" {
			return &version, nil
		}
	}
	return nil, fmt.Errorf("current version not found")
}

func (c Glance) Images() ImageApi {
	return ImageApi{c}
}
func (c ImageApi) List(query url.Values, total int) ([]glance.Image, error) {
	images := []glance.Image{}
	limit := 0
	if query.Get("limit") != "" {
		limit, _ = strconv.Atoi(query.Get("limit"))
	}
	for {
		resp, err := c.Glance.Get("images", query)
		if err != nil {
			return nil, err
		}
		respBbody := glance.ImagesResp{}
		json.Unmarshal(resp.Body(), &respBbody)
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
func (c ImageApi) ListByName(name string) ([]glance.Image, error) {
	return c.List(utility.UrlValues(map[string]string{
		"name": name,
	}), 0)
}
func (c ImageApi) Show(id string) (*glance.Image, error) {
	resp, err := c.Glance.Get(utility.UrlJoin("images", id), nil)
	if err != nil {
		return nil, err
	}

	body := glance.Image{}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	return &body, nil
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
func (c ImageApi) Found(idOrName string) (*glance.Image, error) {
	var (
		image *glance.Image
		err   error
	)
	image, err = c.Show(idOrName)
	if err == nil {
		return image, nil
	}
	if compare.IsType[httpclient.HttpError](err) {
		httpError, _ := err.(httpclient.HttpError)
		if httpError.IsNotFound() {
			image, err = c.FoundByName(idOrName)
		}
	}
	return image, err
}
func (c ImageApi) Create(options glance.Image) (*glance.Image, error) {
	reqBody, err := json.Marshal(&options)
	if err != nil {
		return nil, err
	}
	resp, err := c.Glance.Post("images", reqBody, nil)
	if err != nil {
		return nil, err
	}
	body := glance.Image{}
	json.Unmarshal(resp.Body(), &body)
	return &body, nil
}
func (c ImageApi) Delete(id string) error {
	_, err := c.Glance.Delete(utility.UrlJoin("images", id), nil)
	return err
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
	resp, err := c.Glance.Patch(utility.UrlJoin("images", id), nil, attributies, headers)
	if err != nil {
		return nil, err
	}
	body := glance.Image{}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
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
	req := c.session.GetRequest(resty.MethodPut, c.makeUrl(utility.UrlJoin("images", id, "file"))).
		SetHeader(httpclient.CONTENT_TYPE, httpclient.CONTENT_TYPE_STREAM).
		SetBody(reader)
	_, err = c.session.Request(req)
	return err
}

func (c ImageApi) Download(id string, fileName string, process bool) error {
	image, err := c.Show(id)
	if err != nil {
		return err
	}
	req := c.session.GetRequest(resty.MethodGet, c.makeUrl(utility.UrlJoin("images", id, "file"))).
		SetHeader(httpclient.CONTENT_TYPE, "application/octet-stream").
		SetOutput(fileName)

	if utility.IsFileExists(fileName) {
		if err := os.Remove(fileName); err != nil {
			return err
		}
	}
	// TODO

	done := make(chan bool)
	defer close(done)

	go func() {
		_, err = c.session.Request(req)
		done <- true
	}()
	utility.WatchFileSize(fileName, int(image.Size))
	return err
}
