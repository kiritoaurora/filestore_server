package handler

import (
	"encoding/json"
	dblayer "filestore_server/db"
	"filestore_server/util"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 响应注册页面
func SignupHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

// 处理用户注册请求
func DoSignupHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	encPasswd := c.Request.FormValue("password")
	phone := c.Request.FormValue("phone")
	pubKey := c.Request.FormValue("pubkey")
	privKey := c.Request.FormValue("privkey")

	// 检验参数有效性，用户名长度不得小于3，密码长度不得小于5
	// if len(username) < 3 || len(passwd) < 5 {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"msg":  "Invalid parameter",
	// 		"code": -1,
	// 	})
	// 	return
	// }
	// 加密用户密码
	// encPasswd := util.Sha1([]byte(passwd + config.PwdSalt))
	// fmt.Printf("%v, %v\n", username, encPasswd)

	//将新用户注册到数据库用户表中
	suc := dblayer.UserSignUp(username, encPasswd, phone, pubKey, privKey)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Signup succeeded",
			"code": 0,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Signup failed",
			"code": -2,
		})
	}
}

// 响应登陆界面
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

// 处理用户登陆
func DoSignInHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	// encPasswd := util.Sha1([]byte(password + config.PwdSalt))

	//1、校验手机及密码，返回用户名
	userKeys, pwdChencked := dblayer.UserSignin(username, password)
	if !pwdChencked {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Password error.",
			"code": -1,
		})
		return
	}

	//2、生成访问凭证token，并存储到数据库tbl_user_token表中
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Login failed",
			"code": -2,
		})
		return
	}

	//3、登陆成功后返回username、token，重定向到首页
	// w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
			PubKey   string
			PrivKey  string
		}{
			Location: "http://" + c.Request.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
			PubKey:   userKeys.PubKey,
			PrivKey:  userKeys.PrivKey,
		},
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}

// 生成token
func GenToken(username string) string {
	// 40位的token： md5(username + timestamp + tokenSalt) + timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

//验证token是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}

	//1、判断token的时效性，是否过期
	ts, _ := strconv.ParseInt(token[32:], 16, 64)
	nowTs := time.Now().Unix()
	diff := (nowTs - ts) / 3600
	return diff <= 24

	//2、从数据库表tbl_user_token查询username对应的token信息

	//3、对比两个token是否一致

	// return true
}

// 查询用户信息
func UserInfoHandler(c *gin.Context) {
	//1、解析请求参数
	username := c.Request.FormValue("username")

	//2、查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Get userInfo failed",
		})
		return
	}

	//3、组装并响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}

// 查询用户好友
func UserFriendsHandler(c *gin.Context) {
	// 解析请求参数
	username := c.Request.FormValue("username")

	// 从好友关系表中查询好友
	friends, err := dblayer.GetUserFriends(username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Get user's friends failed.",
		})
		return
	}

	// 返回用户好友数据
	data, err := json.Marshal(friends)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Json Marshal Error.",
		})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

// 搜索用户
func SearchUserHandler(c *gin.Context) {
	// 解析请求参数
	recipient := c.Request.FormValue("recipient")

	// 从用户表中查找用户信息
	user, err := dblayer.GetUserInfo(recipient)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Get userInfo failed",
		})
		return
	}

	// 用户不存在
	if user.Username == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "User not exist.",
		})
		return
	}

	// 响应查找的用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}

// 添加好友请求
func AddFriendHandler(c *gin.Context) {
	// 解析请求参数
	sender := c.Request.FormValue("username")
	recipient := c.Request.FormValue("recipient")

	// 将请求信息存入数据库
	suc := dblayer.AddFriendReq(sender, recipient)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "消息已发送",
			"code": 0,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "系统繁忙，请稍后重试。",
			"code": http.StatusInternalServerError,
		})
	}
}

// 删除好友
func DeleteFriendHandler(c *gin.Context) {
	// 解析请求参数
	sender := c.Request.FormValue("username")
	recipient := c.Request.FormValue("recipient")

	// 将请求信息存入数据库
	suc := dblayer.AddFriendReq(sender, recipient)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "已删除",
			"code": 0,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "系统繁忙，请稍后重试。",
			"code": http.StatusInternalServerError,
		})
	}
}

