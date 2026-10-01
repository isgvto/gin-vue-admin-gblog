package blog

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	blogService "github.com/isgvto/gin-vue-admin-gblog/server/service/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
	"go.uber.org/zap"
)

type AiApi struct{}

// Status 探测 AI 是否可用 + 当日剩余额度。
func (a *AiApi) Status(c *gin.Context) {
	ok, name := aiService.Available()
	data := gin.H{"enabled": ok, "model": name}
	if userID := utils.GetUserID(c); userID != 0 {
		remain, err := aiService.QuotaStatus(c.Request.Context(), userID)
		if err != nil {
			data["quotaError"] = err.Error()
		} else {
			data["dailyRemain"] = remain
		}
	}
	response.OkWithData(data, c)
}

// Chat 写作助手统一流式入口。
func (a *AiApi) Chat(c *gin.Context) {
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	c.Request = c.Request.WithContext(ctx)
	var req blogReq.AiChatRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if err := blogService.ValidateAiChatRequest(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if ok, reason := aiService.Available(); !ok {
		response.FailWithMessage(reason, c)
		return
	}
	if _, err := aiService.ConsumeQuota(c.Request.Context(), utils.GetUserID(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	stream, err := aiService.ChatStream(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("AI 对话启动失败", zap.String("action", req.Action), zap.Error(err))
		if !c.Writer.Written() {
			response.FailWithMessage(aiFriendlyError(err), c)
		}
		return
	}
	defer stream.Close()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.FailWithMessage("当前响应不支持流式输出", c)
		return
	}
	prepareBlogAiSSEHeaders(c)
	c.Status(http.StatusOK)
	flusher.Flush()
	if renderBlogAiSSE(c, sse.Event{Event: "context", Data: aiService.ContextInfo(&req)}) != nil {
		return
	}

	finishReason := "unknown"
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			global.GVA_LOG.Error("AI 流式输出中断", zap.Error(err))
			if c.Request.Context().Err() != nil {
				return // 客户端主动断开，无需再发事件
			}
			_ = renderBlogAiSSE(c, sse.Event{Event: "error", Data: gin.H{"message": "模型输出中断，请重试"}})
			return
		}

		if chunk.ResponseMeta != nil && chunk.ResponseMeta.FinishReason != "" {
			finishReason = strings.ToLower(chunk.ResponseMeta.FinishReason)
		}
		if len(chunk.ToolCalls) > 0 {
			for _, tc := range chunk.ToolCalls {
				args := map[string]any{}
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				if renderBlogAiSSE(c, sse.Event{Event: "tool", Data: gin.H{"name": tc.Function.Name, "args": args}}) != nil {
					return
				}
			}
			continue
		}
		if chunk.Content != "" {
			if renderBlogAiSSE(c, sse.Event{Event: "message", Data: gin.H{"delta": chunk.Content}}) != nil {
				return
			}
		}
	}
	_ = renderBlogAiSSE(c, sse.Event{Event: "done", Data: gin.H{"finishReason": finishReason}})
}

// Summary 生成文章摘要。
func (a *AiApi) Summary(c *gin.Context) {
	var req blogReq.AiChatRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		response.FailWithMessage("正文内容不能为空", c)
		return
	}
	if err := blogService.ValidateAiRequestSize(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if ok, reason := aiService.Available(); !ok {
		response.FailWithMessage(reason, c)
		return
	}
	if _, err := aiService.ConsumeQuota(c.Request.Context(), utils.GetUserID(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	summary, err := aiService.GenerateSummary(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("生成摘要失败", zap.Error(err))
		response.FailWithMessage(aiFriendlyError(err), c)
		return
	}
	req.Action = "summary"
	response.OkWithData(gin.H{"summary": summary, "context": aiService.ContextInfo(&req)}, c)
}

// SuggestTags 推荐分类与标签。
func (a *AiApi) SuggestTags(c *gin.Context) {
	var req blogReq.AiChatRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		response.FailWithMessage("正文内容不能为空", c)
		return
	}
	if err := blogService.ValidateAiRequestSize(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if ok, reason := aiService.Available(); !ok {
		response.FailWithMessage(reason, c)
		return
	}
	if _, err := aiService.ConsumeQuota(c.Request.Context(), utils.GetUserID(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	suggestion, err := aiService.SuggestTags(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("推荐标签失败", zap.Error(err))
		response.FailWithMessage(aiFriendlyError(err), c)
		return
	}
	response.OkWithData(suggestion, c)
}

// ---- SSE 辅助（照抄 sys_auto_code_sse.go 模式）----

// aiFriendlyError 将框架内部错误转成用户可读的提示，细节留在日志里。
func aiFriendlyError(err error) string {
	msg := err.Error()
	switch {
	case strings.HasPrefix(msg, "模型未返回"), strings.HasPrefix(msg, "模型推荐结果"), strings.HasPrefix(msg, "模型输出未完整"), strings.HasPrefix(msg, "分类标签目录过大"):
		return msg
	case strings.Contains(msg, "exceeds max steps"):
		return "AI 处理步数超限，请简化指令或稍后重试"
	case strings.Contains(msg, "未配置可用的大模型"):
		return "未配置可用的大模型，请在「AI 模型配置」中设置默认模型"
	case strings.Contains(msg, "connection refused"), strings.Contains(msg, "timeout"), strings.Contains(msg, "context deadline"):
		return "模型服务连接失败，请检查模型配置或稍后重试"
	default:
		return "AI 调用失败，请重试；若持续失败请联系管理员查看服务日志"
	}
}

func prepareBlogAiSSEHeaders(c *gin.Context) {
	header := c.Writer.Header()
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	header.Set("Cache-Control", "no-cache, no-transform")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
}

func renderBlogAiSSE(c *gin.Context, event sse.Event) error {
	if err := event.Render(c.Writer); err != nil {
		return err
	}
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}
