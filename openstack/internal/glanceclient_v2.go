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

type GlanceV2 struct{ *ServiceClient }

func (c GlanceV2) ListWithTotal(query url.Values, total int) ([]glance.Image, error) {
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
		_, err := c.R().SetQueryParamsFromValues(fixQuery).SetResult(&respBbody).Get(URL_IMAGES.F())
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
func (c GlanceV2) ListImage(query url.Values) ([]glance.Image, error) {
	return c.ListWithTotal(query, 0)
}

func (c GlanceV2) GetImage(id string) (*glance.Image, error) {
	return GetResource[glance.Image](c.ServiceClient, URL_IMAGE.F(id), "")
}
func (c GlanceV2) FoundByName(name string) (*glance.Image, error) {
	images, err := c.ListImage(url.Values{"name": []string{name}})
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("image %s not found", name)
	}
	if len(images) > 1 {
		return nil, fmt.Errorf("found multi images named %s ", name)
	}
	return c.GetImage(images[0].Id)
}
func (c GlanceV2) FindImage(idOrName string) (*glance.Image, error) {
	return QueryByIdOrName(idOrName, c.GetImage, c.ListImage)
}

func (c GlanceV2) DeleteImage(id string) error {
	return DeleteResource(c.ServiceClient, URL_IMAGE.F(id))
}
func (c GlanceV2) CreateImage(options glance.Image) (*glance.Image, error) {
	respBody := glance.Image{}
	_, err := c.R().SetBody(options).SetResult(&respBody).Post(URL_IMAGES.F())
	return &respBody, err

}
func (c GlanceV2) UpdateImage(id string, params map[string]interface{}) (*glance.Image, error) {
	attributies := []glance.AttributeOp{}
	for k, v := range params {
		attributies = append(attributies, glance.AttributeOp{
			Path:  fmt.Sprintf("/%s", k),
			Value: v,
			Op:    "replace",
		})
	}
	headers := map[string][]string{
		"Content-Type": {"application/openstack-images-v2.1-json-patch"},
	}
	body := glance.Image{}
	resp, err := c.R().SetHeaderMultiValues(headers).SetBody(attributies).
		SetResult(&body).Patch(URL_IMAGE.F(id))
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

func (c GlanceV2) UploadImage(id string, file string) error {
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
	_, err = c.R().SetHeader(session.CONTENT_TYPE, session.CONTENT_TYPE_STREAM).
		SetBody(reader).Put(URL_IMAGE_FILE.F(id))
	return err
}

func (c GlanceV2) DownloadImage(id string, fileName string, process bool) error {
	if file.IsFile(fileName) {
		return fmt.Errorf("file %s exists", fileName)
	}
	resp, err := c.Client.R().SetDoNotParseResponse(true).
		SetHeader(session.CONTENT_TYPE, session.CONTENT_TYPE_STREAM).
		Get(URL_IMAGE_FILE.F(id))
	if err != nil {
		return err
	}
	imageWriter, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer imageWriter.Close()
	var writer io.Writer
	if total, err := strconv.Atoi(resp.Header().Get(session.CONTENT_LENGTH)); err == nil {
		writer = utility.NewPbrWriter(total, imageWriter)
	} else {
		writer = imageWriter
	}
	_, err = io.Copy(writer, resp.RawBody())
	return err
}

func (c GlanceV2) GetCurrentVersion() (*model.ApiVersion, error) {
	result := struct{ Versions model.ApiVersions }{}
	if _, err := c.Index(&result); err != nil {
		return nil, err
	}
	return &result.Versions[0], nil
}
