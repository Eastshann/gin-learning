package repository

import (
	"context"
	"gin_learning/internal/repository/cache"
)

var (
	ErrSendCodeTooMany        = cache.ErrSendCodeTooMany
	ErrVerifyCodeTooManyTimes = cache.ErrVerifyCodeTooManyTimes
)

type CodeRepository struct {
	cache *cache.CodeCache
}

func NewCodeRepository(c *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: c,
	}
}

func (repo *CodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCode)
}
