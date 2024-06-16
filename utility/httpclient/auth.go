package httpclient

import (
	"net/http"
)

type AuthInterface interface {
	AuthRequest(req *http.Request) error
	GetSafeHeader(header http.Header) http.Header
}

type BasicAuth struct {
	UserName string
	Password string
}

func (a BasicAuth) AuthRequest(req *http.Request) error {
	req.SetBasicAuth(a.UserName, a.Password)
	return nil
}
func (a BasicAuth) GetSafeHeader(header http.Header) http.Header {
	return header
}
