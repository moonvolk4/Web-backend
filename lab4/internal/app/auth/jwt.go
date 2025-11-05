package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      int    `json:"user_id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret      []byte
	ttl         time.Duration
	cookie      string
	blacklister Blacklister
}

type Blacklister interface {
	IsBlacklisted(jti string) (bool, error)
	Blacklist(jti string, expiresAt time.Time) error
}

func NewJWTService(secret string, ttlMinutes int, cookieName string, bl Blacklister) *JWTService {
	return &JWTService{
		secret:      []byte(secret),
		ttl:         time.Duration(ttlMinutes) * time.Minute,
		cookie:      cookieName,
		blacklister: bl,
	}
}

func (s *JWTService) Generate(userID int, login string, isModerator bool) (string, *Claims, error) {
	jti, _ := randomJTI()
	claims := &Claims{
		UserID:      userID,
		Login:       login,
		IsModerator: isModerator,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	return signed, claims, err
}

func (s *JWTService) Parse(tokenString string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tok.Claims.(*Claims); ok && tok.Valid {
		if s.blacklister != nil && claims.ID != "" {
			bl, _ := s.blacklister.IsBlacklisted(claims.ID)
			if bl {
				return nil, errors.New("token blacklisted")
			}
		}
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (s *JWTService) CookieName() string { return s.cookie }

func (s *JWTService) TTL() time.Duration { return s.ttl }

func (s *JWTService) Blacklist(jti string) error {
	if s.blacklister == nil || jti == "" {
		return nil
	}
	return s.blacklister.Blacklist(jti, time.Now().Add(s.ttl))
}

func SetUserContext(c *gin.Context, claims *Claims) {
	c.Set("user_id", claims.UserID)
	c.Set("login", claims.Login)
	c.Set("is_moderator", claims.IsModerator)
	c.Set("jti", claims.ID)
}

func GetUserContext(c *gin.Context) (userID int, login string, isModerator bool, ok bool) {
	uidVal, ok1 := c.Get("user_id")
	loginVal, ok2 := c.Get("login")
	modVal, ok3 := c.Get("is_moderator")
	if !ok1 || !ok2 || !ok3 {
		return 0, "", false, false
	}
	uid, _ := uidVal.(int)
	login, _ = loginVal.(string)
	isModerator, _ = modVal.(bool)
	return uid, login, isModerator, true
}

func randomJTI() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
