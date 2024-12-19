package internal

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model/glance"
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

type GlanceV2 ServiceClient

func (c GlanceV2) Images() ImageApi {
	return ImageApi{
		ResourceApi: ResourceApi{
			Client:  c.rawClient,
			BaseUrl: c.Endpoint,
		},
	}
}
