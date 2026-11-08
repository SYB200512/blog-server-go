package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/database"
	"blog-backend/model/errs"
	"blog-backend/model/request"
	"blog-backend/model/response"
	"blog-backend/utils/cache"
)

// 缓存 TTL 常量。列表类用"短 TTL + 写失效"双保险；meta 较长，靠写失效保证新鲜度。
const (
	articleMetaTTL  = 10 * time.Minute
	nullSentinelTTL = 60 * time.Second
	defaultListTTL  = 60 * time.Second
	archiveTTL      = 30 * time.Minute

	// viewsFlushInterval 浏览量回写 MySQL 的周期；优雅停机时还会强制回写一次
	viewsFlushInterval = 60 * time.Second
	// delayedDeleteDelay 延迟双删间隔：压缩"读请求在写事务提交前读到旧值、
	// 提交后才回填缓存"的竞态窗口
	delayedDeleteDelay = 500 * time.Millisecond
)

// frontListCacheEntry 首页列表缓存载荷。
// response.PageResult.List 是 interface{}，直接序列化会丢类型，
// 所以缓存里用具体类型，返回时再组装 PageResult。
type frontListCacheEntry struct {
	List  []database.Article `json:"list"`
	Total int64              `json:"total"`
}

// ArticleService 文章业务
type ArticleService struct {
	// group 合并并发回源：同一 key 的缓存击穿流量只查一次 DB（防击穿）
	group singleflight.Group
}

// —— 浏览量：Redis INCR 计数 + 定时绝对值回写 MySQL ——

var (
	viewsDirty   = make(map[uint64]struct{})
	viewsDirtyMu sync.Mutex
	flusherOnce  sync.Once
)

func markViewsDirty(id uint64) {
	viewsDirtyMu.Lock()
	viewsDirty[id] = struct{}{}
	viewsDirtyMu.Unlock()
}

// startViewsFlusher 首次计数时启动后台回写协程（进程级单例）
func (s *ArticleService) startViewsFlusher() {
	flusherOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(viewsFlushInterval)
			defer ticker.Stop()
			for range ticker.C {
				FlushArticleViews(context.Background())
			}
		}()
	})
}

// FlushArticleViews 把 Redis 中脏的浏览量计数回写 MySQL（绝对值，幂等，重试安全）。
// 定时协程与优雅停机都会调用；单次失败保留 dirty 标记，下轮重试。
func FlushArticleViews(ctx context.Context) {
	viewsDirtyMu.Lock()
	ids := make([]uint64, 0, len(viewsDirty))
	for id := range viewsDirty {
		ids = append(ids, id)
	}
	viewsDirtyMu.Unlock()

	for _, id := range ids {
		v, ok := cache.GetView(ctx, id)
		if !ok {
			continue // Redis 故障：保留 dirty，下轮再试
		}
		if err := dao.SetArticleViews(ctx, id, v); err != nil {
			global.Log.Warn("回写文章浏览量失败，下轮重试", zap.Uint64("article_id", id), zap.Error(err))
			continue
		}
		viewsDirtyMu.Lock()
		delete(viewsDirty, id)
		viewsDirtyMu.Unlock()
	}
}

// countView 浏览量 +1 并返回最新值。
// base 是计数 key 不存在时的初始化种子（冷路径为 DB 真实值，热路径为 meta 中的快照值）。
// Redis 不可用时返回 base：计数冻结但接口可用，Redis 恢复后自动继续。
func (s *ArticleService) countView(ctx context.Context, id uint64, base int) int {
	s.startViewsFlusher()
	n, ok := cache.AddView(ctx, id, int64(base))
	if !ok {
		return base
	}
	markViewsDirty(id)
	return int(n)
}

// —— 缓存失效 ——

// invalidateArticle 单篇文章变更后的失效：详情 meta + 所有含该文章的列表类缓存。
// 必须在数据库事务提交成功后调用；删除失败只记 warn，由 TTL 兜底。
func invalidateArticle(ctx context.Context, id uint64) {
	cache.Del(ctx,
		cache.ArticleMetaKey(id),
		cache.DefaultListKey(),
		cache.ArchiveKey(),
		cache.CategoryListKey(),
		cache.TagListKey(),
	)
	// 延迟双删：压缩读写竞态的脏窗口。独立于请求生命周期，失败同样由 TTL 兜底
	go func() {
		time.Sleep(delayedDeleteDelay)
		delCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cache.Del(delCtx, cache.ArticleMetaKey(id))
	}()
}

// invalidateArticleLists 新建文章后的列表类缓存失效（尚无详情缓存可删）
func invalidateArticleLists(ctx context.Context) {
	cache.Del(ctx,
		cache.DefaultListKey(),
		cache.ArchiveKey(),
		cache.CategoryListKey(),
		cache.TagListKey(),
	)
}

