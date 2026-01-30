package main

import (
	"gin_learning/config"
	"gin_learning/internal/repository"
	"gin_learning/internal/repository/dao"
	"gin_learning/internal/service"
	"gin_learning/internal/web"
	"gin_learning/internal/web/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"

	//"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initUser(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u, err := web.NewUserHandler(svc)
	if err != nil {
		panic(err)
	}

	u.RegisterRoutersV2(server.Group("/users"))
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		// 只在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	// CORS中间件
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://127.0.0.1:3000",
			"http://localhost:3000",
		},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 基于 redis 的 session 校验中间件
	//useSession(server)

	// JWT 校验中间件
	useJWT(server)

	return server
}

func useSession(server *gin.Engine) {
	store, err := redis.NewStore(16, "tcp", "localhost:6378", "", "",
		[]byte("U5rRMiqk12W0KGEh1YVry64U7ruRDLAm"),
		[]byte("5HivTKFUdMHsrcviM5aLrvNECAu42WNb"))

	if err != nil {
		panic(err)
	}
	store.Options(sessions.Options{
		Path:   "/",
		MaxAge: 30, // 30s 过期
	})

	server.Use(sessions.Sessions("mysession", store))
	server.Use(
		middleware.NewLoginMiddlewareBuilder().
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			CheckLogin(),
	)

}

func useJWT(server *gin.Engine) {
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/loginjwt").
		IgnorePaths("/users/signup").
		IgnorePaths("/hello").
		CheckLogin())
}

func main() {

	db := initDB()
	server := initWebServer()
	initUser(db, server)

	server.Run(":8080")
}
