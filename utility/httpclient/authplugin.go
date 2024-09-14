package httpclient

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

type AuthPluginInterface interface {
	AuthRequest(req *resty.Request) error
	GetSafeHeader(header http.Header) http.Header
	IsAdmin() bool
}

type BasicAuthPlugin struct {
	UserName string
	Password string
	Admin    bool
}

func (a BasicAuthPlugin) AuthRequest(req *resty.Request) error {
	req.SetBasicAuth(a.UserName, a.Password)
	return nil
}
func (a BasicAuthPlugin) GetSafeHeader(header http.Header) http.Header {
	return header
}
func (a BasicAuthPlugin) IsAdmin() bool {
	return a.Admin
}
