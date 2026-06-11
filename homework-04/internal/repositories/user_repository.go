package repositories

import (
	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/entity"
	"gorm.io/gorm"
)

type UserDAO struct {
	gorm.Model
	Id        int64  `gorm:"primaryKey"`
	Username  string `gorm:"unique,not null,size:256"`
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
	Email     string `gorm:"not null"`
	Phone     string `gorm:"not null"`
	Updated   int64  `gorm:"autoUpdateTime"` // Use unix milli seconds as updating time
	Created   int64  `gorm:"autoCreateTime"` // Use unix seconds as creating time
}

type UserRepository interface {
	GetAllUsers(c *gin.Context) ([]*entity.User, error)
	GetUserByID(c *gin.Context, userID int64) (*entity.User, error)
	CreateUser(c *gin.Context, user *entity.User) (*entity.User, error)
	UpdateUser(c *gin.Context, user *entity.User) (*entity.User, error)
	DeleteUser(c *gin.Context, userID int64) error
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db}
}

func (u *UserRepositoryImpl) GetAllUsers(c *gin.Context) ([]*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserRepositoryImpl) GetUserByID(c *gin.Context, userID int64) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserRepositoryImpl) CreateUser(c *gin.Context, user *entity.User) (*entity.User, error) {
	result := u.db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (u *UserRepositoryImpl) UpdateUser(c *gin.Context, user *entity.User) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserRepositoryImpl) DeleteUser(c *gin.Context, userID int64) error {
	//TODO implement me
	panic("implement me")
}
