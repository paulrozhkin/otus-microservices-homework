package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/entity"
	"gorm.io/gorm"
)

const uniqueViolationCode = "23505"

type User struct {
	gorm.Model
	Id           int64  `gorm:"primaryKey"`
	Username     string `gorm:"unique;not null;size:256"`
	PasswordHash string `gorm:"not null;default:''"`
	FirstName    string `gorm:"not null"`
	LastName     string `gorm:"not null"`
	Email        string `gorm:"not null"`
	Phone        string `gorm:"not null"`
	Roles        string `gorm:"not null;default:'user'"`
}

type Session struct {
	ID        string `gorm:"primaryKey;size:36"`
	UserID    int64  `gorm:"not null;index"`
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	GetAllUsers(c context.Context) ([]*entity.User, error)
	GetUserByID(c context.Context, userID int64) (*entity.User, error)
	GetUserByUsername(c context.Context, username string) (*entity.User, error)
	CreateUser(c context.Context, user *entity.User) (*entity.User, error)
	UpdateUser(c context.Context, user *entity.User) (*entity.User, error)
	DeleteUser(c context.Context, userID int64) error
	CreateSession(c context.Context, sessionID string, userID int64, expiresAt time.Time) error
	GetUserBySessionID(c context.Context, sessionID string) (*entity.User, error)
	DeleteSession(c context.Context, sessionID string) error
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
			Roles:     rolesFromString(userDao.Roles),
		})
	}
	return users, nil
}

func (u *UserRepositoryImpl) GetUserByID(c context.Context, userID int64) (*entity.User, error) {
	userDao, err := u.getUserDaoById(c, userID)
	if err != nil {
		return nil, err
	}
	return userFromDao(userDao), nil
}

func (u *UserRepositoryImpl) GetUserByUsername(c context.Context, username string) (*entity.User, error) {
	userDao := &User{}
	result := u.db.WithContext(c).Where("username = ?", username).First(userDao)
	if result.Error != nil {
		if isNotFoundViolation(result.Error) {
			return nil, fmt.Errorf("%w: user with username %q", entity.ErrNotFound, username)
		}

		return nil, result.Error
	}
	return userFromDao(userDao), nil
}

func (u *UserRepositoryImpl) CreateUser(c context.Context, user *entity.User) (*entity.User, error) {
	userDao := &User{
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Phone:        user.Phone,
		Roles:        rolesToString(user.Roles),
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
	if user.PasswordHash != "" {
		userDao.PasswordHash = user.PasswordHash
	}
	userDao.FirstName = user.FirstName
	userDao.LastName = user.LastName
	userDao.Email = user.Email
	userDao.Phone = user.Phone
	if len(user.Roles) > 0 {
		userDao.Roles = rolesToString(user.Roles)
	}
	result := u.db.WithContext(c).Save(userDao)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return nil, fmt.Errorf("%w: user with username %q", entity.ErrAlreadyExists, user.Username)
		}
		return nil, result.Error
	}
	return userFromDao(userDao), nil
}

func (u *UserRepositoryImpl) DeleteUser(c context.Context, userID int64) error {
	result := u.db.WithContext(c).Delete(&User{}, userID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *UserRepositoryImpl) CreateSession(c context.Context, sessionID string, userID int64, expiresAt time.Time) error {
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	result := u.db.WithContext(c).Create(session)
	return result.Error
}

func (u *UserRepositoryImpl) GetUserBySessionID(c context.Context, sessionID string) (*entity.User, error) {
	session := &Session{}
	result := u.db.WithContext(c).First(session, "id = ? AND expires_at > ?", sessionID, time.Now())
	if result.Error != nil {
		if isNotFoundViolation(result.Error) {
			return nil, fmt.Errorf("%w: session %q", entity.ErrUnauthorized, sessionID)
		}

		return nil, result.Error
	}

	return u.GetUserByID(c, session.UserID)
}

func (u *UserRepositoryImpl) DeleteSession(c context.Context, sessionID string) error {
	result := u.db.WithContext(c).Delete(&Session{}, "id = ?", sessionID)
	return result.Error
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

func userFromDao(userDao *User) *entity.User {
	return &entity.User{
		Id:           userDao.Id,
		Username:     userDao.Username,
		PasswordHash: userDao.PasswordHash,
		FirstName:    userDao.FirstName,
		LastName:     userDao.LastName,
		Email:        userDao.Email,
		Phone:        userDao.Phone,
		Roles:        rolesFromString(userDao.Roles),
	}
}

func rolesToString(roles []string) string {
	if len(roles) == 0 {
		return "user"
	}
	return strings.Join(roles, ",")
}

func rolesFromString(roles string) []string {
	if roles == "" {
		return []string{"user"}
	}
	return strings.Split(roles, ",")
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func isNotFoundViolation(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
