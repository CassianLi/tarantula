package oss

import (
	"bytes"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"log"
)

// AliOss is the oss configuration params
type AliOss struct {
	Endpoint   string
	AccessID   string
	AccessKey  string
	BucketName string
}

// PutBytesOnOSS illustrates two methods for uploading a file: simple upload and multipart upload.
//  @param objectKey like filename need suffix，exp: oss-image.png
//  @param imgByte []byte
func (aliOss AliOss) PutBytesOnOSS(objectKey string, imgByte []byte) bool {
	client, err := oss.New(aliOss.Endpoint, aliOss.AccessID, aliOss.AccessKey)
	if err != nil {
		log.Printf("oss.bucket failed: %v", err)
		return false
	}

	bucket, err := client.Bucket(aliOss.BucketName)
	if err != nil {
		log.Printf("oss.bucket failed: %v", err)
		return false
	}

	err = bucket.PutObject(objectKey, bytes.NewReader(imgByte))
	if err != nil {
		log.Printf("oss.bytes upload failed: %v", err)
		return false
	}

	return true
}

// PutLocalFileOnOSS
//  @receiver aliOss
//  @param objectKey like filename need suffix，exp: oss-image.png
//  @param filePath is the path of local file
func (aliOss AliOss) PutLocalFileOnOSS(objectKey string, filePath string) {
	client, err := oss.New(aliOss.Endpoint, aliOss.AccessID, aliOss.AccessKey)
	if err != nil {
		log.Fatalf("oss.client new failed: %v", err)
	}

	// 填写存储空间名称，例如examplebucket。
	bucket, err := client.Bucket(aliOss.BucketName)
	if err != nil {
		log.Fatalf("oss.bucket failed: %v", err)
	}

	err = bucket.PutObjectFromFile(objectKey, filePath)
	if err != nil {
		log.Fatalf("oss.file upload failed: %v", err)
	}
}
