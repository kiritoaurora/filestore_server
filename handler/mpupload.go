package handler

import (
	rPool "filestore_server/cache/redis"
	dblayer "filestore_server/db"
	"filestore_server/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

// 初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// 初始化分块上传
func InitiateMultipartUploadHandler(c *gin.Context) {
	// 解析用户请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil {
		c.Data(http.StatusOK, "application/json", util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	// 获得redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 生成分块上传的初始化信息
	upinfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024,                                       // 5M
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))), // 文件大小除5M向上取整
	}

	// 将初始化信息写入到redis缓存
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "chunkcount", upinfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filehash", upinfo.FileHash)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filesize", upinfo.FileSize)

	// 将初始化信息返回给客户端
	c.Data(http.StatusOK, "application/json", util.NewRespMsg(0, "OK", upinfo).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(c *gin.Context) {
	// 解析用户请求参数

	// username := c.Request.FormValue("username")
	uploadID := c.Request.FormValue("uploadid")
	chunkindex := c.Request.FormValue("index")

	// 获取redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 获取文件句柄，用于存储分块内容
	fpath := "/data/" + uploadID + "/" + chunkindex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		c.Data(http.StatusOK, "application/json", util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := c.Request.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 更新redis缓存数据
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkindex, 1)

	// 返回处理结果给客户端
	c.Data(http.StatusOK, "application/json", util.NewRespMsg(0, "OK", nil).JSONBytes())

}

// 通知上传合并
func CompleteUploadHanlder(c *gin.Context) {
	// 解析请求参数
	uploadid := c.Request.FormValue("uploadid")
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil {
		c.Data(http.StatusOK, "application/json", util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	// 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 通过UploadId查询redis并判断所有分块是否上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadid))
	if err != nil {
		c.Data(http.StatusOK, "application/json", util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		key := string(data[i].([]byte))
		value := string(data[i+1].([]byte))
		if key == "chunkcount" {
			totalCount, _ = strconv.Atoi(value)
		} else if strings.HasPrefix(key, "chkidx_") && value == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		c.Data(http.StatusOK, "application/json", util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}

	// TODO：合并分块

	// 更新唯一文件表和用户文件表
	// dblayer.OnFileUploadFinished(filehash, filename, int64(filesize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))

	// 像客户端相应处理结果
	c.Data(http.StatusOK, "application/json", util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// 取消上传
func CancelUploadPartHandler(c *gin.Context) {
	// 删除已存在的分块文件

	// 删除redis缓存，根据username和uploadid查到数据进行删除

	// 更新MySQL文件状态，非必须

}

// 上传状态查询
func MultipartUploadStatusHandler(c *gin.Context) {
	// 检查分块上传状态是否有效

	// 获取分块初始化信息

	// 获取已上传的分块信息

}
