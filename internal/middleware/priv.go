package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/scshark/Hato/internal/conf"
	"github.com/scshark/Hato/internal/model"
	"github.com/scshark/Hato/pkg/app"
	"github.com/scshark/Hato/pkg/errcode"
)

func Priv() gin.HandlerFunc {
	if conf.CfgIf("PhoneBind") {
		return func(c *gin.Context) {
			if u, exist := c.Get("USER"); exist {
				if user, ok := u.(*model.User); ok {
					if user.Status == model.UserStatusNormal {
						if user.Phone == "" {
							response := app.NewResponse(c)
							response.ToErrorResponse(errcode.AccountNoPhoneBind)
							c.Abort()
							return
						}
						c.Next()
						return
					}
				}
			}
			response := app.NewResponse(c)
			response.ToErrorResponse(errcode.UserHasBeenBanned)
			c.Abort()
		}
	} else {
		return func(c *gin.Context) {
			if u, exist := c.Get("USER"); exist {
				if user, ok := u.(*model.User); ok && user.Status == model.UserStatusNormal {
					c.Next()
					return
				}
			}
			response := app.NewResponse(c)
			response.ToErrorResponse(errcode.UserHasBeenBanned)
			c.Abort()
		}
	}
}
