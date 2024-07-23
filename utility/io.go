package utility

import (
	"bufio"
	"encoding/base64"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/fileutils"
	"github.com/BytemanD/skyman/utility/httpclient"
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
	return EncodedUserdata(content), nil
}
func EncodedUserdata(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

func IsFileExists(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil && !os.IsExist(err) {
		return false
	}
	return !fileInfo.Mode().IsDir()
}

func SaveBody(resp *httpclient.Response, file *os.File, process bool) error {
	defer resp.CloseReader()
	var reader io.Reader
	if process && resp.GetContentLength() > 0 {
		reader = NewProcessReader(resp.BodyReader(), resp.GetContentLength())
	} else {
		reader = bufio.NewReaderSize(resp.BodyReader(), 1024*32)
	}

	_, err := io.Copy(bufio.NewWriter(file), reader)
	return err
}

func GetAllIpaddress() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	ips := []string{}
	for _, adddr := range addrs {
		if ipnet, ok := adddr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ips = append(ips, ipnet.IP.String())
		}
	}
	return ips, nil
}

func WatchFileSize(filePath string, size int) {
	bar := pb.StartNew(size)
	currentSize := int64(0)
	for {
		if !IsFileExists(filePath) {
			time.Sleep(time.Second * 2)
			continue
		}
		stat, err := os.Stat(filePath)
		if err != nil {
			break
		}
		bar.Add(int(stat.Size() - currentSize))
		if bar.Current() == bar.Total() {
			break
		}
		currentSize = stat.Size()
		time.Sleep(time.Second * 2)
	}
	bar.Finish()
}
