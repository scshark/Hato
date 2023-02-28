package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/scshark/Hato/internal/model"
	"github.com/scshark/Hato/pkg/app"
	"github.com/scshark/Hato/pkg/errcode"
)

func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, exist := c.Get("USER"); exist {
			if userModel, ok := user.(*model.User); ok {
				if userModel.Status == model.UserStatusNormal && userModel.IsAdmin {
					c.Next()
					return
				}
			}
		}

		response := app.NewResponse(c)
		response.ToErrorResponse(errcode.NoAdminPermission)
		c.Abort()
	}
}
