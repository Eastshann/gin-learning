package service

import (
	"context"
	"fmt"
	"gin_learning/internal/repository"
	"gin_learning/internal/service/sms"
	"math/rand"
)

var (
	ErrSendCodeTooMany        = repository.ErrSendCodeTooMany
	ErrVerifyCodeTooManyTimes = repository.ErrVerifyCodeTooManyTimes
)

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *CodeService) Send(ctx context.Context,
	biz string, // 用于区别业务场景
	phone string, // 手机号
) error {
	// 生成一个验证码
	code := svc.generateCode()
	// 塞进去 redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	err = svc.smsSvc.Send(ctx, "11111", []string{code}, phone)
	return err
}

func (svc *CodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	// phone_code:$biz:xxxx
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}
