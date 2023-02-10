package db

import (
	"database/sql"
	mydb "filestore_server/db/mysql"
	"fmt"
)

// 文件上传完成，保存meta
func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string, saveType int, dataKey string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file(`file_Sha1`,`file_name`,`file_size`," +
			"`file_addr`,`save_type`,`data_key`) values(?,?,?,?,?,?)")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr, saveType, dataKey)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before", filehash)
		}
		return true
	}
	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
	FileKey  string
}

// 从数据库获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,file_addr,data_key from tbl_file " +
			"where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(filehash).Scan(
		&tfile.FileHash, &tfile.FileName, &tfile.FileSize, &tfile.FileAddr, &tfile.FileKey)
	if err != nil {
		if err == sql.ErrNoRows {
			// 查不到对应文件，返回参数和错误均为nil
			return nil, nil
		} else {
			fmt.Println(err.Error())
			return nil, err
		}
	}
	return &tfile, nil
}

// 从数据库批量获取文件元信息
func GetFileMetaList(limit int) ([]TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,file_addr from tbl_file" +
			"where status=1 limit ?")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	columns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(columns))
	var tfiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileName, &tfile.FileSize, &tfile.FileAddr)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}
	fmt.Println(len(tfiles))
	return tfiles, nil
}

// 文件转移后更新文件的存储地址
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set `file_addr`=? where `file_sha1`=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); err == nil {
		if rf <= 0 {
			fmt.Printf("更新文件location失败,filehash:%s\n", filehash)
		}
		return true
	}
	return false
}

// 生成文件共享信息
func NewFileShareMsg(filehash, sender, recipient, shareKey string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"INSERT INTO tbl_share_message(`file_sha1`, `sender_name`, `recipient_name`, `share_key`)" +
			"VALUES(?,?,?,?)")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, sender, recipient, shareKey)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if rf, err := ret.RowsAffected(); err == nil && rf > 0 {
		return true
	}
	return false
}

// 判断是否加密
func IsEncrypt(filehash string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT save_type FROM tbl_file WHERE file_sha1 = ?")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	var isEncrypt int
	err = stmt.QueryRow(filehash).Scan(&isEncrypt)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if isEncrypt == 0 {
		return false
	}
	return true
}

func GetShareFileKey(filehash, recipient string) string {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT share_key FROM `tbl_share_message` WHERE " +
			"file_sha1 = ? AND recipient_name = ? LIMIT 1;")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return ""
	}
	defer stmt.Close()

	var shareKey string
	err = stmt.QueryRow(filehash, recipient).Scan(&shareKey)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return shareKey
}