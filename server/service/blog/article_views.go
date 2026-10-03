package blog

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const blogViewsRedisKey = "gvto:{blog}:views"

func (s *ArticleService) IncrementViews(id uint, baseViews int) int {
	if !blogRedisAvailable() {
		_ = incrementBlogViewsInDB(id, 1)
		return baseViews + 1
	}

	delta, err := global.GVA_REDIS.HIncrBy(context.Background(), blogViewsRedisKey, strconv.FormatUint(uint64(id), 10), 1).Result()
	if err != nil {
		global.GVA_LOG.Sugar().Errorf("increment blog views in redis failed: %v", err)
		_ = incrementBlogViewsInDB(id, 1)
		return baseViews + 1
	}
	return baseViews + int(delta)
}

func (s *ArticleService) SyncViewsToDatabase() error {
	if !blogRedisAvailable() {
		return nil
	}

	ctx := context.Background()
	syncKey := fmt.Sprintf("%s:sync:%d", blogViewsRedisKey, time.Now().UnixNano())
	if err := global.GVA_REDIS.Rename(ctx, blogViewsRedisKey, syncKey).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}

	viewMap, err := global.GVA_REDIS.HGetAll(ctx, syncKey).Result()
	if err != nil {
		return err
	}
	if len(viewMap) == 0 {
		return global.GVA_REDIS.Del(ctx, syncKey).Err()
	}

	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for idText, viewsText := range viewMap {
			id, err := strconv.ParseUint(idText, 10, 64)
			if err != nil {
				return err
			}
			views, err := strconv.ParseInt(viewsText, 10, 64)
			if err != nil {
				return err
			}
			if views <= 0 {
				continue
			}
			if err := tx.Model(&blogModel.Blog{}).
				Where("id = ?", id).
				UpdateColumn("views", gorm.Expr("views + ?", views)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if restoreErr := mergeBlogViewsBack(ctx, syncKey); restoreErr != nil {
			global.GVA_LOG.Sugar().Errorf("restore blog views from sync key failed: %v", restoreErr)
		}
		return err
	}
	return global.GVA_REDIS.Del(ctx, syncKey).Err()
}

func blogRedisAvailable() bool {
	return global.GVA_CONFIG.System.UseRedis && global.GVA_REDIS != nil
}

func incrementBlogViewsInDB(id uint, views int64) error {
	return global.GVA_DB.Model(&blogModel.Blog{}).
		Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", views)).Error
}

func mergeBlogViewsBack(ctx context.Context, syncKey string) error {
	viewMap, err := global.GVA_REDIS.HGetAll(ctx, syncKey).Result()
	if err != nil {
		return err
	}
	for idText, viewsText := range viewMap {
		views, err := strconv.ParseInt(viewsText, 10, 64)
		if err != nil {
			return err
		}
		if views <= 0 {
			continue
		}
		if err := global.GVA_REDIS.HIncrBy(ctx, blogViewsRedisKey, idText, views).Err(); err != nil {
			return err
		}
	}
	return global.GVA_REDIS.Del(ctx, syncKey).Err()
}
