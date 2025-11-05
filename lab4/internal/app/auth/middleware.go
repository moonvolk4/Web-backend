package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Optional: parse JWT if present in Authorization header or cookie; don't block if absent
func OptionalJWT(jwt *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c, jwt.CookieName())
		if token != "" {
			if claims, err := jwt.Parse(token); err == nil {
				SetUserContext(c, claims)
			}
		}
		c.Next()
	}
}

// RequireJWT: block if no valid token
func RequireJWT(jwt *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c, jwt.CookieName())
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "unauthorized"})
			return
		}
		claims, err := jwt.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "invalid token"})
			return
		}
		SetUserContext(c, claims)
		c.Next()
	}
}

func RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, isMod, ok := GetUserContext(c)
		if !ok || !isMod {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": "forbidden"})
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context, cookieName string) string {
	// Authorization: Bearer <token>
	authz := c.GetHeader("Authorization")
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	if cookie, err := c.Cookie(cookieName); err == nil {
		return strings.TrimSpace(cookie)
	}
	return ""
}
