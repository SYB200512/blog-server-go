package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"blog-backend/global"
)

// AccessClaims access token 的声明：携带展示所需的最小身份信息，避免每次请求查库。
type AccessClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     int8   `json:"role"`
	jwt.RegisteredClaims
}

// RefreshClaims refresh token 的声明：只用于换取新的 access token。
// SessionID 用于定位服务端会话，支持多设备同时在线。
type RefreshClaims struct {
	UserID    uint64 `json:"user_id"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

// JWT 双 token 签发与校验器。access 与 refresh 使用不同密钥，
// 防止 refresh token 被当作 access token 使用。
type JWT struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewJWT() *JWT {
	return &JWT{
		accessSecret:  []byte(global.Config.JWT.AccessTokenSecret),
		refreshSecret: []byte(global.Config.JWT.RefreshTokenSecret),
	}
}

// CreateAccessToken 生成短期 access token，并返回其过期时间。
func (j *JWT) CreateAccessToken(userID uint64, username string, role int8) (string, time.Time, error) {
	ttl, err := time.ParseDuration(global.Config.JWT.AccessTokenExpiryTime)
	if err != nil {
		return "", time.Time{}, err
	}
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := AccessClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    global.Config.JWT.Issuer,
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.accessSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// CreateRefreshToken 生成长期 refresh token，并返回其过期时间。
func (j *JWT) CreateRefreshToken(userID uint64, sessionID string) (string, time.Time, error) {
	ttl, err := time.ParseDuration(global.Config.JWT.RefreshTokenExpiryTime)
	if err != nil {
		return "", time.Time{}, err
	}
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := RefreshClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    global.Config.JWT.Issuer,
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.refreshSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// ParseAccessToken 解析 access token。过期时返回的错误可用 IsTokenExpired 判断。
func (j *JWT) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	if err := j.parse(tokenString, claims, j.accessSecret); err != nil {
		return nil, err
	}
	return claims, nil
}

// ParseRefreshToken 解析 refresh token。
func (j *JWT) ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	if err := j.parse(tokenString, claims, j.refreshSecret); err != nil {
		return nil, err
	}
	return claims, nil
}

func (j *JWT) parse(tokenString string, claims jwt.Claims, secret []byte) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非法的签名算法")
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("无效的 token")
	}
	return nil
}

// IsTokenExpired 判断解析错误是否为“已过期”，用于区分可续签与非法 token。
func IsTokenExpired(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired)
}

// NewSessionID 生成随机的会话标识，用于支持多设备登录与单设备注销。
func NewSessionID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
