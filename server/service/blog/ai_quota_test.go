package blog

import (
	"context"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/redis/go-redis/v9"
)

func TestAiQuotaWindow(t *testing.T) {
	now := time.Date(2026, 12, 31, 23, 59, 0, 0, time.FixedZone("CST", 8*3600))
	key, expires := aiQuotaWindow(7, now)
	if key != "ai:quota:7:20261231" || expires-now.Unix() != 60 {
		t.Fatal(key, expires)
	}
}

func TestAiRequestLimits(t *testing.T) {
	if ValidateAiChatRequest(&AiChatRequest{Action: "hack"}) == nil {
		t.Fatal("invalid action")
	}
	if ValidateAiRequestSize(&AiChatRequest{Instruction: strings.Repeat("字", 8001)}) == nil {
		t.Fatal("instruction limit")
	}
	if ValidateAiRequestSize(&AiChatRequest{History: []AiChatMessage{{Content: strings.Repeat("字", 8001)}}}) == nil {
		t.Fatal("history limit")
	}
}

func TestAiQuotaUnavailable(t *testing.T) {
	oldLimit, oldRedis := global.GVA_CONFIG.AI.DailyLimit, global.GVA_REDIS
	defer func() { global.GVA_CONFIG.AI.DailyLimit, global.GVA_REDIS = oldLimit, oldRedis }()
	svc := &AiService{}
	global.GVA_CONFIG.AI.DailyLimit, global.GVA_REDIS = 5, nil
	if _, err := svc.ConsumeQuota(context.Background(), 1); err == nil {
		t.Fatal("must not bypass enabled limit")
	}
	if _, err := svc.QuotaStatus(context.Background(), 1); err == nil {
		t.Fatal("must report unavailable")
	}
	global.GVA_CONFIG.AI.DailyLimit = 0
	if remain, err := svc.ConsumeQuota(context.Background(), 1); err != nil || remain != -1 {
		t.Fatal(remain, err)
	}
}

// 使用专用临时 Redis，禁止指向业务实例。未配置时跳过集成用例。
func TestAiQuotaRedis(t *testing.T) {
	addr := os.Getenv("AI_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set AI_TEST_REDIS_ADDR to an isolated Redis")
	}
	client := redis.NewClient(&redis.Options{Addr: addr, DB: 15})
	defer client.Close()
	oldLimit, oldRedis := global.GVA_CONFIG.AI.DailyLimit, global.GVA_REDIS
	defer func() { global.GVA_CONFIG.AI.DailyLimit, global.GVA_REDIS = oldLimit, oldRedis }()
	global.GVA_CONFIG.AI.DailyLimit, global.GVA_REDIS = 5, client
	ctx := context.Background()
	const userID = 987654321
	key, _ := aiQuotaWindow(userID, time.Now())
	defer client.Del(ctx, key)
	svc := &AiService{}
	for i := 0; i < 3; i++ {
		if remain, err := svc.QuotaStatus(ctx, userID); err != nil || remain != 5 {
			t.Fatal(remain, err)
		}
	}
	if n, _ := client.Exists(ctx, key).Result(); n != 0 {
		t.Fatal("status wrote quota")
	}
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.ConsumeQuota(ctx, userID); err == nil {
				accepted.Add(1)
			}
		}()
	}
	wg.Wait()
	if accepted.Load() != 5 {
		t.Fatal("accepted", accepted.Load())
	}
	if count, _ := client.Get(ctx, key).Int(); count != 5 {
		t.Fatal("count inflated", count)
	}
	if ttl, _ := client.TTL(ctx, key).Result(); ttl <= 0 || ttl > 24*time.Hour {
		t.Fatal("invalid expiry", ttl)
	}
	if remain, err := svc.QuotaStatus(ctx, userID); err != nil || remain != 0 {
		t.Fatal(remain, err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := svc.ConsumeQuota(cancelled, userID+1); err == nil {
		t.Fatal("cancel ignored")
	}
	if n, _ := client.Exists(ctx, strings.Replace(key, "987654321", "987654322", 1)).Result(); n != 0 {
		t.Fatal("cancelled request consumed quota")
	}
}
