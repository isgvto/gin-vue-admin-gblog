package blog

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/redis/go-redis/v9"
)

// 一个生成请求占用一次；失败或主动取消不退回，避免重复请求绕过每日限制。
// 状态查询只读。检查、递增和设置过期时间在同一个 Redis 脚本中完成。
const consumeAiQuotaScript = `
local count = tonumber(redis.call('GET', KEYS[1]) or '0')
local limit = tonumber(ARGV[1])
if count >= limit then return -1 end
count = redis.call('INCR', KEYS[1])
redis.call('EXPIREAT', KEYS[1], ARGV[2])
return limit - count
`

func aiQuotaWindow(userID uint, now time.Time) (string, int64) {
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return fmt.Sprintf("ai:quota:%d:%s", userID, now.Format("20060102")), next.Unix()
}

func (s *AiService) QuotaStatus(ctx context.Context, userID uint) (int64, error) {
	limit := global.GVA_CONFIG.AI.DailyLimit
	if limit <= 0 {
		return -1, nil
	}
	if global.GVA_REDIS == nil {
		return 0, errors.New("AI 配额服务不可用，请稍后重试")
	}
	key, _ := aiQuotaWindow(userID, time.Now())
	value, err := global.GVA_REDIS.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return int64(limit), nil
	}
	if err != nil {
		return 0, errors.New("AI 配额服务不可用，请稍后重试")
	}
	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil || count < 0 {
		return 0, errors.New("AI 配额数据异常，请联系管理员")
	}
	return max(0, int64(limit)-count), nil
}

func (s *AiService) ConsumeQuota(ctx context.Context, userID uint) (int64, error) {
	limit := global.GVA_CONFIG.AI.DailyLimit
	if limit <= 0 {
		return -1, nil
	}
	if global.GVA_REDIS == nil {
		return 0, errors.New("AI 配额服务不可用，请稍后重试")
	}
	key, expires := aiQuotaWindow(userID, time.Now())
	remain, err := global.GVA_REDIS.Eval(ctx, consumeAiQuotaScript, []string{key}, limit, expires).Int64()
	if err != nil {
		return 0, errors.New("AI 配额服务不可用，请稍后重试")
	}
	if remain < 0 {
		return 0, fmt.Errorf("今日 AI 调用次数已达上限（%d 次/日）", limit)
	}
	return remain, nil
}
