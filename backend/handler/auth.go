package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/irham/topup-backend/config"
	"github.com/irham/topup-backend/middleware"
	"github.com/irham/topup-backend/model"
	"github.com/irham/topup-backend/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users *repository.UserRepo
	jwt   config.JWTConfig
}

func NewAuthHandler(users *repository.UserRepo, jwt config.JWTConfig) *AuthHandler {
	return &AuthHandler{users: users, jwt: jwt}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal hash password"))
		return
	}

	user := model.User{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Username:     req.Username,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		Phone:        req.Phone,
	}
	if err := h.users.Create(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusConflict, model.ErrorResponse("Email atau username sudah dipakai"))
		return
	}

	token, err := h.issueToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal generate token"))
		return
	}
	c.JSON(http.StatusCreated, model.SuccessResponse(model.AuthResponse{Token: token, User: user}))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
		return
	}

	user, err := h.users.FindByEmail(c.Request.Context(), strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse("Email atau password salah"))
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal login"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse("Email atau password salah"))
		return
	}

	_ = h.users.TouchLastLogin(c.Request.Context(), user.ID)

	token, err := h.issueToken(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal generate token"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(model.AuthResponse{Token: token, User: *user}))
}

func (h *AuthHandler) Me(c *gin.Context) {
	uid, _ := c.Get(middleware.CtxUserID)
	idStr, _ := uid.(string)
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse("User tidak valid"))
		return
	}
	u, err := h.users.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse("User tidak ditemukan"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(u))
}

func (h *AuthHandler) issueToken(u model.User) (string, error) {
	claims := middleware.Claims{
		UserID: u.ID.String(),
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   u.ID.String(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(h.jwt.Secret))
}
