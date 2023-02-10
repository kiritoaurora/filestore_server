package ceph

import (
	"filestore_server/config"
	"time"

	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3

// 获取ceph连接
func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}
	//初始化ceph的一些信息
	auth := aws.Auth{
		AccessKey: config.CephAccessKey,
		SecretKey: config.CephSecretKey,
	}

	regoin := aws.Region{
		Name:                 "default",
		EC2Endpoint:          config.CephGWEndpoint,
		S3Endpoint:           config.CephGWEndpoint,
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}

	//创建S3类型的连接
	return s3.New(auth, regoin)

}

// 获取指定的bucket对象
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConnection()
	return conn.Bucket(bucket)
}

// 上传文件到ceph集群
func PutObject(bucket string, path string, data []byte) error {
	return GetCephBucket(bucket).Put(path, data, "octet-stream", s3.PublicRead)
}

// 从ceph集群中下载文件
func GetObject(bucket string, path string) ([]byte, error) {
	data, err := GetCephBucket(bucket).Get(path)
	return data, err
}

// 获取临时下载URL
func DownloadURL(bucket string, cephPath string) string {
	signedURL := GetCephBucket(bucket).SignedURL(cephPath, time.Now().Add(time.Hour))
	return signedURL
}