// —— 查询 ——

// ListArticles 前台文章列表。
// 只缓存"默认首页"这一种组合（page=1, page_size=10, 无关键词/分类/标签）：
// 筛选组合的 key 会爆炸且命中率极低，得不偿失，直连 DB。
func (s *ArticleService) ListArticles(ctx context.Context, q *request.ArticleQuery) (*response.PageResult, error) {
	p := &request.PageParams{Page: q.Page, PageSize: q.PageSize}
	p.Normalize()

	if isDefaultListQuery(q, p) {
		return s.cachedDefaultList(ctx, p)
	}

	articles, total, err := dao.ListArticles(ctx, p, q.Keyword, q.CategoryID, q.TagID, nil, false)
	if err != nil {
		global.Log.Error("查询前台文章列表失败", zap.Error(err))
		return nil, err
	}
	return &response.PageResult{List: articles, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// isDefaultListQuery 是否默认首页参数（与 Normalize 归一化后的默认值比较）
func isDefaultListQuery(q *request.ArticleQuery, p *request.PageParams) bool {
	return q.Keyword == "" && q.CategoryID == 0 && q.TagID == 0 &&
		p.Page == 1 && p.PageSize == 10
}

// cachedDefaultList 默认首页列表：整包缓存 + singleflight 回源
func (s *ArticleService) cachedDefaultList(ctx context.Context, p *request.PageParams) (*response.PageResult, error) {
	key := cache.DefaultListKey()

	var cached frontListCacheEntry
	if cache.Get(ctx, key, &cached) == cache.Hit {
		return &response.PageResult{List: cached.List, Total: cached.Total, Page: p.Page, PageSize: p.PageSize}, nil
	}

	v, err, _ := s.group.Do(key, func() (any, error) {
		// 二次检查：等待 singleflight 期间可能已被其他请求回填
		var e frontListCacheEntry
		if cache.Get(ctx, key, &e) == cache.Hit {
			return &e, nil
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		articles, total, err := dao.ListArticles(dbCtx, p, "", 0, 0, nil, false)
		if err != nil {
			return nil, err
		}
		e = frontListCacheEntry{List: articles, Total: total}
		cache.Set(ctx, key, &e, cache.WithJitter(defaultListTTL))
		return &e, nil
	})
	if err != nil {
		global.Log.Error("查询前台文章列表失败", zap.Error(err))
		return nil, err
	}
	entry := v.(*frontListCacheEntry)
	return &response.PageResult{List: entry.List, Total: entry.Total, Page: p.Page, PageSize: p.PageSize}, nil
}

// AdminListArticles 后台文章列表
func (s *ArticleService) AdminListArticles(ctx context.Context, q *request.ArticleQuery) (*response.PageResult, error) {
	p := &request.PageParams{Page: q.Page, PageSize: q.PageSize}
	p.Normalize()

	articles, total, err := dao.ListArticles(ctx, p, q.Keyword, q.CategoryID, q.TagID, q.Status, true)
	if err != nil {
		global.Log.Error("查询后台文章列表失败", zap.Error(err))
		return nil, err
	}
	return &response.PageResult{List: articles, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// GetArticle 前台文章详情（浏览量 +1）。
//
// 缓存策略：meta（不含正文，约 1KB）走 Redis，正文走 MySQL 主键查询。
// 热路径 2 次往返（Redis GET + 主键 SELECT content），换来：
//   - 缓存里没有 longtext 大字段，彻底避开 bigkey 风险
//   - 正文永远实时，编辑正文无需等待缓存失效
//
// 防护：空值哨兵防穿透（60s），singleflight 防击穿，TTL 抖动防雪崩。
func (s *ArticleService) GetArticle(ctx context.Context, id uint64) (*database.Article, error) {
	key := cache.ArticleMetaKey(id)

	var meta database.Article
	switch cache.Get(ctx, key, &meta) {
	case cache.Null:
		// 空值哨兵：该 id 在数据库里也不存在
		global.Log.Warn("前台文章不存在", zap.Uint64("article_id", id))
		return nil, errs.Wrap(errs.ErrNotFound, "文章不存在", errs.ErrNotFound)
	case cache.Hit:
		return s.assembleFromMeta(ctx, &meta)
	}

	// 未命中：singleflight 合并并发回源
	v, err, _ := s.group.Do(key, func() (any, error) {
		// 二次检查：等待 singleflight 期间可能已被其他请求回填
		var m database.Article
		switch cache.Get(ctx, key, &m) {
		case cache.Null:
			return nil, errs.ErrNotFound
		case cache.Hit:
			return s.assembleFromMeta(ctx, &m)
		}
		// 独立 ctx：不能用请求 ctx，第一个请求被客户端取消会带崩所有等待者
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		article, err := dao.GetPublishedArticleByID(dbCtx, id)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				// 空值缓存：随机 id 枚举的重复请求在 60 秒内只打一次 DB
				cache.Set(ctx, key, "", nullSentinelTTL)
			}
			return nil, err
		}
		// 写 meta（不含正文）；浅拷贝后置空 Content，不影响本次返回的完整对象
		m2 := *article
		m2.Content = ""
		cache.Set(ctx, key, &m2, cache.WithJitter(articleMetaTTL))
		article.Views = s.countView(ctx, id, article.Views)
		return article, nil
	})
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			global.Log.Warn("前台文章不存在", zap.Uint64("article_id", id))
			return nil, errs.Wrap(errs.ErrNotFound, "文章不存在", err)
		}
		global.Log.Error("查询前台文章详情失败", zap.Uint64("article_id", id), zap.Error(err))
		return nil, err
	}
	return v.(*database.Article), nil
}

// assembleFromMeta meta 命中后的拼装：正文查主键 + 浏览量计数
func (s *ArticleService) assembleFromMeta(ctx context.Context, meta *database.Article) (*database.Article, error) {
	content, err := dao.GetArticleContentByID(ctx, meta.ID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			// meta 是残影：文章刚被删除或转为草稿，顺手清掉并返回 404
			cache.Del(ctx, cache.ArticleMetaKey(meta.ID))
			global.Log.Warn("文章 meta 缓存残影，已清除", zap.Uint64("article_id", meta.ID))
			return nil, errs.Wrap(errs.ErrNotFound, "文章不存在", err)
		}
		global.Log.Error("查询文章正文失败", zap.Uint64("article_id", meta.ID), zap.Error(err))
		return nil, err
	}
	meta.Content = content
	meta.Views = s.countView(ctx, meta.ID, meta.Views)
	return meta, nil
}

