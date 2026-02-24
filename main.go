package main

import (
	"gin_learning/config"
	"gin_learning/internal/repository"
	"gin_learning/internal/repository/cache"
	"gin_learning/internal/repository/dao"
	"gin_learning/internal/service"
	"gin_learning/internal/service/sms/local_sms"
	"gin_learning/internal/web"
	"gin_learning/internal/web/middleware"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	redisSession "github.com/gin-contrib/sessions/redis"

	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initUser(db *gorm.DB, rdb redis.Cmdable, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, uc)
	svc := service.NewUserService(repo)

	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := local_sms.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)

	u, err := web.NewUserHandler(svc, codeSvc)
	if err != nil {
		panic(err)
	}
	u.RegisterRoutersV1(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		// 只在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		log.Printf("数据库连接错误：%v", config.Config.DB.DSN)
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	return db
}

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
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
	store, err := redisSession.NewStore(16, "tcp", "localhost:6378", "", "",
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
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login-jwt").
		IgnorePaths("/users/login-sms/code/send").
		IgnorePaths("/users/login-sms").
		IgnorePaths("/hello").
		CheckLogin())
}

func main() {

	db := initDB()
	rdb := InitRedis()
	server := initWebServer()
	initUser(db, rdb, server)

	server.Run(":8080")
}
