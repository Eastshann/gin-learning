package web

import (
	"fmt"
	"gin_learning/internal/domain"
	"gin_learning/internal/service"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")

type UserClaims struct {
	jwt.RegisteredClaims
	UserId    int64
	UserAgent string
}

type UserHandler struct {
	svc                     *service.UserService
	emailRegexPattern       *regexp.Regexp
	nameRegexPattern        *regexp.Regexp
	birthdayRegexPattern    *regexp.Regexp
	descriptionRegexPattern *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) (*UserHandler, error) {
	emailRe, err := regexp.Compile("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")
	nameRe, err := regexp.Compile("^.{1,15}$")
	birthdayRe, err := regexp.Compile("^\\d{4}-\\d{2}-\\d{2}$")
	description, err := regexp.Compile("^.{0,30}$")
	if err != nil {
		return &UserHandler{}, err
	}
	return &UserHandler{
		svc:                     svc,
		emailRegexPattern:       emailRe,
		nameRegexPattern:        nameRe,
		birthdayRegexPattern:    birthdayRe,
		descriptionRegexPattern: description,
	}, nil
}

func (u *UserHandler) RegisterRoutersV1(server *gin.Engine) {
	// 分组路由
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
	ug.POST("/profile", u.Profile)
}

func (u *UserHandler) RegisterRoutersV2(ug *gin.RouterGroup) {
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/loginjwt", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.POST("/profile", u.Profile)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	// Bind 方法会根据 Content-Type 来解析你的数据到 req 里面
	// 解析错了就会之间返回一个 400的错误
	if err := ctx.Bind(&req); err != nil {

		return
	}

	ok := u.emailRegexPattern.MatchString(req.Email)
	if !ok {
		ctx.String(http.StatusBadRequest, "你的邮箱格式不对")
		return
	}

	// 数据库操作
	err := u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusInternalServerError, "邮箱冲突")
		return
	}

	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
	fmt.Printf("%v\n", req)
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusInternalServerError, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	// 设置 session
	session := sessions.Default(ctx)
	session.Set("userId", user.Id)
	session.Save()

	ctx.String(http.StatusOK, "登录成功, %v", user.Id)
	return
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusInternalServerError, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	// 设置 JWT
	uc := UserClaims{
		UserId:    user.Id,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 60)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	ctx.String(http.StatusOK, "登录成功, %v", user.Id)
	return
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Name        string `json:"name"`
		Birthday    string `json:"birthday"`
		Description string `json:"description"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 校验
	ok := u.nameRegexPattern.MatchString(req.Name)
	if !ok {
		ctx.String(http.StatusInternalServerError, "昵称错误")
		return
	}
	ok = u.birthdayRegexPattern.MatchString(req.Birthday)
	if !ok {
		ctx.String(http.StatusInternalServerError, "生日错误")
		return
	}
	ok = u.descriptionRegexPattern.MatchString(req.Description)
	if !ok {
		ctx.String(http.StatusInternalServerError, "描述错误")
		return
	}

	// session 校验
	session := sessions.Default(ctx)
	val := session.Get("userId")

	var userId int64
	if val != nil {
		// 尝试断言为 int（最常见情况）
		if id, ok := val.(int); ok {
			userId = int64(id)
			fmt.Println("获取成功:", userId)
		} else if id64, ok := val.(int64); ok {
			// 预防某些驱动直接返回 int64
			userId = id64
			fmt.Println("获取成功:", userId)
		}
	}

	// 数据库操作
	err := u.svc.Edit(ctx, domain.User{
		Id:          userId,
		Name:        req.Name,
		Birthday:    req.Birthday,
		Description: req.Description,
	})

	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "编辑成功")
	return

}

func (u *UserHandler) Profile(ctx *gin.Context) {
	uc, ok := ctx.Get("user")
	if !ok {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	userChaim, ok := uc.(*UserClaims)
	if !ok {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "profile\nuserID: %v", userChaim.UserId)
	return
}
