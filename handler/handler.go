package handler

import (
	"encoding/json"
	"filestore_server/common"
	"filestore_server/config"
	dblayer "filestore_server/db"
	"filestore_server/meta"
	"filestore_server/mq"
	"filestore_server/store/ceph"
	"filestore_server/util"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 响应上传界面
func UploadHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/upload.html")
}

// 处理文件上传
func DoUploadHandler(c *gin.Context) {
	//接收文件流及存储到本地目录
	file, head, err := c.Request.FormFile("file")
	if err != nil {
		fmt.Printf("Failed to get data, err:%s\n", err.Error())
		return
	}
	defer file.Close()

	// 构建文件元信息
	fileMeta := meta.FileMeta{
		FileName: head.Filename,
		Location: "/tmp/" + head.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	// 判定文件是否加密
	dataKey := c.Request.FormValue("datakey")
	if dataKey == "" {
		fileMeta.SaveType = 0
	} else {
		fileMeta.SaveType = 1
	}
	fileMeta.DataKey = dataKey

	// 在系统本地创建对应文件
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		fmt.Println("Failed to create file, err:", err.Error())
		return
	}
	defer newFile.Close()

	// 上传文件
	fileMeta.FileSize, err = io.Copy(newFile, file)
	if err != nil {
		fmt.Println("Failed to save data into file, err:", err.Error())
		return
	}
	// 生成文件SHA1值
	newFile.Seek(0, 0)
	fileMeta.FileSha1 = util.FileSha1(newFile)

	// 游标移回文件头部
	newFile.Seek(0, 0)

	if !config.AsyncTransferEnable {
		// 将文件同步写入ceph存储中
		data, err := ioutil.ReadAll(newFile)
		if err != nil {
			fmt.Printf("Read file err:%v\n", err)
		}
		cephPath := "/ceph/" + fileMeta.FileSha1
		err = ceph.PutObject("userfile", cephPath, data)
		if err != nil {
			fmt.Printf("Ceph upload err:%v\n", err)
		}
		fileMeta.Location = cephPath
	} else {
		// 写入异步转移任务队列
		cephPath := "/ceph/" + fileMeta.FileSha1
		data := mq.TransferData{
			FileHash:      fileMeta.FileSha1,
			CurLocation:   fileMeta.Location,
			DestLocation:  cephPath,
			DestStoreType: common.StoreCeph,
		}
		pubData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("Json Marshal err:%v\n", err)
		}
		pubSuc := mq.Publish(
			config.TransExchangeName,
			config.TransCephRoutingKey,
			pubData,
		)
		if !pubSuc {
			// TODO:当前发送转移信息失败，稍后重试
			fmt.Println("当前发送转移信息失败，稍后重试")
		}
	}

	// 保存文件信息到数据库
	//meta.UpdateFileMeta(fileMeta)
	_ = meta.UpdateFileMetaDB(fileMeta)

	// 更新用户文件表
	username := c.Request.FormValue("username")
	suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Upload Finished.",
			"code": 0,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Upload Failed.",
			"code": -1,
		})
	}
}

// 获取文件元信息
func GetFileMetaHandler(c *gin.Context) {
	// 获取传入的文件哈希值
	filehash := c.Request.FormValue("filehash")
	// fMeta := meta.GetFileMeta(filehash)

	// 根据哈希值获取文件元信息
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "GetFileMeta Error.",
			"code": http.StatusInternalServerError,
		})
		return
	}
	// 对文件元信息进行json序列化并返回
	data, err := json.Marshal(fMeta)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Json Marshal Error.",
			"code": http.StatusInternalServerError,
		})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

// 查询批量的文件元信息
func FileQueryHandler(c *gin.Context) {
	//1、解析请求参数
	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	username := c.Request.FormValue("username")
	// fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)

	//2、获取用户的文件列表
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Query file Error.",
			"code": http.StatusInternalServerError,
		})
		return
	}

	//3、返回文件列表给客户端
	data, err := json.Marshal(userFiles)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Json Marshal Error.",
			"code": http.StatusInternalServerError,
		})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

