package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"gin_learning/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExists = redis.Nil

type UserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

// NewUserCache
// A 用到了 B，B 一定是 A 的接口
// A 用到了 B，B 一定是 A 的字段
// A 用到了 B，A 绝对不初始化 B，而是外面注入
func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (cache *UserCache) Key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)

}

func (cache *UserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(&u)
	if err != nil {
		return err
	}
	key := cache.Key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.Key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}
