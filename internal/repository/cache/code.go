package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ErrSendCodeTooMany        = errors.New("验证码发送太频繁")
	ErrVerifyCodeTooManyTimes = errors.New("验证次数太多")
	ErrUnknownForCode         = errors.New("未知错误")
)

// 编译器会在编译的时候，把 set_code 的代码放进来这个 luaSetCode 变量里
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *CodeCache {
	return &CodeCache{
		client: client,
	}
}

func (c *CodeCache) Key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func (c *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.Key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		// 毫无问题
		return nil
	case -1:
		// 发送太频繁
		return ErrSendCodeTooMany
	default:
		// 系统错误
		return errors.New("系统错误")
	}
}

func (c *CodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.Key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		// 验证成功
		return true, nil
	case -1:
		// 验证太多次
		return false, ErrVerifyCodeTooManyTimes
	case -2:
		// 验证失败
		return false, nil
	default:
		// 未知错误
		return false, ErrUnknownForCode
	}
}
