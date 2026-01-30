package middleware

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 不需要校验
		for _, path := range l.paths {
			if ctx.Request.RequestURI == path {
				return
			}
		}

		// 校验
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 刷新 session
		now := time.Now()
		const updateTimeKey = "updateTime"
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Second*10 {
			sess.Set(updateTimeKey, now)
			sess.Set("userId", userId)
			err := sess.Save()
			fmt.Println("刷新session")
			if err != nil {
				// 打日志
				fmt.Println(err)
			}
		}

	}
}
