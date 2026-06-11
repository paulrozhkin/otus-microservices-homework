package repositories

import (
	"context"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/entity"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Id        int64  `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null;size:256"`
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
	Email     string `gorm:"not null"`
	Phone     string `gorm:"not null"`
}

type UserRepository interface {
	GetAllUsers(c context.Context) ([]*entity.User, error)
	GetUserByID(c context.Context, userID int64) (*entity.User, error)
	CreateUser(c context.Context, user *entity.User) (*entity.User, error)
	UpdateUser(c context.Context, user *entity.User) (*entity.User, error)
	DeleteUser(c context.Context, userID int64) error
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db}
}

func (u *UserRepositoryImpl) GetAllUsers(c context.Context) ([]*entity.User, error) {
	var usersDao []*User
	var users []*entity.User
	result := u.db.Find(&usersDao)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, userDao := range usersDao {
		users = append(users, &entity.User{
			Id:        userDao.Id,
			Username:  userDao.Username,
			FirstName: userDao.FirstName,
			LastName:  userDao.LastName,
			Email:     userDao.Email,
			Phone:     userDao.Phone,
		})
	}
	return users, nil
}

func (u *UserRepositoryImpl) GetUserByID(c context.Context, userID int64) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserRepositoryImpl) CreateUser(c context.Context, user *entity.User) (*entity.User, error) {
	userDao := &User{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
	}
	result := u.db.WithContext(c).Create(userDao)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (u *UserRepositoryImpl) UpdateUser(c context.Context, user *entity.User) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserRepositoryImpl) DeleteUser(c context.Context, userID int64) error {
	//TODO implement me
	panic("implement me")
}
