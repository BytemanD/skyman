package common

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/stackcrud/openstack/identity"
)

type VersionBody struct {
	Version Version `json:"version"`
}
type Version struct {
	MinVersion string `json:"min_version"`
	Version    string `json:"version"`
}
type ResourceClient struct {
	Endpoint    string
	AuthClient  identity.V3AuthClient
	Version     Version
	APiVersion  string
	BaseHeaders map[string]string
	baseUrl     string
}

func (client ResourceClient) getUrl(path ...string) string {
	if client.baseUrl == "" {
		if client.APiVersion != "" {
			client.baseUrl = client.Endpoint + "/" + client.APiVersion
		} else {
			client.baseUrl = client.Endpoint
		}
	}
	return strings.Join(append([]string{client.baseUrl}, path...), "/")
}
func (client ResourceClient) List(resource string, query url.Values, obj interface{}) error {
	resp, err1 := client.AuthClient.Get(client.getUrl(resource), query, client.BaseHeaders)
	if err1 != nil {
		logging.Error("request error,%s", err1)
		return fmt.Errorf("%s", resp.BodyString())
	}
	json.Unmarshal(resp.Body, &obj)
	return nil
}
func (client ResourceClient) Show(resource string, id string, query url.Values, obj interface{}) error {
	resp, err1 := client.AuthClient.Get(client.getUrl(resource, id), query, client.BaseHeaders)
	if err2 := resp.JudgeStatus(); err2 != nil {
		return fmt.Errorf("%s", resp.BodyString())
	}
	json.Unmarshal(resp.Body, &obj)
	return err1
}
func (client ResourceClient) Delete(resource string, id string) error {
	resp, err1 := client.AuthClient.Delete(client.getUrl(resource, id), client.BaseHeaders)
	if err2 := resp.JudgeStatus(); err2 != nil {
		return fmt.Errorf("%s", resp.BodyString())
	}
	return err1
}
func (client ResourceClient) Create(resource string, body []byte, obj interface{}) error {
	resp, err1 := client.AuthClient.Post(
		client.getUrl(resource), body, client.BaseHeaders)
	json.Unmarshal(resp.Body, &obj)
	if err2 := resp.JudgeStatus(); err2 != nil {
		return fmt.Errorf("%s", resp.BodyString())
	}
	return err1
}

func (client ResourceClient) EndpointHasVersion() bool {
	found, _ := regexp.MatchString(".*/v[0-9.]+", client.Endpoint)
	return found
}

func (client ResourceClient) GetVersionEndpoint() string {
	if !client.EndpointHasVersion() && client.APiVersion != "" {
		return client.Endpoint + "/" + client.APiVersion
	}
	return client.Endpoint
}
