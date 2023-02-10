package main

import (
	"filestore_server/store/ceph"
	"os"
)

func main() {
	bucket := ceph.GetCephBucket("userfile")

	data, _ := bucket.Get("/ceph/086f58479eff416584dea0672a67e7c915d079f0")
	tmpFile, _ := os.Create("/tmp/test_file")
	tmpFile.Write(data)

	// bucket := ceph.GetCephBucket("testbucket1")

	// // 创建一个新的bucket
	// err := bucket.PutBucket(s3.PublicRead)
	// fmt.Printf("create bucket error: %v\n", err)

	// // 查询这个bucket下指定条件的object keys
	// res, _ := bucket.List("", "", "", 100)
	// fmt.Printf("object keys:%+v\n", res)

	// // 新上传一个对象
	// err = bucket.Put("/testupload/a.txt", []byte("just for test"), "octet-stream", s3.PublicRead)
	// fmt.Printf("upload err:%+v\n", err)

	// // 查询这个bucket下指定条件的object keys
	// res, _ = bucket.List("", "", "", 100)
	// fmt.Printf("object keys:%+v\n", res)

}
