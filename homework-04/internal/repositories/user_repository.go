package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/entity"
	"gorm.io/gorm"
)

const uniqueViolationCode = "23505"

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
	result := u.db.WithContext(c).Find(&usersDao)
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
	userDao, err := u.getUserDaoById(c, userID)
	if err != nil {
		return nil, err
	}
	user := &entity.User{
		Id:        userDao.Id,
		Username:  userDao.Username,
		FirstName: userDao.FirstName,
		LastName:  userDao.LastName,
		Email:     userDao.Email,
		Phone:     userDao.Phone,
	}
	return user, nil
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
		if isUniqueViolation(result.Error) {
			return nil, fmt.Errorf("%w: user with username %q", entity.ErrAlreadyExists, user.Username)
		}

		return nil, result.Error
	}
	user.Id = userDao.Id
	return user, nil
}

func (u *UserRepositoryImpl) UpdateUser(c context.Context, user *entity.User) (*entity.User, error) {
	userDao, err := u.getUserDaoById(c, user.Id)
	if err != nil {
		return nil, err
	}
	userDao.Username = user.Username
	userDao.FirstName = userDao.FirstName
	userDao.LastName = userDao.LastName
	userDao.Email = userDao.Email
	userDao.Phone = userDao.Phone
	result := u.db.WithContext(c).Save(userDao)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return nil, fmt.Errorf("%w: user with username %q", entity.ErrAlreadyExists, user.Username)
		}
		return nil, result.Error
	}
	return user, nil
}

func (u *UserRepositoryImpl) DeleteUser(c context.Context, userID int64) error {
	result := u.db.WithContext(c).Delete(&User{}, userID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *UserRepositoryImpl) getUserDaoById(c context.Context, userID int64) (*User, error) {
	userDao := &User{}
	result := u.db.WithContext(c).First(userDao, userID)
	if result.Error != nil {
		if isNotFoundViolation(result.Error) {
			return nil, fmt.Errorf("%w: user with id %d", entity.ErrNotFound, userID)
		}

		return nil, result.Error
	}
	return userDao, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func isNotFoundViolation(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
