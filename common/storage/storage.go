package storage

import (
	"io"
	"msg/common/log"
	setting "msg/conf"
	"msg/constants"
	"os"
)

var FileStorage FileStorageInterface

type FileStorageInterface interface {
	SignURL(objectKey string, method constants.HTTPMethod, expiredInSec int64) (signedURL string, err error)
	Get(objectKey string) (content io.ReadCloser, err error)
	Put(objectKey string, reader io.Reader) (err error)
	IsExist(objectKey string) (ok bool, err error)
	PutFromFile(objectKey string, filePath string) (err error)
	Delete(objectKeys ...string) (deletedObjects []string, err error)
}

func Setup(conf setting.StorageConfig) {
	var err error
	if conf.Type == string(constants.AliyunStorage) {
		FileStorage, err = NewOSS(conf)
		if err != nil {
			log.TracedError("NewOSS failed", err)
			os.Exit(1)
			return
		}
	}

	if conf.Type == string(constants.QcloudStorage) {
		FileStorage, err = NewCOS(conf)
		if err != nil {
			log.TracedError("NewCOS failed", err)
			os.Exit(1)
			return
		}
	}

	//if conf.Type == string(constants.LocalStorage) {
	//	FileStorage, err = NewLocalStorage(conf)
	//	if err != nil {
	//		log.TracedError("NewLocalStorage failed", err)
	//		os.Exit(1)
	//		return
	//	}
	//}

	return
}
