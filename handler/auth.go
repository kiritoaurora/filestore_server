package handler

import (
	"filestore_server/common"
	"filestore_server/util"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// http请求拦截器
func HTTPInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		// 验证token是否有效
		if len(username) < 3 || !IsTokenValid(token) {
			c.Abort()
			// token验证无效则返回失败提示
			resp := util.NewRespMsg(
				int(common.StatusTokenInvalid),
				"token无效",
				nil,
			)
			c.JSON(http.StatusOK, resp)
			return
		}
		c.Next()
		log.Printf("username:%v, token:%v, isValid:%v\n", username, token, IsTokenValid(token))
	}
}
