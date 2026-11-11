package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"blog-backend/global"
)

// 认证会话与普通缓存语义不同：读写失败必须上抛、不能 fail-open，
// 否则 Redis 故障期间已注销的 refresh token 仍可能通过校验。
// 因此这里不走熔断与降级，由调用方决定是拒绝请求还是提示稍后重试。

// refreshSessionKey 每个会话一个 key，天然支持同一用户多设备同时在线。
func refreshSessionKey(userID uint64, sessionID string) string {
	return fmt.Sprintf("%s:v1:auth:session:%d:%s", global.Config.Redis.KeyPrefix, userID, sessionID)
}

// SaveRefreshSession 保存 refresh 会话，ttl 为 refresh token 剩余有效期。
func SaveRefreshSession(ctx context.Context, userID uint64, sessionID, token string, ttl time.Duration) error {
	if err := global.Redis.Set(ctx, refreshSessionKey(userID, sessionID), token, ttl).Err(); err != nil {
		return fmt.Errorf("保存 refresh 会话失败: %w", err)
	}
	return nil
}

// GetRefreshSession 读取 refresh 会话；found=false 表示会话不存在或已过期。
func GetRefreshSession(ctx context.Context, userID uint64, sessionID string) (string, bool, error) {
	token, err := global.Redis.Get(ctx, refreshSessionKey(userID, sessionID)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("读取 refresh 会话失败: %w", err)
	}
	return token, true, nil
}

// DeleteRefreshSession 删除 refresh 会话，使其立即失效。
func DeleteRefreshSession(ctx context.Context, userID uint64, sessionID string) error {
	if err := global.Redis.Del(ctx, refreshSessionKey(userID, sessionID)).Err(); err != nil {
		return fmt.Errorf("删除 refresh 会话失败: %w", err)
	}
	return nil
}
