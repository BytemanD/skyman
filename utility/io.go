package utility

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"

	"github.com/BytemanD/easygo/pkg/fileutils"
	"github.com/cheggaaa/pb/v3"
	"github.com/samber/lo"
)

func GetStructTags(i any) []string {
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

type ProcessReader struct {
	ReadCloser io.ReadCloser
	bar        *pb.ProgressBar
}

func (reader *ProcessReader) Read(p []byte) (int, error) {
	n, err := reader.ReadCloser.Read(p)
	reader.bar.Add(n)
	return n, err
}
func (reader *ProcessReader) Close() error {
	reader.bar.Finish()
	return reader.ReadCloser.Close()
}

func NewProcessReader(reader io.ReadCloser, size int64) *ProcessReader {
	return &ProcessReader{reader, pb.Start64(size)}
}

type ProgressWriter struct {
	pbr *pb.ProgressBar
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	if pw.pbr != nil {
		pw.pbr.Add(len(p))
	}
	return len(p), nil
}

func NewProgressWriter(writer io.Writer, total int) io.Writer {
	return io.MultiWriter(writer, &ProgressWriter{pbr: pb.StartNew(total)})
}

func DefaultScanComfirm(message string) bool {
	return ScanfComfirm(message, []string{"yes", "y"}, []string{"no", "n"})
}
func ScanfComfirm(message string, yes, no []string) bool {
	var confirm string
	for {
		fmt.Printf("%s [%s|%s]: ", message, strings.Join(yes, "/"), strings.Join(no, "/"))
		fmt.Scanf("%s %d %f", &confirm)
		if lo.Contains(yes, confirm) {
			return true
		} else if lo.Contains(no, confirm) {
			return false
		} else {
			fmt.Print("输入错误, 请重新输入!")
		}
	}
}
