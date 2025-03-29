package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/BytemanD/skyman/utility"
	"github.com/wxnacy/wgo/file"
)

type ImageApi struct{ ResourceApi }

func (c ImageApi) ListWithTotal(query url.Values, total int) ([]glance.Image, error) {
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
		console.Debug("query params: %s", query.Encode())
		_, err := c.R().SetQuery(fixQuery).SetResult(&respBbody).Get("images")
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
func (c ImageApi) List(query url.Values) ([]glance.Image, error) {
	return c.ListWithTotal(query, 0)
}

func (c ImageApi) ListByName(name string) ([]glance.Image, error) {
	return c.ListWithTotal(url.Values{"name": []string{name}}, 0)
}
func (c ImageApi) Show(id string) (*glance.Image, error) {
	result := glance.Image{}
	_, err := c.R().SetResult(&result).Get("images/" + id)
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
	return FindIdOrName(c, idOrName)
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
		SetHeader(session.CONTENT_TYPE, session.CONTENT_TYPE_STREAM).
		Send()

	return err
}

func (c ImageApi) Download(id string, fileName string, process bool) error {
	if file.IsFile(fileName) {
		return fmt.Errorf("file %s exists", fileName)
	}
	urlFormat := "images/%s/file"
	resp, err := c.Client.R().
		SetHeader(session.CONTENT_TYPE, session.CONTENT_TYPE_STREAM).
		SetDoNotParseResponse(true).
		Get(fmt.Sprintf(urlFormat, id))
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	imageWriter, err := os.Create(fileName)
	if err != nil {
		return err
	}
	var writers io.Writer
	if total, err := strconv.Atoi(resp.Header().Get(session.CONTENT_LENGTH)); err == nil {
		writers = io.MultiWriter(imageWriter, utility.NewPbrWriter(total))
	} else {
		writers = imageWriter
	}
	_, err = io.Copy(writers, resp.RawBody())
	return err
}

type GlanceV2 struct{ *ServiceClient }

func (c GlanceV2) Images() ImageApi {
	return ImageApi{
		ResourceApi: ResourceApi{
			Client:      c.Client,
			ResourceUrl: "images",
		},
	}
}

func (c GlanceV2) GetCurrentVersion() (*model.ApiVersion, error) {
	result := struct{ Versions model.ApiVersions }{}
	if _, err := c.Index(&result); err != nil {
		return nil, err
	}
	return &result.Versions[0], nil
}
