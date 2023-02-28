package internal

import (
	"github.com/scshark/Hato/internal/migration"
	"github.com/scshark/Hato/internal/routers/api"
	"github.com/scshark/Hato/internal/service"
)

func Initialize() {
	// migrate database if needed TODO 暂时关闭
	migration.Run()

	// initialize service
	service.Initialize()
	// TODO 暂时关闭
	api.Initialize()
}
