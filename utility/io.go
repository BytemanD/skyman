package utility

import (
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/BytemanD/easygo/pkg/fileutils"
)

type ReaderWithProcess struct {
	io.Reader
	Size      int
	completed int
}

func (reader *ReaderWithProcess) PrintProcess(r int) {
	if r == 0 {
		return
	}
	reader.completed = reader.completed + r
	percent := reader.completed * 100 / reader.Size
	fmt.Printf("\r[%-50s]", strings.Repeat("=", percent/2)+">")
	if percent >= 100 {
		fmt.Println()
	}
}

func (reader *ReaderWithProcess) Read(p []byte) (n int, err error) {
	n, err = reader.Reader.Read(p)
	reader.PrintProcess(n)
	return n, err
}

type WriterWithProces struct {
	io.Writer
	Size      int
	completed int
}

func (reader *WriterWithProces) PrintProcess(r int) {
	reader.completed = reader.completed + r
	percent := reader.completed * 100 / reader.Size
	fmt.Printf("\r[%-50s]", strings.Repeat("=", percent/2)+">")
	if r == 0 {
		fmt.Println()
	}
}

func (reader *WriterWithProces) Write(p []byte) (n int, err error) {
	n, err = reader.Writer.Write(p)
	reader.PrintProcess(n)
	return n, err
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
