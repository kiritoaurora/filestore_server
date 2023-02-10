package db

import (
	mydb "filestore_server/db/mysql"
	"fmt"
)

// 通过手机号、用户名及密码完成user表的注册操作
func UserSignUp(username, passwd, phone, pubKey, privKey string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user(`user_name`,`user_pwd`,`phone`, `pubKey`, `privKey`) values(?,?,?,?,?)")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, passwd, phone, pubKey, privKey)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	rowsAffected, err := ret.RowsAffected()
	if err == nil && rowsAffected > 0 {
		return true
	}
	return false
}

type UserKey struct {
	PubKey  string
	PrivKey string
}

// 判断密码是否一致
func UserSignin(username string, encPwd string) (UserKey, bool) {
	var userKeys UserKey

	stmt, err := mydb.DBConn().Prepare("select user_pwd, pubKey, privKey from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return userKeys, false
	}
	defer stmt.Close()

	var passwd string
	err = stmt.QueryRow(username).Scan(&passwd, &userKeys.PubKey, &userKeys.PrivKey)
	if err != nil {
		fmt.Println(err.Error())
		return userKeys, false
	}
	fmt.Println(encPwd, passwd)
	if len(passwd) > 0 && passwd == encPwd {
		return userKeys, true
	}

	return userKeys, false
}

// 刷新用户登录的token
func UpdateToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"replace into tbl_user_token (`user_name`, `user_token`) values(?,?)")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// 获取用户token
func GetUserToken(username string) (string, error) {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT user_token FROM tbl_user_token WHERE user_name = ?")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return "", err
	}
	defer stmt.Close()

	var token string
	//执行查询操作
	err = stmt.QueryRow(username).Scan(&token)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return token, nil
}

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
	PubKey       string
	PrivKey      string
}

// 查询用户信息
func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBConn().Prepare(
		"select user_name,signup_at from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return user, err
	}
	defer stmt.Close()

	//执行查询操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	return user, nil
}

// 获取用户好友信息
func GetUserFriends(username string) ([]User, error) {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT friend_name as friends FROM tbl_friends WHERE user_name = ? " +
			"UNION ALL SELECT user_name as friends FROM tbl_friends WHERE friend_name = ? ")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, username)
	if err != nil {
		fmt.Println("查询失败" + err.Error())
		return nil, err
	}

	friends := make([]User, 0)
	for rows.Next() {
		var fName User
		err := rows.Scan(&fName.Username)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		friends = append(friends, fName)
	}
	return friends, nil
}

// 申请好友消息
func AddFriendReq(sender string, recipient string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"INSERT IGNORE INTO tbl_user_message(`sender_name`,`recipient_name`) VALUES(?,?)")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(sender, recipient)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	rowsAffected, err := ret.RowsAffected()
	if err == nil && rowsAffected > 0 {
		return true
	}
	return false
}

// 获取用户好友申请消息
func GetFriendsReqMsg(username string) ([]User, error) {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT sender_name FROM tbl_user_message WHERE recipient_name = ? AND `status` = 0")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	applicants := make([]User, 0)
	for rows.Next() {
		var applicant User
		err := rows.Scan(&applicant.Username)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		applicants = append(applicants, applicant)
	}
	return applicants, nil
}

// 更新申请消息状态为已处理
func UpdateUserMsgStatus(sender string, recipient string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"UPDATE tbl_user_message SET `status` = 1 WHERE	sender_name = ? AND recipient_name = ?")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(sender, recipient)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}

	if rowsAffected, err := ret.RowsAffected(); err == nil && rowsAffected > 0 {
		return true
	}
	return false
}

// 申请好友通过
func NewFriend(sender string, recipient string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"INSERT IGNORE INTO tbl_friends(`user_name`,`friend_name`) VALUES(?,?)")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(sender, recipient)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}

	if rowsAffected, err := ret.RowsAffected(); err == nil {
		if rowsAffected <= 0 {
			fmt.Println("好友关系已存在")
		}
		return true
	}
	return false
}

// 获取用户公钥
func GetUserPubKey(username string) string {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT pubKey FROM tbl_user WHERE user_name = ? LIMIT 1;")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return ""
	}
	defer stmt.Close()

	var pubKey string
	err = stmt.QueryRow(username).Scan(&pubKey)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return pubKey
}
