package db

import (
	mydb "filestore_server/db/mysql"
	"fmt"
	"time"
)

// 用户文件表结构体
type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// 更新用户文件表
func OnUserFileUploadFinished(username, filehash, filename string, filesize int64) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user_file (`user_name`,`file_sha1`,`file_name`," +
			"`file_size`,`upload_at`) values(?,?,?,?,?)")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash, filename, filesize, time.Now())
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// 批量获取用户文件信息
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,upload_at," +
			"last_update from tbl_user_file where user_name=? and `status`=0 limit ?")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var userFiles []UserFile
	for rows.Next() {
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize,
			&ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles, ufile)
	}
	return userFiles, nil
}

// 用户删除文件
func DeleteUserFile(username string, filehash string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"UPDATE tbl_user_file SET `status` = 1 WHERE user_name = ? AND file_sha1 = ?")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); err == nil && rf > 0 {
			return true
	}
	return false
}

type ShareMsg struct {
	FileHash string
	FileName string
	FileSize int64
	Sender   string
}

// 获取文件共享消息
func GetShareMsg(username string) ([]ShareMsg, error) {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT f.file_sha1, file_name, file_size, sender_name " + 
		"FROM tbl_file f, tbl_share_message s " + 
		"WHERE recipient_name = ? AND s.`status` = 0 AND "+
		"f.file_sha1 = s.file_sha1;")
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

	shareMsgs := make([]ShareMsg, 0)
	for rows.Next() {
		var shareMsg ShareMsg
		err := rows.Scan(&shareMsg.FileHash, &shareMsg.FileName, &shareMsg.FileSize, &shareMsg.Sender)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		shareMsgs = append(shareMsgs, shareMsg)
	}
	return shareMsgs, nil
}

// 更新共享消息状态为已处理
func UpdateShareMsgStatus(sender, recipient, filehash string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"UPDATE tbl_share_message SET `status` = 1 " + 
		"WHERE sender_name = ? AND recipient_name = ? AND file_sha1 = ?")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(sender, recipient, filehash)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}

	if rowsAffected, err := ret.RowsAffected(); err == nil && rowsAffected > 0 {
		return true
	}
	return false
}