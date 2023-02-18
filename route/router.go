package route

import (
	"filestore_server/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	// gin framework,包括Logger，Recovery
	router := gin.Default()

	// 处理静态资源
	router.LoadHTMLGlob("../../static/view/*")
	router.StaticFS("/static", http.Dir("../../static"))

	// 不需要验证就能访问的接口
	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignupHandler)
	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSignInHandler)
	
	// 用于校验的拦截器
	router.Use(handler.HTTPInterceptor())
	// 这之后的所有handler都会经过拦截器

	// 文件存取接口
	router.GET("/file/upload", handler.UploadHandler)
	router.POST("/file/upload", handler.DoUploadHandler)
	router.POST("/file/meta", handler.GetFileMetaHandler)
	router.GET("/file/download", handler.DownloadHanlder)
	router.POST("/file/update", handler.FileMetaUpdateHandler)
	router.POST("/file/delete", handler.FileDeleteHandler)
	router.POST("/file/query", handler.FileQueryHandler)
	router.GET("/file/downloadceph", handler.DownloadCephHandler)
	router.POST("/file/downloadurl", handler.DownloadURLHandler)
	router.POST("/file/saveshare", handler.SaveShareHandler)
	router.POST("/file/shareignore", handler.IgnoreShareHandler)

	// 秒传接口
	router.POST("/file/fastupload", handler.TryFastUploadHandler)	

	// 分块上传接口
	router.POST("/file/mpupload/init", handler.InitiateMultipartUploadHandler)
	router.POST("/file/mpupload/uppart", handler.UploadPartHandler)
	router.POST("/file/mpupload/complete", handler.CompleteUploadHanlder)


	// 用户相关接口
	router.POST("/user/info", handler.UserInfoHandler)
	router.POST("/user/friends", handler.UserFriendsHandler)
	router.POST("/user/searchuser", handler.SearchUserHandler)
	router.POST("/user/addfriend", handler.AddFriendHandler)
	router.POST("/user/queryfriendreq", handler.FriendsReqMsgHandler)
	router.POST("/user/doaddfriend", handler.DoAddFriendHandler)
	router.POST("/user/deletefriend",handler.DeleteFriendHandler)
	router.POST("/user/sharemsg", handler.ShareMsgHandler)
	router.POST("/user/querysharemsg", handler.QueryShareMsgHandler)
	router.POST("/user/getrecipientpubkey", handler.GetRecipientPubKey)

	return router
}
