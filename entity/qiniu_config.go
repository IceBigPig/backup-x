package entity

import (
	"backup-x/util"
	"context"
	"errors"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"log"
)

// QiniuConfig QiniuConfig
type QiniuConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}

var ErrQiniuEmpty = errors.New("qiniu config is empty")

func (qiniuConfig QiniuConfig) CheckNotEmpty() bool {
	return qiniuConfig.AccessKey != "" &&
		qiniuConfig.SecretKey != "" &&
		qiniuConfig.Bucket != "" &&
		qiniuConfig.Domain != ""
}

func (qiniuConfig QiniuConfig) getMac() (*qbox.Mac, error) {

	if !qiniuConfig.CheckNotEmpty() {
		return nil, ErrQiniuEmpty
	}

	conf, err := GetConfigCache()
	if err != nil {
		return nil, err
	}

	secretKey, err := util.DecryptByEncryptKey(conf.EncryptKey, qiniuConfig.SecretKey)
	if err != nil {
		return nil, err
	}

	mac := qbox.NewMac(qiniuConfig.AccessKey, secretKey)
	return mac, nil
}

// UploadFile 上传
func (qiniuConfig QiniuConfig) UploadFile(localFile string, key string) error {

	mac, err := qiniuConfig.getMac()
	if err != nil {
		return err
	}

	putPolicy := storage.PutPolicy{
		Scope: qiniuConfig.Bucket,
	}
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{}
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}

	putExtra := storage.PutExtra{
		Params: map[string]string{
			"x:name": "github logo",
		},
	}

	err = formUploader.PutFile(context.Background(), &ret, upToken, key, localFile, &putExtra)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// ListFiles 列出文件
func (qiniuConfig QiniuConfig) ListFiles(prefix string) ([]string, error) {

	mac, err := qiniuConfig.getMac()
	if err != nil {
		return nil, err
	}

	cfg := storage.Config{}
	bucketManager := storage.NewBucketManager(mac, &cfg)

	var listFiles []string

	entries, _, _, _, err := bucketManager.ListFiles(qiniuConfig.Bucket, prefix, "", "", 1000)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		listFiles = append(listFiles, entry.Key)
	}

	return listFiles, nil
}

// DeleteFile 删除文件
func (qiniuConfig QiniuConfig) DeleteFile(key string) error {

	mac, err := qiniuConfig.getMac()
	if err != nil {
		return err
	}

	cfg := storage.Config{}
	bucketManager := storage.NewBucketManager(mac, &cfg)
	delErr := bucketManager.Delete(qiniuConfig.Bucket, key)

	if delErr != nil {
		log.Println(delErr)
		return delErr
	}

	return nil
}