// AdminGetArticle 后台文章详情（不走缓存：管理端要求实时，且低频）
func (s *ArticleService) AdminGetArticle(ctx context.Context, id uint64) (*database.Article, error) {
	article, err := dao.GetArticleByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			global.Log.Warn("后台文章不存在", zap.Uint64("article_id", id))
			return nil, errs.Wrap(errs.ErrNotFound, "文章不存在", err)
		}
		global.Log.Error("查询后台文章详情失败", zap.Uint64("article_id", id), zap.Error(err))
		return nil, err
	}
	return article, nil
}

// CreateArticle 创建文章
func (s *ArticleService) CreateArticle(ctx context.Context, req *request.ArticleCreateRequest) (*database.Article, error) {
	article := &database.Article{
		Title:      req.Title,
		Summary:    req.Summary,
		Content:    req.Content,
		Cover:      req.Cover,
		CategoryID: req.CategoryID,
		Status:     req.Status,
		IsTop:      req.IsTop,
	}
	// 默认发布
	if article.Status != 0 {
		article.Status = 1
	}
	if err := dao.CreateArticle(ctx, article, req.TagIDs); err != nil {
		global.Log.Error("创建文章失败", zap.String("title", article.Title), zap.Error(err))
		return nil, err
	}
	// 事务提交成功后失效列表类缓存（分类/标签列表含文章计数，一并失效）
	invalidateArticleLists(ctx)
	// 回写图片台账分类：封面 / 正文插图
	ServiceGroupApp.ImageService.MarkArticleImages(ctx, article.Cover, article.Content)
	// 发布且未指定封面时，异步生成封面并回写（不阻塞发布）
	if article.Status == 1 && article.Cover == "" {
		s.generateCoverAsync(article.ID, article.Title, article.Summary)
	}
	return article, nil
}