// 处理下载文件
func DownloadHanlder(c *gin.Context) {
	// 获取文件哈希值
	fsha1 := c.Request.FormValue("filehash")
	fm, _ := meta.GetFileMetaDB(fsha1)
	//打开文件
	f, err := os.Open(fm.Location)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Open file Error.",
			"code": http.StatusInternalServerError,
		})
		return
	}
	defer f.Close()
	// 读取文件数据
	data, err := ioutil.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Read file Error.",
			"code": http.StatusInternalServerError,
		})
		return
	}

	// 发送给客户端
	c.Header("Content-Type", "application/octect-stream")
	c.Header("content-disposition", "attachment;filename=\""+fm.FileName+"\"")
	c.Data(http.StatusOK, "application/octect-stream", data)
}

// 更新元信息（重命名）
func FileMetaUpdateHandler(c *gin.Context) {
	// 获取操作码，文件哈希值和新文件名
	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")
	// 对操作码进行判断
	if opType != "0" {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "opType not zero.",
			"code": http.StatusForbidden,
		})
		return
	}
	// 获取文件元信息并修改，同步到数据库中
	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)
	// 对文件元信息进行json序列发并返回
	data, err := json.Marshal(curFileMeta)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "JSON Marshal Error.",
			"code": http.StatusInternalServerError,
		})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

// 删除文件及元信息
func FileDeleteHandler(c *gin.Context) {
	// 解析请求参数
	fileSha1 := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")

	// 将用户文件表中对应记录的status置为“已删除”
	suc := dblayer.DeleteUserFile(username, fileSha1)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Remove file Finished.",
			"code": 0,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "系统错误",
		"code": http.StatusInternalServerError,
	})
}

// 尝试秒传接口
func TryFastUploadHandler(c *gin.Context) {
	//1、解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, _ := strconv.Atoi(c.Request.FormValue("filesize"))

	//2、从文件表中查询相同hash的文件
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Get fileMeta error.",
			"code": http.StatusInternalServerError,
		})
		return
	}

	//3、查不到文件则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}

	//4、查到则将文件信息写入用户文件表，返回成功
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}
}

// 处理Ceph文件下载
func DownloadCephHandler(c *gin.Context) {
	// 从数据库中获取文件元信息
	fsha1 := c.Request.FormValue("filehash")
	fm, _ := meta.GetFileMetaDB(fsha1)
	// 获取ceph路径
	cephPath := fm.Location
	data, err := ceph.GetObject("userfile", cephPath)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Server Error.",
			"code": -1,
		})
		return
	}
	// 发送给客户端
	c.Header("Content-Type", "application/octet-stream")
	c.Header("content-disposition", "attachment;filename=\""+fm.FileName+"\"")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// 生成文件下载URL
func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")
	token := c.Request.FormValue("token")
	sender := c.Request.FormValue("sender")

	// 从数据库文件表中查找记录
	row, err := dblayer.GetFileMeta(filehash)
	if err != nil {
		fmt.Println(err.Error())
	}

	// 从共享文件表中查找记录，获取数据密钥
	dataKey := dblayer.GetShareFileKey(filehash, username)
	if dataKey == "" {
		// 共享表中没有，数据密钥为文件表中的
		dataKey = row.FileKey
	}

	// 判断是否是从文件共享处下载
	if sender != "" {
		username = sender
	}

	// 判断文件存储在本地还是Ceph集群中
	var Url string
	if strings.HasPrefix(row.FileAddr.String, "/tmp") {
		Url = fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		// c.String(http.StatusOK, tmpURL)
	} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
		// ceph下载url
		Url = fmt.Sprintf("http://%s/file/downloadceph?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		// c.String(http.StatusOK, cephURL)
	}
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			URL      string
			FileName string
			DataKey  string
		}{
			URL:      Url,
			FileName: row.FileName.String,
			DataKey:  dataKey,
		},
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}

// 保存分享文件
func SaveShareHandler(c *gin.Context) {
	// 解析数据
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, _ := strconv.Atoi(c.Request.FormValue("filesize"))
	username := c.Request.FormValue("username")
	sender := c.Request.FormValue("sender")

	// 将数据保存到用户文件表中
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Save success.",
			"code": 0,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Save Failed.",
			"code": -1,
		})
	}

	// 更新共享信息表
	_ = dblayer.UpdateShareMsgStatus(sender, username, filehash)
}

// 忽略分享
func IgnoreShareHandler(c *gin.Context) {
	// 解析数据
	filehash := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")
	sender := c.Request.FormValue("sender")

	// 更新共享信息表
	suc := dblayer.UpdateShareMsgStatus(sender, username, filehash)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "操作成功",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误",
		})
	}
}
