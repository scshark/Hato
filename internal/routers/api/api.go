package api

import (
	"github.com/scshark/Hato/internal/core"
	"github.com/scshark/Hato/internal/dao"
)

var (
	objectStorage core.ObjectStorageService
)

func Initialize() {
	objectStorage = dao.ObjectStorageService()
}
