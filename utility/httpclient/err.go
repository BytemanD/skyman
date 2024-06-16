package httpclient

import "fmt"

const (
	CODE_404 = 404
)

type HttpError struct {
	Status  int
	Reason  string
	Message string
}

func (err HttpError) Error() string {
	return fmt.Sprintf("%d %s", err.Status, err.Reason)
}

func (err HttpError) IsNotFound() bool {
	return err.Status == CODE_404
}
