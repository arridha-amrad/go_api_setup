package middleware

import (
	"my-go-api/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type VerificationAuthTokenMiddleware struct {
	authService *services.AuthService
}

func RegisterTokenVerificationMiddleware(authService *services.AuthService) *VerificationAuthTokenMiddleware {
	return &VerificationAuthTokenMiddleware{
		authService: authService,
	}
}

func (m VerificationAuthTokenMiddleware) RequireAuth(c *gin.Context) {
	authorization := c.GetHeader("Authorization")
	if authorization == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authorization, bearerPrefix) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		c.Abort()
		return
	}

	tokenStr := strings.TrimSpace(strings.TrimPrefix(authorization, bearerPrefix))
	userId, err := m.authService.ValidateToken(tokenStr, "access")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.Set("authenticatedUserId", userId)
	c.Next()

}
