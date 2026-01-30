package middleware

import (
	"encoding/gob"
	"fmt"
	"gin_learning/internal/web"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 不需要校验
		for _, path := range l.paths {
			if ctx.Request.RequestURI == path {
				return
			}
		}

		// JWT校验
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			// 没登录
			println("tokenHeader is empty")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			// jwt 不对，有人瞎搞
			println("tokenHeader format error")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenStr := segs[1]
		uc := &web.UserClaims{}
		// jwt.ParseWithClaims 里面要传uc指针，会往里面写数据
		token, err := jwt.ParseWithClaims(tokenStr, uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
		if err != nil {
			println("token parse error:")
			println(err.Error())
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// err 为 nil, token 不为 nil
		if token == nil || !token.Valid {
			println("token is invalid")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if uc.UserAgent != ctx.Request.UserAgent() {
			// 严重的安全问题
			// 理论上要加监控
			println("user agent not match")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expireTime := uc.ExpiresAt
		remainingTime := expireTime.Sub(time.Now())
		println("剩余有效时间 ", remainingTime)
		if remainingTime < time.Second*30 {
			fmt.Println("刷新 token")
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Second * 60))
			tokenStr, err = token.SignedString(web.JWTKey)
			ctx.Header("x-jwt-token", tokenStr)
			if err != nil {
				log.Println(err)
			}
		}

		ctx.Set("user", uc)
	}
}
