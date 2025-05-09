package server_actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
	"github.com/wxnacy/wgo/file"
)

var indexData []byte

func IndexHandler(respWriter http.ResponseWriter, request *http.Request) {
	console.Debug("请求地址 %s", request.URL.Path)
	var err error
	if indexData == nil {
		for _, indexPath := range []string{"static/index.html", "/usr/share/skyman/static/index.html"} {
			if file.IsFile(indexPath) {
				indexData, err = os.ReadFile(indexPath)
				break
			}
		}
	}
	if err != nil {
		respWriter.WriteHeader(http.StatusBadRequest)
		respWriter.Write([]byte("read index failed"))
	} else {
		respWriter.WriteHeader(http.StatusOK)
		respWriter.Write(indexData)
	}
}

func TasksHandler(respWriter http.ResponseWriter, request *http.Request) {
	console.Info("请求地址 %s", request.URL.Path)
	reportBody := struct {
		CaseReports []CaseReport `json:"tasks"`
	}{
		CaseReports: []CaseReport{},
	}
	data, err := json.Marshal(&reportBody)
	console.Debug("tasks json: %s", string(data))
	respWriter.Header().Set("Content-Type", "application/json")
	if err != nil {
		respWriter.WriteHeader(http.StatusBadRequest)
		respWriter.Write([]byte("get report data failed"))
	} else {
		respWriter.WriteHeader(http.StatusOK)
		respWriter.Write(data)
	}
}

func RunSimpleWebServer(webCnof common.Web) error {
	port := webCnof.Port
	//设置访问的路由
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/tasks", TasksHandler)

	webAddr := []string{}
	if ips, err := utility.GetAllIpaddress(); err != nil {
		return err
	} else {
		for _, ip := range ips {
			webAddr = append(webAddr, fmt.Sprintf("http://%s:%d", ip, port))
		}
	}
	console.Info("启动web服务:\n----\n%s\n----", strings.Join(webAddr, "\n"))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
func WaitWebServer() error {
	for {
		time.Sleep(time.Second * 60)
	}
}
