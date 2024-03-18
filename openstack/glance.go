package openstack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/utility"
)

type Glance struct {
	RestClient
}
type ImageApi struct {
	Glance
}

func (o Openstack) GlanceV2() Glance {
	endpoint, err := o.AuthPlugin.GetServiceEndpoint("image", "glance", "public")
	if err != nil {
		logging.Fatal("get compute endpoint falied: %v", err)
	}
	return Glance{
		RestClient{BaseUrl: utility.VersionUrl(endpoint, "v2"), AuthPlugin: o.AuthPlugin},
	}
}
func (c Glance) GetCurrentVersion() (*model.ApiVersion, error) {
	resp, err := c.Index()
	if err != nil {
		return nil, err
	}
	apiVersions := struct{ Versions []model.ApiVersion }{}
	if err := resp.BodyUnmarshal(&apiVersions); err != nil {
		return nil, err
	}
	for _, version := range apiVersions.Versions {
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

	image := glance.Image{}
	if err := resp.BodyUnmarshal(&image); err != nil {
		return nil, err
	}
	return &image, nil
}
func (c ImageApi) FoundByName(name string) (*glance.Image, error) {
	query := url.Values{}
	query.Set("name", name)
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
	if httpError, ok := err.(*utility.HttpError); ok {
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
	image := glance.Image{}
	resp.BodyUnmarshal(&image)
	return &image, nil
}
func (c ImageApi) Delete(id string) (err error) {
	_, err = c.Glance.Delete(utility.UrlJoin("images", id), nil)
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
	body, _ := json.Marshal(attributies)

	req := utility.Request{
		Url:  utility.UrlJoin("images", id),
		Body: body,
		Headers: map[string]string{
			"Content-Type": "application/openstack-images-v2.1-json-patch",
		},
	}
	resp, err := c.Glance.Patch(req)
	if err != nil {
		return nil, err
	}
	image := glance.Image{}
	resp.BodyUnmarshal(&image)
	rawBody := map[string]interface{}{}
	resp.BodyUnmarshal(&rawBody)
	image.SetRaw(rawBody)
	return &image, nil
}
func (c ImageApi) Upload(id string, fileName string) error {
	// headers := utility.CloneHeaders(client.BaseHeaders,
	// 	map[string]string{"Content-Type": "application/octet-stream"},
	// )

	// file, err := os.Open(fileName)
	// if err != nil {
	// 	return err
	// }
	// fileStat, err := file.Stat()
	// buffer := make([]byte, fileStat.Size())
	// file.Read(buffer)

	// req := common.NewResourcePutRequest(client.endpoint, "images", id+"/file", buffer, headers)
	// req.ShowProcess = true
	// _, err = client.Request(req)
	// return err
	return fmt.Errorf("not impleted")
}

func (c ImageApi) Download(id string, fileName string, process bool) error {
	// headers := utility.CloneHeaders(client.BaseHeaders,
	// 	map[string]string{"Content-Type": "application/octet-stream"},
	// )
	// file, err := os.Create(fileName)
	// if err != nil {
	// 	return err
	// }
	// req := common.NewResourceShowRequest(client.endpoint, "images", id+"/file", headers)
	// req.ShowProcess = true
	// resp, err := client.Request(req)
	// if err != nil {
	// 	return err
	// }
	// resp.SaveBody(file, process)
	return fmt.Errorf("not impleted")
}
