package dao

import (
	"sync"

	"github.com/scshark/Hato/internal/conf"
	"github.com/scshark/Hato/internal/core"
	"github.com/scshark/Hato/internal/dao/jinzhu"
	"github.com/scshark/Hato/internal/dao/sakila"
	"github.com/scshark/Hato/internal/dao/search"
	"github.com/scshark/Hato/internal/dao/slonik"
	"github.com/scshark/Hato/internal/dao/storage"
	"github.com/sirupsen/logrus"
)

var (
	ts                                    core.TweetSearchService
	ds                                    core.DataService
	oss                                   core.ObjectStorageService
	localOss                              core.ObjectStorageService
	onceTs, onceDs, onceOss, onceLocalOss sync.Once
)

func DataService() core.DataService {
	onceDs.Do(func() {
		var v core.VersionInfo
		if conf.CfgIf("Gorm") {
			ds, v = jinzhu.NewDataService()
		} else if conf.CfgIf("Sqlx") && conf.CfgIf("MySQL") {
			ds, v = sakila.NewDataService()
		} else if conf.CfgIf("Sqlx") && (conf.CfgIf("Postgres") || conf.CfgIf("PostgreSQL")) {
			ds, v = slonik.NewDataService()
		} else {
			// default use gorm as orm for sql database
			ds, v = jinzhu.NewDataService()
		}
		logrus.Infof("use %s as data service with version %s", v.Name(), v.Version())
	})
	return ds
}

func ObjectStorageService() core.ObjectStorageService {
	onceOss.Do(func() {
		var v core.VersionInfo
		if conf.CfgIf("AliOSS") {
			oss, v = storage.MustAliossService()
		} else if conf.CfgIf("COS") {
			oss, v = storage.NewCosService()
		} else if conf.CfgIf("HuaweiOBS") {
			oss, v = storage.MustHuaweiobsService()
		} else if conf.CfgIf("MinIO") {
			oss, v = storage.MustMinioService()
		} else if conf.CfgIf("S3") {
			oss, v = storage.MustS3Service()
			logrus.Infof("use S3 as object storage by version %s", v.Version())
			return
		} else if conf.CfgIf("LocalOSS") {
			oss, v = storage.MustLocalossService()
		} else {
			// default use AliOSS as object storage service
			oss, v = storage.MustAliossService()
			logrus.Infof("use default AliOSS as object storage by version %s", v.Version())
			return
		}
		logrus.Infof("use %s as object storage by version %s", v.Name(), v.Version())
	})
	return oss
}

func LocalObjectStorageService() core.ObjectStorageService {
	onceLocalOss.Do(func() {
		var v core.VersionInfo
		localOss, v = storage.MustLocalossService()
		logrus.Infof("use %s as local object storage by version %s", v.Name(), v.Version())
	})
	return localOss
}

func TweetSearchService() core.TweetSearchService {
	onceTs.Do(func() {
		var v core.VersionInfo
		ams := newAuthorizationManageService()
		if conf.CfgIf("Zinc") {
			ts, v = search.NewZincTweetSearchService(ams)
		} else if conf.CfgIf("Meili") {
			ts, v = search.NewMeiliTweetSearchService(ams)
		} else {
			// default use Zinc as tweet search service
			ts, v = search.NewZincTweetSearchService(ams)
		}
		logrus.Infof("use %s as tweet search serice by version %s", v.Name(), v.Version())

		ts = search.NewBridgeTweetSearchService(ts)
	})
	return ts
}

func newAuthorizationManageService() (s core.AuthorizationManageService) {
	if conf.CfgIf("Gorm") {
		s = jinzhu.NewAuthorizationManageService()
	} else if conf.CfgIf("Sqlx") && conf.CfgIf("MySQL") {
		s = sakila.NewAuthorizationManageService()
	} else if conf.CfgIf("Sqlx") && (conf.CfgIf("Postgres") || conf.CfgIf("PostgreSQL")) {
		s = slonik.NewAuthorizationManageService()
	} else {
		s = jinzhu.NewAuthorizationManageService()
	}
	return
}
