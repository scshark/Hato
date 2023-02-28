//go:build docs
// +build docs

package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/scshark/Hato/docs"
)

// registerDocs register docs asset route
func registerDocs(e *gin.Engine) {
	e.StaticFS("/docs", docs.NewFileSystem())
}
