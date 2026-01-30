package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *UserDAO) Edit(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	/*
		UPDATE `users`
		SET
		    `name` = 'u.Name的值',
		    `birthday` = 'u.Birthday的值',
		    `description` = 'u.Description的值',
		    `utime` = 1736752644000 -- 这里的数字是 now 变量的值
		WHERE
		    `id` = u.Id的值;
	*/
	err := dao.db.WithContext(ctx).Debug().Model(&User{}).
		Where("id = ?", u.Id).
		Select("Name", "Birthday", "Description", "Utime").
		Updates(User{
			Name:        u.Name,
			Birthday:    u.Birthday,
			Description: u.Description,
			Utime:       now,
		}).Error
	return err
}

// User 直接对应数据库表结构
// 有些人叫做 entity, 有些人叫做 model, 有些人叫做 PO(persistent object)
type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"`
	Password string

	Name        string
	Birthday    string
	Description string

	// 创建时间, 毫秒数
	Ctime int64
	// 更新时间, 毫秒数
	Utime int64
}
