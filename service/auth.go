package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/errs"
	"blog-backend/model/request"
	"blog-backend/model/response"
	"blog-backend/utils"
	"blog-backend/utils/cache"
)

// AuthService 认证业务
type AuthService struct{}

// LoginResult 登录结果：access token 返回给前端，refresh token 由 API 层写入 Cookie。
type LoginResult struct {
	AccessToken  string
	ExpiresAt    time.Time
	RefreshToken string
	RefreshTTL   time.Duration
	User         *response.UserInfo
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *request.LoginRequest) (*LoginResult, error) {
	user, err := dao.GetUserByUsername(ctx, req.Username)
	if err != nil {
		global.Log.Warn("登录失败：用户查询错误", zap.String("username", req.Username), zap.Error(err))
		if errors.Is(err, errs.ErrNotFound) {
			// 用户名不存在与密码错误返回同一文案，避免泄露用户是否存在
			return nil, errs.New(errs.ErrInvalid, "用户名或密码错误")
		}
		return nil, err
	}
	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		global.Log.Warn("登录失败：密码错误", zap.String("username", req.Username))
		return nil, errs.New(errs.ErrInvalid, "用户名或密码错误")
	}

	j := utils.NewJWT()
	accessToken, expiresAt, err := j.CreateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		global.Log.Error("生成 access token 失败", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}
	// 每个设备一个 sessionID：多设备互不影响，退出时只注销当前设备
	sessionID, err := utils.NewSessionID()
	if err != nil {
		global.Log.Error("生成会话标识失败", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}
	refreshToken, refreshExpiresAt, err := j.CreateRefreshToken(user.ID, sessionID)
	if err != nil {
		global.Log.Error("生成 refresh token 失败", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}
	refreshTTL := time.Until(refreshExpiresAt)
	if err := cache.SaveRefreshSession(ctx, user.ID, sessionID, refreshToken, refreshTTL); err != nil {
		global.Log.Error("保存登录会话失败", zap.Uint64("user_id", user.ID), zap.Error(err))
		return nil, errs.Wrap(errs.ErrStorage, "登录状态保存失败，请稍后重试", err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		ExpiresAt:    expiresAt,
		RefreshToken: refreshToken,
		RefreshTTL:   refreshTTL,
		User:         toUserInfo(user),
	}, nil
}

// ValidateSession 校验 refresh token 对应的服务端会话是否仍然有效。
// 会话被注销、过期或与 Cookie 中的 token 不一致时返回 false。
func (s *AuthService) ValidateSession(ctx context.Context, userID uint64, sessionID, refreshToken string) bool {
	stored, found, err := cache.GetRefreshSession(ctx, userID, sessionID)
	if err != nil {
		global.Log.Warn("校验登录会话失败", zap.Uint64("user_id", userID), zap.Error(err))
		return false
	}
	if !found {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(stored), []byte(refreshToken)) == 1
}

// Logout 注销当前设备的会话；refresh token 非法时静默忽略。
func (s *AuthService) Logout(ctx context.Context, refreshToken string) {
	if refreshToken == "" {
		return
	}
	claims, err := utils.NewJWT().ParseRefreshToken(refreshToken)
	if err != nil {
		return
	}
	if err := cache.DeleteRefreshSession(ctx, claims.UserID, claims.SessionID); err != nil {
		global.Log.Warn("注销会话失败", zap.Uint64("user_id", claims.UserID), zap.Error(err))
	}
}

// GetProfile 获取用户信息
func (s *AuthService) GetProfile(ctx context.Context, userID uint64) (*response.UserInfo, error) {
	user, err := dao.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			global.Log.Warn("获取用户信息失败：用户不存在", zap.Uint64("user_id", userID))
			return nil, errs.Wrap(errs.ErrNotFound, "用户不存在", err)
		}
		global.Log.Error("获取用户信息失败", zap.Uint64("user_id", userID), zap.Error(err))
		return nil, err
	}
	return toUserInfo(user), nil
}

// toUserInfo 模型转响应结构体
func toUserInfo(user *database.User) *response.UserInfo {
	return &response.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Email:    user.Email,
		Role:     user.Role,
	}
}
