package httpclient

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/BytemanD/skyman/utility"
)

type Response struct {
	Status     int
	Reason     string
	Body       []byte
	Headers    http.Header
	bodyReader io.ReadCloser
}

func (resp *Response) SetBodyReader(reader io.ReadCloser) {
	resp.bodyReader = reader
}

func (resp Response) BodyString() string {
	return string(resp.Body)
}

func (resp *Response) ReadAll() error {
	var err error
	resp.Body, err = io.ReadAll(resp.bodyReader)
	defer resp.bodyReader.Close()
	return err
}

func (resp Response) GetHeader(key string) string {
	return resp.Headers.Get(key)
}

func (resp Response) BodyUnmarshal(object interface{}) error {
	return json.Unmarshal(resp.Body, object)
}
func (resp Response) GetContentLength() int {
	length, _ := strconv.Atoi(resp.Headers.Get("Content-Length"))
	return length
}
func (resp Response) SaveBody(file *os.File, process bool) error {
	defer resp.bodyReader.Close()
	var reader io.Reader
	if process && resp.GetContentLength() > 0 {
		reader = utility.NewProcessReader(resp.bodyReader, resp.GetContentLength())
	} else {
		reader = bufio.NewReaderSize(resp.bodyReader, 1024*32)
	}

	_, err := io.Copy(bufio.NewWriter(file), reader)
	return err
}

func (resp Response) IsNotFound() bool {
	return resp.Status == 404
}
func (resp *Response) MustNotError() error {
	switch {
	case resp.Status < 400:
		return nil
	default:
		return &HttpError{Status: resp.Status, Reason: resp.Reason,
			Message: resp.BodyString()}
	}
}
