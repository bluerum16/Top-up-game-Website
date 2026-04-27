package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/irham/topup-backend/config"
	"github.com/irham/topup-backend/model"
)

const (
	CtxUserID = "user_id"
	CtxRole   = "role"
)

type Claims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func RequireAuth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("Token tidak ditemukan"))
			return
		}

		claims := &Claims{}
		_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(cfg.Secret), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("Token tidak valid"))
			return
		}

		if _, err := uuid.Parse(claims.UserID); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse("Subject token tidak valid"))
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role, _ := c.Get(CtxRole)
		s, _ := role.(string)
		if _, ok := allowed[s]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, model.ErrorResponse("Akses ditolak"))
			return
		}
		c.Next()
	}
}

func extractBearer(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}
