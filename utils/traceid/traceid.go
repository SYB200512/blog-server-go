package traceid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

// ctxKey 未导出类型，避免与其它包的 context key 冲突
type ctxKey struct{}

// With 生成 trace id 并写入 ctx
func With(ctx context.Context) context.Context {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// 随机源异常时退化为时间戳，保证链路仍有标识
		return context.WithValue(ctx, ctxKey{}, strconv.FormatInt(time.Now().UnixNano(), 10))
	}
	return context.WithValue(ctx, ctxKey{}, hex.EncodeToString(buf[:]))
}

// From 从 ctx 取出 trace id，取不到返回空字符串
func From(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}
