package result

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/BytemanD/skyman/openstack/session"
	"github.com/go-resty/resty/v2"
)

type Server struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func getItemsResult() ItemsResult[Server] {
	reqBody, err := json.Marshal(struct {
		Servers []Server `json:"servers"`
	}{
		Servers: []Server{
			{Id: "1111", Name: "foo"},
			{Id: "2222", Name: "bar"},
		},
	})
	rawResp := session.Response{
		Response: &resty.Response{RawResponse: &http.Response{}}}

	rawResp.SetBody(reqBody)
	result := NewItemsResult[Server](&rawResp, err)
	result.SetKey("servers")
	return *result
}

func TestItemsResult(t *testing.T) {
	result := getItemsResult()
	if servers, err := result.Items(); err != nil {
		t.Error(err)
	} else if len(servers) != 2 {
		t.Errorf("expect 2 items, but got %d", len(servers))
	}
}
