package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zoshc/secunda-task-manager/internal/config"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func Generate(userID int64, issuer, key string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(key))
}

func Validate(tokenStr, key string, leeway time.Duration) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(key), nil
		},
		jwt.WithLeeway(leeway),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

type Provider struct {
	issuer     string
	leeway     time.Duration
	accessKey  string
	accessTTL  time.Duration
	refreshKey string
	refreshTTL time.Duration
}

func NewProvider(cfg config.JWTConfig) *Provider {
	return &Provider{
		issuer:     cfg.Issuer,
		leeway:     cfg.Leeway,
		accessKey:  cfg.Access.Key,
		accessTTL:  cfg.Access.TTL,
		refreshKey: cfg.Refresh.Key,
		refreshTTL: cfg.Refresh.TTL,
	}
}

func (p *Provider) GenerateAccess(userID int64) (string, error) {
	return Generate(userID, p.issuer, p.accessKey, p.accessTTL)
}

func (p *Provider) GenerateRefresh(userID int64) (string, error) {
	return Generate(userID, p.issuer, p.refreshKey, p.refreshTTL)
}

func (p *Provider) ValidateAccess(tokenStr string) (*Claims, error) {
	claims, err := Validate(tokenStr, p.accessKey, p.leeway)
	if err != nil {
		return nil, errors.New("invalid access token")
	}
	return claims, nil
}

func (p *Provider) ValidateRefresh(tokenStr string) (int64, error) {
	claims, err := Validate(tokenStr, p.refreshKey, p.leeway)
	if err != nil {
		return 0, errors.New("invalid refresh token")
	}
	return claims.UserID, nil
}
