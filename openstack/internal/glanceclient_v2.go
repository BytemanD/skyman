package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/BytemanD/skyman/utility"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/wxnacy/wgo/file"
)

type GlanceV2 struct{ *ServiceClient }

const DEFAULT_IMAGE_LIMIT = 1000

func (c GlanceV2) ListWithTotal(query url.Values, total int) ([]glance.Image, error) {
	limit := DEFAULT_IMAGE_LIMIT
	if query.Has("limit") {
		limit, _ = strconv.Atoi(query.Get("limit"))
	}
	if total > 0 {
		limit = min(limit, total)
	}
	fixQuery := query
	images := []glance.Image{}
	for {
		fixQuery.Set("limit", strconv.Itoa(limit))
		console.Debug("query params: %s", fixQuery.Encode())
		respBbody := glance.ImagesResp{}
		_, err := c.R().SetQueryParamsFromValues(fixQuery).SetResult(&respBbody).Get(URL_IMAGES.F())
		if err != nil {
			return nil, err
		}
		if len(respBbody.Images) > 0 {
			images = append(images, respBbody.Images...)
		}
		if respBbody.Next == "" || (total > 0 && len(images) >= total) {
			break
		}
		parsedUrl, _ := url.Parse(respBbody.Next)
		fixQuery = parsedUrl.Query()
		if parsedUrl.Query().Has("limit") && total > 0 {
			limit, _ = strconv.Atoi(parsedUrl.Query().Get("limit"))
			limit = min(limit, total-len(images))
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
func (c GlanceV2) UpdateImage(id string, params map[string]any) (*glance.Image, error) {
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
	rawBody := map[string]any{}
	if err := json.Unmarshal(resp.Body(), &rawBody); err != nil {
		return nil, err
	}
	body.SetRaw(rawBody)
	return &body, nil
}

func (c GlanceV2) UploadImage(id string, file string, progress ...bool) error {
	fileReader, err := os.Open(file)
	if err != nil {
		return err
	}
	showProgress := lo.FirstOrEmpty(progress)
	defer fileReader.Close()

	fileStat, err := os.Stat(file)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	if showProgress {
		_, err = c.Clone().
			SetHeader(session.CONTENT_TYPE, session.CONTENT_TYPE_STREAM).
			SetPreRequestHook(func(c *resty.Client, r *http.Request) error {
				r.Body = utility.NewProcessReader(fileReader, fileStat.Size())
				return nil
			}).
			R().ForceContentType(session.CONTENT_TYPE).
			Put(URL_IMAGE_FILE.F(id))
		return err
	} else {
		_, err = c.R().SetHeader(session.CONTENT_TYPE, session.CONTENT_TYPE_STREAM).
			SetBody(fileReader).Put(URL_IMAGE_FILE.F(id))
		return err
	}

}

func (c GlanceV2) DownloadImage(id string, fileName string, process ...bool) error {
	showProgress := lo.FirstOrEmpty(process)
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
	if showProgress {
		if total, err := strconv.Atoi(resp.Header().Get(session.CONTENT_LENGTH)); err == nil {
			writer = utility.NewProgressWriter(imageWriter, total)
		} else {
			writer = imageWriter
		}
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
