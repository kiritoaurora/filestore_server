package main

import (
	"filestore_server/config"
	"filestore_server/route"
)

func main() {
	// log.Printf("上传服务启动，开始监听[%s]...\n", config.UploadServiceHost)
	// err := http.ListenAndServe(config.UploadServiceHost, nil)
	// if err != nil {
	// 	fmt.Printf("Failed to start server,err:%s", err.Error())
	// }
	router := route.Router()
	router.Run(config.UploadServiceHost)
}
