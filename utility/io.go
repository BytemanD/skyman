package utility

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strings"

	"github.com/BytemanD/easygo/pkg/fileutils"
	"github.com/BytemanD/go-console/console"
	"github.com/cheggaaa/pb/v3"
	"github.com/samber/lo"
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

type ProgressWriter struct {
	pbr *pb.ProgressBar
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	if pw.pbr != nil {
		pw.pbr.Add(len(p))
	}
	return len(p), nil
}

func NewPbrWriter(total int, writer io.Writer) io.Writer {
	return io.MultiWriter(writer, &ProgressWriter{pbr: pb.StartNew(total)})
}

func NewPbrReader(total int64, reader io.Reader) io.Reader {
	return io.TeeReader(reader, &ProgressWriter{pbr: pb.Start64(total)})
}

type FileProgress struct {
	*os.File
	Total  int64
	readed int64
}

func (f *FileProgress) Read(p []byte) (int, error) {
	n, err := f.File.Read(p)
	f.readed += int64(n)
	console.Info("========> read: %d/%d", f.readed, f.Total)
	return n, err
}

func OpenWithProgress(name string) (*FileProgress, error) {
	fileStat, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return &FileProgress{File: f, Total: fileStat.Size()}, nil
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