// generateCoverAsync 后台协程生成封面，成功后回写 cover 字段。
// 这里刻意使用 context.Background() 而不是请求的 ctx：
// 用户发布文章后可能立刻关闭页面，请求 ctx 会被取消，但封面生成属于
// 后台任务，必须独立于请求生命周期，因此自己派生一个超时预算。
func (s *ArticleService) generateCoverAsync(id uint64, title, summary string) {
	// 封面功能未开启时直接跳过，避免无意义的协程与日志
	if !global.Config.AI.Cover.Enabled {
		return
	}
	// 异步任务独立于请求生命周期，用户断开后仍需完成
	timeout := 90 * time.Second
	if global.Config.AI.Cover.CoverTimeout > 0 {
		timeout = time.Duration(global.Config.AI.Cover.CoverTimeout) * time.Second
	}
	go func() {
		// 先建 ctx 再 defer cancel，保证协程退出时释放定时器
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// recover 必须放在 goroutine 边界，否则一次 panic 会带崩整个进程
		defer func() {
			if r := recover(); r != nil {
				global.Log.Error("生成封面协程 panic", zap.Any("error", r), zap.Uint64("article_id", id))
			}
		}()
		cover, err := generateArticleCover(ctx, title, summary)
		if err != nil {
			global.Log.Error("自动生成封面失败（异步）", zap.Uint64("article_id", id), zap.Error(err))
			return
		}
		if cover == "" {
			return
		}
		if err := dao.UpdateArticleCover(ctx, id, cover); err != nil {
			global.Log.Error("回写封面失败", zap.Uint64("article_id", id), zap.Error(err))
		} else {
			global.Log.Info("文章封面已生成", zap.Uint64("article_id", id), zap.String("cover", cover))
			// 封面回写后失效详情与首页列表，否则 TTL 内前台看不到新封面
			cache.Del(ctx, cache.ArticleMetaKey(id), cache.DefaultListKey())
		}
	}()
}

// UpdateArticle 更新文章
func (s *ArticleService) UpdateArticle(ctx context.Context, id uint64, req *request.ArticleUpdateRequest) (*database.Article, error) {
	article, err := dao.GetArticleByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			global.Log.Warn("更新文章时文章不存在", zap.Uint64("article_id", id))
			return nil, errs.Wrap(errs.ErrNotFound, "文章不存在", err)
		}
		global.Log.Error("更新文章时查询失败", zap.Uint64("article_id", id), zap.Error(err))
		return nil, err
	}
	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Summary != "" {
		article.Summary = req.Summary
	}
	if req.Content != "" {
		article.Content = req.Content
	}
	if req.Cover != "" {
		article.Cover = req.Cover
	}
	if req.CategoryID != nil {
		article.CategoryID = req.CategoryID
	}
	if req.Status != nil {
		article.Status = *req.Status
	}
	if req.IsTop != nil {
		article.IsTop = *req.IsTop
	}
	// 已发布且仍未设置封面时，异步生成封面并回写（不阻塞更新）
	if article.Status == 1 && article.Cover == "" {
		s.generateCoverAsync(article.ID, article.Title, article.Summary)
	}
	if err := dao.UpdateArticle(ctx, article, req.TagIDs); err != nil {
		global.Log.Error("更新文章失败", zap.Uint64("article_id", id), zap.Error(err))
		return nil, err
	}
	// 事务提交成功后失效（含延迟双删）
	invalidateArticle(ctx, id)
	// 回写图片台账分类：封面 / 正文插图
	ServiceGroupApp.ImageService.MarkArticleImages(ctx, article.Cover, article.Content)
	return article, nil
}

// DeleteArticle 删除文章（软删除）
func (s *ArticleService) DeleteArticle(ctx context.Context, id uint64) error {
	if _, err := dao.GetArticleByID(ctx, id); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			global.Log.Warn("删除文章时文章不存在", zap.Uint64("article_id", id))
			return errs.Wrap(errs.ErrNotFound, "文章不存在", err)
		}
		global.Log.Error("删除文章时查询失败", zap.Uint64("article_id", id), zap.Error(err))
		return err
	}
	if err := dao.DeleteArticle(ctx, id); err != nil {
		global.Log.Error("删除文章失败", zap.Uint64("article_id", id), zap.Error(err))
		return err
	}
	// 事务提交成功后失效（含延迟双删）；浏览量计数 key 一并清理，
	// 即使软删除被恢复也会从 DB 值重新播种，不会丢历史计数
	invalidateArticle(ctx, id)
	cache.Del(ctx, cache.ViewsKey(id))
	return nil
}

// Archive 时间归档：整包缓存 + singleflight 回源
func (s *ArticleService) Archive(ctx context.Context) (map[string][]database.Article, error) {
	key := cache.ArchiveKey()

	var cached map[string][]database.Article
	if cache.Get(ctx, key, &cached) == cache.Hit {
		return cached, nil
	}

	v, err, _ := s.group.Do(key, func() (any, error) {
		// 二次检查
		var m map[string][]database.Article
		if cache.Get(ctx, key, &m) == cache.Hit {
			return m, nil
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		result, err := dao.ArchiveByMonth(dbCtx)
		if err != nil {
			return nil, err
		}
		cache.Set(ctx, key, result, cache.WithJitter(archiveTTL))
		return result, nil
	})
	if err != nil {
		global.Log.Error("查询时间归档失败", zap.Error(err))
		return nil, err
	}
	return v.(map[string][]database.Article), nil
}
