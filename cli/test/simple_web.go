package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/test/server_actions"
)

var indexData []byte

func IndexHandler(respWriter http.ResponseWriter, request *http.Request) {
	logging.Info("请求地址 %s", request.URL.Path)
	var err error
	if indexData == nil {
		indexData, err = os.ReadFile("resources/index.html")
	}
	if err != nil {
		respWriter.WriteHeader(http.StatusBadRequest)
		respWriter.Write([]byte("read index failed"))
	} else {
		respWriter.WriteHeader(http.StatusOK)
		respWriter.Write(indexData)
	}
	logging.Info(">>>> %s", request.URL.Path)
}

func ReportHandler(respWriter http.ResponseWriter, request *http.Request) {
	logging.Info("请求地址 %s", request.URL.Path)
	reportBody := struct {
		Report []server_actions.ServerReport `json:"report"`
	}{
		Report: report,
	}
	if len(reportBody.Report) == 0 {
		reportBody.Report = []server_actions.ServerReport{}
	}
	data, err := json.Marshal(&reportBody)
	logging.Debug("report json: %s", string(data))
	respWriter.Header().Set("Content-Type", "application/json")
	if err != nil {
		respWriter.WriteHeader(http.StatusBadRequest)
		respWriter.Write([]byte("get report data failed"))
	} else {
		respWriter.WriteHeader(http.StatusOK)
		respWriter.Write(data)
	}
}

func RunSimpleWebServer() error {
	port := 80
	http.HandleFunc("/", IndexHandler)        //设置访问的路由
	http.HandleFunc("/report", ReportHandler) //设置访问的路由

	logging.Info("启动web服务器, 端口: %d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