// 获取好友申请消息
func FriendsReqMsgHandler(c *gin.Context) {
	// 解析请求数据
	username := c.Request.FormValue("username")

	// 从用户消息表中获取好友申请消息
	reqMsgs, err := dblayer.GetFriendsReqMsg(username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "系统错误，请稍后重试。",
			"code": -1,
		})
		return
	}

	// 返回申请消息
	data, err := json.Marshal(reqMsgs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Json Marshal Error.",
			"code": http.StatusInternalServerError,
		})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

// 处理好友请求
func DoAddFriendHandler(c *gin.Context) {
	// 解析请求参数
	username := c.Request.FormValue("username")
	sender := c.Request.FormValue("sender")
	respType, _ := strconv.Atoi(c.Request.FormValue("resptype"))

	// 判断是否通过好友申请，0同意，1拒绝
	if respType == 0 {
		suc := dblayer.NewFriend(sender, username)
		if suc {
			c.JSON(http.StatusOK, gin.H{
				"msg":  "添加成功",
				"code": 0,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"msg":  "系统错误，请稍后重试。",
				"code": -1,
			})
			return
		}
	} else {

		c.JSON(http.StatusOK, gin.H{
			"msg":  "申请已拒绝",
			"code": 0,
		})
	}

	// 更新用户消息表中记录的状态
	_ = dblayer.UpdateUserMsgStatus(sender, username)
}

// 生成文件共享通知
func ShareMsgHandler(c *gin.Context) {
	// 解析数据
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	checkUsers := c.PostFormArray("checkusers")
	shareKeys := c.PostFormArray("sharekeys")

	// 生成共享信息
	for i, checkUser := range checkUsers {
		suc := dblayer.NewFileShareMsg(filehash, username, checkUser, shareKeys[i])
		if suc {
			continue
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "系统错误",
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "处理完成",
	})
}

// 获取共享通知
func QueryShareMsgHandler(c *gin.Context) {
	// 解析数据
	username := c.Query("username")

	// 从共享消息表中获取文件共享消息
	reqMsgs, err := dblayer.GetShareMsg(username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "系统错误，请稍后重试。",
			"code": -1,
		})
		return
	}

	// 返回文件共享消息
	data, err := json.Marshal(reqMsgs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Json Marshal Error.",
			"code": http.StatusInternalServerError,
		})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

// 加密共享，获取接收者公钥
func GetRecipientPubKey(c *gin.Context) {
	// 解析数据
	checkUsers := c.PostFormArray("checkusers")
	filehash := c.Request.FormValue("filehash")
	username := c.Request.FormValue("username")

	pubKeys := make([]string, len(checkUsers))
	var dataKey string

	// 判定文件是否为加密类型
	isEnc := dblayer.IsEncrypt(filehash)
	encCode := 0 // 回复code, 0未加密/1加密
	if isEnc {
		// 加密文件共享，返回接收者公钥
		for i, checkUser := range checkUsers {
			pubKey := dblayer.GetUserPubKey(checkUser)
			if pubKey == "" {
				c.JSON(http.StatusOK, gin.H{
					"msg":  "Server Error.",
					"code": http.StatusInternalServerError,
				})
				return
			}
			pubKeys[i] = pubKey
		}
		// 获取文件密钥
		row, err := dblayer.GetFileMeta(filehash)
		if err != nil {
			fmt.Println(err.Error())
		}
		dataKey = dblayer.GetShareFileKey(filehash, username)
		if dataKey == "" {
			// 共享表中没有，数据密钥为文件表中的
			dataKey = row.FileKey
		}

		encCode = 1
	}
	// 返回数据
	resp := util.RespMsg{
		Code: encCode,
		Msg:  "OK",
		Data: struct {
			PubKeys []string
			DataKey string
		}{
			PubKeys: pubKeys,
			DataKey: dataKey,
		},
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}
