package master

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

//任务的http接口
type ApiServer struct {
	httpServer *http.Server
}

//初始化服务
func NewApiServer() (apiServer *ApiServer, err error) {
	var (
		mux           *http.ServeMux
		lister        net.Listener
		httpServer    *http.Server
		staticDir     http.Dir
		staticHandler http.Handler
	)
	//配置路由
	mux = SetRoute()

	//静态文件目录
	staticDir = http.Dir(ConfigSingle.WebRoot)
	staticHandler = http.FileServer(staticDir)
	mux.Handle("/", staticHandler)

	//启动TCP监听
	if lister, err = net.Listen("tcp", ":"+strconv.Itoa(ConfigSingle.ApiPort)); err != nil {
		return nil, err
	}
	//创建HTTP服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(ConfigSingle.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(ConfigSingle.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}
	go runServer(httpServer, lister)

	return &ApiServer{
		httpServer: httpServer,
	}, nil
}

//主程http运行
func runServer(httpServer *http.Server, lister net.Listener) {
	if err := httpServer.Serve(lister); err != nil {
		log.Println(err)
	}
}
