// Core service implement base sqlx+mysql. All sub-service
// will declare here and provide initial function.

package sakila

import (
	"github.com/scshark/Hato/internal/core"
	"github.com/sirupsen/logrus"
)

func NewDataService() (core.DataService, core.VersionInfo) {
	logrus.Fatal("not support now")
	return nil, nil
}

func NewAuthorizationManageService() core.AuthorizationManageService {
	logrus.Fatal("not support now")
	return nil
}
