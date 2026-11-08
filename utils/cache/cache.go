// Package cache 是 Redis 的唯一入口：service 层不直接触碰 global.Redis。
//
// 三条核心语义：
//  1. fail-open：任何 Redis 错误都只记 warn 并视为未命中/忽略，业务接口不感知缓存故障；
//  2. 熔断：连续失败后短路一段时间，避免 Redis 宕机期间每个请求白等超时；
//  3. key 收口：所有 key 由本包拼装，带前缀与版本段，禁止散落各处的字符串拼接。
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"blog-backend/global"
)

// Result 缓存读取结果
type Result int

const (
	// Miss 未命中（含 Redis 故障降级）
	Miss Result = iota
	// Hit 命中，dst 已填充
	Hit
	// Null 命中空值哨兵：表示"数据库里也不存在"，调用方应直接返回 NotFound，防穿透
	Null
)

// nullSentinel 空值哨兵：调用 Set 时传入表示"该 key 对应的数据不存在"。
// 正常缓存值一定是非空 JSON（以 { 或 [ 开头），空串不会被误解为合法值。
const nullSentinel = ""

// bigKeyWarnBytes 写入监控阈值：超过则记 warn，便于尽早发现异常大的缓存值
const bigKeyWarnBytes = 100 * 1024

// —— key 拼装：统一 blog:v1:... 形态。v1 是结构版本段，
// 以后缓存结构变更时升级为 v2，旧 key 靠 TTL 自然淘汰，永远不需要 FLUSHDB。 ——

func key(parts ...any) string {
	return fmt.Sprintf("%s:v1:%s", global.Config.Redis.KeyPrefix, fmt.Sprint(parts...))
}

// ArticleMetaKey 文章元数据（不含正文，正文走 MySQL 主键查询）
func ArticleMetaKey(id uint64) string { return key("pub:article:meta:", id) }

// DefaultListKey 默认首页列表（page=1,page_size=10,无筛选）
func DefaultListKey() string { return key("pub:article:list:default") }

// ArchiveKey 时间归档
func ArchiveKey() string { return key("pub:archive") }

// CategoryListKey 分类列表（含文章数）
func CategoryListKey() string { return key("pub:category:list") }

// TagListKey 标签列表（含文章数）
func TagListKey() string { return key("pub:tag:list") }

// ViewsKey 文章浏览量计数（进程内定时回写 MySQL，不随 LRU 过期策略淘汰热点）
func ViewsKey(id uint64) string { return key("article:views:", id) }

// —— 熔断器：连续 failThreshold 次失败后短路 openDuration ——

const (
	failThreshold = 5
	openDuration  = 30 * time.Second
)

var (
	failStreak atomic.Int32
	openUntil  atomic.Int64 // unix nano；0 表示闭合
)

// allowed 熔断期间直接跳过 Redis。
// 熔断窗口结束后放行一次半开试探：成功则闭合，失败立即重新熔断（见 onFailure）。
func allowed() bool {
	until := openUntil.Load()
	return until == 0 || time.Now().UnixNano() >= until
}

func onSuccess() {
	failStreak.Store(0)
	openUntil.Store(0)
}

func onFailure(op, k string, err error) {
	global.Log.Warn("Redis 操作失败，降级直连数据库",
		zap.String("op", op), zap.String("key", k), zap.Error(err))
	// 半开试探失败：立即重新熔断，不让故障期的请求连续穿透
	if openUntil.Load() != 0 {
		openUntil.Store(time.Now().Add(openDuration).UnixNano())
		return
	}
	if failStreak.Add(1) >= failThreshold {
		failStreak.Store(0)
		openUntil.Store(time.Now().Add(openDuration).UnixNano())
		global.Log.Warn("Redis 连续失败，熔断 30 秒（期间所有缓存操作直接跳过）")
	}
}

// —— 命中率统计：每 5 分钟由 LogHitRateLoop 汇总输出，供上线后调 TTL ——

var (
	hits   atomic.Int64
	misses atomic.Int64
)

