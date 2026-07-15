package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/repositories"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "session_id"

type AuthHandler struct {
	userRepository repositories.UserRepository
	logger         *zap.Logger
}

type RegisterRequest struct {
	Login     string   `json:"login" binding:"required,max=256"`
	Password  string   `json:"password" binding:"required"`
	Email     string   `json:"email" binding:"required,email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Phone     string   `json:"phone"`
	Roles     []string `json:"roles"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(logger *zap.Logger, userRepository repositories.UserRepository) *AuthHandler {
	return &AuthHandler{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	request := &RegisterRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Error(err)
		return
	}

	user := &entity.User{
		Username:     request.Login,
		PasswordHash: string(passwordHash),
		FirstName:    request.FirstName,
		LastName:     request.LastName,
		Email:        request.Email,
		Phone:        request.Phone,
		Roles:        request.Roles,
	}
	if len(user.Roles) == 0 {
		user.Roles = []string{"user"}
	}

	createdUser, err := h.userRepository.CreateUser(c, user)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

func (h *AuthHandler) Login(c *gin.Context) {
	request := &LoginRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	user, err := h.userRepository.GetUserByUsername(c, request.Login)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.Error(entity.ErrUnauthorized)
			return
		}
		c.Error(err)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		c.Error(entity.ErrUnauthorized)
		return
	}

	sessionID := uuid.NewString()
	sessionTTL := 24 * time.Hour
	if err = h.userRepository.CreateSession(c, sessionID, user.Id, time.Now().Add(sessionTTL)); err != nil {
		c.Error(err)
		return
	}

	c.SetCookie(sessionCookieName, sessionID, int(sessionTTL.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AuthHandler) Auth(c *gin.Context) {
	sessionID, err := c.Cookie(sessionCookieName)
	if err != nil {
		c.Error(entity.ErrUnauthorized)
		return
	}

	user, err := h.userRepository.GetUserBySessionID(c, sessionID)
	if err != nil {
		c.Error(err)
		return
	}

	c.Header("X-Auth-UserId", strconv.FormatInt(user.Id, 10))
	c.Header("X-Auth-User", user.Username)
	c.Header("X-Auth-Email", user.Email)
	c.Header("X-Auth-Roles", strings.Join(user.Roles, ","))
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
