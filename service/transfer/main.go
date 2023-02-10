package main

import (
	"encoding/json"
	"filestore_server/config"
	dblayer "filestore_server/db"
	"filestore_server/mq"
	"filestore_server/store/ceph"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// 处理文件转移
func ProcessTransfer(msg []byte) bool {
	// 解析msg
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 根据临时存储文件路径，创建文件句柄
	file, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer file.Close()

	// 通过文件句柄将文件上传到ceph
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("Read file err:%v\n", err)
	}
	cephPath := "/ceph/" + pubData.FileHash
	err = ceph.PutObject("userfile", cephPath, data)
	if err != nil {
		fmt.Printf("Ceph upload err:%v\n", err)
	}

	// 更新文件的存储路径到文件表
	_ = dblayer.UpdateFileLocation(pubData.FileHash, pubData.DestLocation)

	return true
}

func main() {
	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能被禁用，请检查相关配置！")
		return
	}
	log.Println("文件转移服务启动，开始监听转移任务队列...")
	mq.StartConsume(
		config.TransCephQueueName,
		"transfer_ceph",
		ProcessTransfer)
}
