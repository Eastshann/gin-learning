package repository

import (
	"context"
	"database/sql"
	"gin_learning/internal/domain"
	"gin_learning/internal/repository/cache"
	"gin_learning/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, c *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: c,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(u))
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(u), nil
}

func (repo *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := repo.cache.Get(ctx, id)
	// 必然有数据
	if err == nil {
		return u, nil
	}

	// 没有就去数据库拿
	ue, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = repo.entityToDomain(ue)

	// 存入缓存
	err = repo.cache.Set(ctx, u)
	return u, err
}

func (repo *UserRepository) Update(ctx context.Context, u domain.User) error {
	return repo.dao.Edit(ctx, dao.User{
		Id:          u.Id,
		Name:        u.Name,
		Birthday:    u.Birthday,
		Description: u.Description,
	})
}

func (repo *UserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password:    u.Password,
		Name:        u.Name,
		Birthday:    u.Birthday,
		Description: u.Description,
	}
}

func (repo *UserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:          u.Id,
		Email:       u.Email.String,
		Phone:       u.Phone.String,
		Password:    u.Password,
		Name:        u.Name,
		Birthday:    u.Birthday,
		Description: u.Description,
	}
}
