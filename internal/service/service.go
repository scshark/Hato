package service

import (
	"github.com/scshark/Hato/internal/conf"
	"github.com/scshark/Hato/internal/core"
	"github.com/scshark/Hato/internal/dao"
	"github.com/scshark/Hato/internal/model"
	"github.com/sirupsen/logrus"
)

var (
	ds                 core.DataService
	ts                 core.TweetSearchService
	oss                core.ObjectStorageService
	DisablePhoneVerify bool
)

func Initialize() {

	// TODO 暂时关闭
	ds = dao.DataService()
	ts = dao.TweetSearchService()
	oss = dao.ObjectStorageService()

	//localOss = dao.LocalObjectStorageService()

	DisablePhoneVerify = !conf.CfgIf("Sms")

	// TODO init crawler sync data to hato
	InitCrawler()
}

// persistMediaContents 获取媒体内容并持久化
func persistMediaContents(contents []*PostContentItem) (items []string, err error) {
	items = make([]string, 0, len(contents))
	for _, item := range contents {
		switch item.Type {
		case model.CONTENT_TYPE_IMAGE,
			model.CONTENT_TYPE_VIDEO,
			model.CONTENT_TYPE_AUDIO,
			model.CONTENT_TYPE_ATTACHMENT,
			model.CONTENT_TYPE_CHARGE_ATTACHMENT:
			items = append(items, item.Content)
			if err != nil {
				continue
			}
			if err = oss.PersistObject(oss.ObjectKey(item.Content)); err != nil {
				logrus.Errorf("service.persistMediaContents failed: %s", err)
			}
		}
	}
	return
}