// LogHitRateLoop 周期性输出缓存命中统计；无请求时不打日志。由 InitRedis 启动。
func LogHitRateLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		h := hits.Swap(0)
		m := misses.Swap(0)
		if h+m == 0 {
			continue
		}
		global.Log.Info("缓存命中统计",
			zap.Int64("hit", h),
			zap.Int64("miss", m),
			zap.Float64("hit_rate", float64(h)/float64(h+m)),
		)
	}
}

// —— 读写原语 ——

// Get 读取并反序列化。任何错误（含反序列化失败）都降级为 Miss，绝不上抛。
func Get(ctx context.Context, k string, dst any) Result {
	if !allowed() {
		misses.Add(1)
		return Miss
	}
	data, err := global.Redis.Get(ctx, k).Bytes()
	if err != nil {
		misses.Add(1)
		if !errors.Is(err, redis.Nil) {
			onFailure("get", k, err)
		}
		return Miss
	}
	onSuccess()
	hits.Add(1)
	if len(data) == 0 {
		return Null
	}
	if err := json.Unmarshal(data, dst); err != nil {
		// 数据被污染（如手动改过）：按未命中处理，让 DB 值随后覆盖它
		global.Log.Warn("缓存反序列化失败，按未命中处理", zap.String("key", k), zap.Error(err))
		misses.Add(1)
		return Miss
	}
	return Hit
}

// Set 序列化并写入。v 传 nullSentinel（空串）表示空值缓存；失败只记 warn。
func Set(ctx context.Context, k string, v any, ttl time.Duration) {
	if !allowed() {
		return
	}
	var data []byte
	if s, ok := v.(string); ok && s == nullSentinel {
		data = []byte{}
	} else {
		b, err := json.Marshal(v)
		if err != nil {
			global.Log.Warn("缓存序列化失败，跳过写入", zap.String("key", k), zap.Error(err))
			return
		}
		data = b
	}
	if len(data) > bigKeyWarnBytes {
		global.Log.Warn("缓存值超过 100KB，请关注 bigkey 风险",
			zap.String("key", k), zap.Int("bytes", len(data)))
	}
	if err := global.Redis.Set(ctx, k, data, ttl).Err(); err != nil {
		onFailure("set", k, err)
		return
	}
	onSuccess()
}

// Del 删除 key；失败只记 warn（靠 TTL 兜底，不能让缓存故障把写接口拖垮）。
func Del(ctx context.Context, keys ...string) {
	if !allowed() || len(keys) == 0 {
		return
	}
	if err := global.Redis.Del(ctx, keys...).Err(); err != nil {
		onFailure("del", keys[0], err)
		return
	}
	onSuccess()
}

// WithJitter 在 TTL 上加 0~10% 随机偏移，防止批量 key 同刻过期造成雪崩。
func WithJitter(ttl time.Duration) time.Duration {
	return ttl + time.Duration(rand.Int64N(int64(ttl/10)+1))
}

// —— 浏览量计数 ——

// AddView 浏览量 +1 并返回最新值。
// base 是 key 不存在时的初始化种子（SET NX，已存在则无副作用）：
//   - 冷路径调用方传数据库里的真实 views；
//   - 热路径传缓存 meta 里的 views（仅作兜底种子，最坏情况是短暂显示偏小）。
//
// SetNX+Incr 走 Pipeline，单次往返。Redis 故障时返回 (base, false)，调用方原样展示。
func AddView(ctx context.Context, id uint64, base int64) (int64, bool) {
	if !allowed() {
		return base, false
	}
	k := ViewsKey(id)
	pipe := global.Redis.Pipeline()
	pipe.SetNX(ctx, k, base, 0)
	incr := pipe.Incr(ctx, k)
	if _, err := pipe.Exec(ctx); err != nil {
		onFailure("incr", k, err)
		return base, false
	}
	onSuccess()
	return incr.Val(), true
}

// GetView 读取当前计数值，供定时回写 MySQL 使用。
func GetView(ctx context.Context, id uint64) (int64, bool) {
	if !allowed() {
		return 0, false
	}
	k := ViewsKey(id)
	n, err := global.Redis.Get(ctx, k).Int64()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			onFailure("get", k, err)
		}
		return 0, false
	}
	onSuccess()
	return n, true
}
