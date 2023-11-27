package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	uuid "github.com/satori/go.uuid"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

func GetIndentJson(v interface{}) (string, error) {
	jsonBytes, _ := json.Marshal(v)
	var buffer bytes.Buffer
	json.Indent(&buffer, jsonBytes, "", "  ")
	return buffer.String(), nil
}

func RaiseIfError(err error, msg string) {
	if err == nil {
		return
	}
	if httpError, ok := err.(*HttpError); ok {
		logging.Fatal("%s, %s, %s", msg, httpError.Reason, httpError.Message)
	} else {
		logging.Fatal("%s, %v", msg, err)
	}
}

func ContainsString(stringList []string, s string) bool {
	for _, str := range stringList {
		if s == str {
			return true
		}
	}
	return false
}

func IsUUID(s string) bool {
	uuid.NewV4()
	if _, err := uuid.FromString(s); err != nil {
		return false
	} else {
		return true
	}
}

func UrlJoin(path ...string) string {
	return strings.Join(path, "/")
}

type ReaderWithProcess struct {
	io.Reader
	Size      int
	completed int
}

func (reader *ReaderWithProcess) PrintProcess(r int) {
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
