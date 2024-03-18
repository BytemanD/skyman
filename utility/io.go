package utility

import (
	"bufio"
	"encoding/base64"
	"io"
	"reflect"
	"strings"

	"github.com/BytemanD/easygo/pkg/fileutils"
	"github.com/cheggaaa/pb/v3"
)

type ReaderWithProcess struct {
	io.Reader
	Bar *pb.ProgressBar
}

func (reader *ReaderWithProcess) increment(n int) {
	reader.Bar.Add(n)
	if reader.Bar.Current() >= reader.Bar.Total() {
		reader.Bar.Finish()
	}
}

func (reader *ReaderWithProcess) Read(p []byte) (int, error) {
	n, err := reader.Reader.Read(p)
	defer reader.increment(n)
	return n, err
}

func NewProcessReader(reader io.ReadCloser, size int) *ReaderWithProcess {
	return &ReaderWithProcess{
		Reader: bufio.NewReaderSize(reader, 1024*32),
		Bar:    pb.StartNew(size),
	}
}

func GetStructTags(i interface{}) []string {
	tags := []string{}
	iType := reflect.TypeOf(i)
	for i := 0; i < iType.NumField(); i++ {
		tag := iType.Field(i).Tag
		values := strings.Split(tag.Get("json"), ",")
		if len(values) >= 1 {
			tags = append(tags, strings.TrimSpace(values[0]))
		}
	}
	return tags
}

func LoadUserData(file string) (string, error) {
	content, err := fileutils.ReadAll(file)
	if err != nil {
		return "", err
	}
	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
	return encodedContent, nil
}
