package httpclient

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

type AuthPluginInterface interface {
	AuthRequest(req *resty.Request) error
	GetSafeHeader(header http.Header) http.Header
}

type BasicHttpAuth struct {
	UserName string
	Password string
}

func (a BasicHttpAuth) AuthRequest(req *resty.Request) error {
	req.SetBasicAuth(a.UserName, a.Password)
	return nil
}
func (a BasicHttpAuth) GetSafeHeader(header http.Header) http.Header {
	return header
}
