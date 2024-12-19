package internal

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/BytemanD/easygo/pkg/compare"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/utility/httpclient"
)

type ImageApi struct{ ResourceApi }

func (c ImageApi) List(query url.Values, total int) ([]glance.Image, error) {
	images := []glance.Image{}
	limit := 0
	if query.Get("limit") != "" {
		limit, _ = strconv.Atoi(query.Get("limit"))
	}
	req := c.NewGetRequest("images", query, nil)
	for {
		respBbody := glance.ImagesResp{}
		_, err := req.SetResult(&respBbody).Send()
		if err != nil {
			return nil, err
		}
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
func (c ImageApi) Delete(id string) error {
	_, err := c.ResourceApi.Delete("images/" + id)
	return err
}

type GlanceV2 ServiceClient

func (c GlanceV2) Images() ImageApi {
	return ImageApi{
		ResourceApi: ResourceApi{
			Client:  c.rawClient,
			BaseUrl: c.Endpoint,
		},
	}
}
