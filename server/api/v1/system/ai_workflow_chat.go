package system

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	aiService "github.com/isgvto/gin-vue-admin-gblog/server/service/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
)

// WorkflowChat runs requirement analysis or prompt planning using a database model.
// @Tags AI需求工作流
// @Summary 使用配置的模型进行需求分析或工作流对话
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body ai.WorkflowChatRequest true "本轮问题和上下文"
// @Success 200 {object} response.Response
// @Router /autoCode/aiWorkflowChat [post]
func (a *AutoCodeApi) WorkflowChat(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 512<<10)
	var req aiModel.WorkflowChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("工作流请求格式错误或内容过长", c)
		return
	}
	run, err := (aiService.WorkflowService{}).Prepare(c.Request.Context(), utils.GetUserID(c), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(run.Config.TimeoutSeconds)*time.Second)
	defer cancel()
	metadata := gin.H{"conversation_id": run.ConversationID, "message_id": run.MessageID, "model": run.Model.Name + " / " + run.Model.Model}
	finish := func(text string) (gin.H, error) {
		structured, err := aiService.ValidateWorkflowResult(req.Mode, text)
		if err != nil {
			return nil, err
		}
		canonical, err := json.MarshalIndent(structured, "", "  ")
		if err != nil {
			return nil, err
		}
		payload := gin.H{"event": "done", "answer": string(canonical), "structured": structured}
		for key, value := range metadata {
			payload[key] = value
		}
		return payload, nil
	}
	if req.ResponseMode != "streaming" {
		text, err := run.Generate(ctx)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		payload, err := finish(text)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		response.OkWithData(payload, c)
		return
	}
	stream, err := run.Stream(ctx)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	defer stream.Close()
	prepareSSEHeaders(c)
	c.Status(http.StatusOK)
	contextEvent := gin.H{"event": "context"}
	for key, value := range metadata {
		contextEvent[key] = value
	}
	if renderSSE(c, sse.Event{Event: "context", Data: contextEvent}) != nil {
		return
	}
	fail := func(err error) {
		if c.Request.Context().Err() == nil {
			_ = renderSSE(c, sse.Event{Event: "error", Data: gin.H{"event": "error", "message": run.SafeError(err).Error()}})
		}
	}
	var text strings.Builder
	pending := ""
	emitPending := func(final bool) error {
		pending = strings.ReplaceAll(pending, run.Model.APIKey, "[已脱敏]")
		cut := len(pending)
		if !final {
			cut -= len(run.Model.APIKey) - 1
			if cut <= 0 {
				return nil
			}
			for cut < len(pending) && cut > 0 && !utf8.RuneStart(pending[cut]) {
				cut--
			}
		}
		delta := pending[:cut]
		pending = pending[cut:]
		if delta == "" {
			return nil
		}
		return renderSSE(c, sse.Event{Event: "message", Data: gin.H{"event": "message", "delta": delta}})
	}
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fail(err)
			return
		}
		if chunk == nil || chunk.Content == "" {
			continue
		}
		if text.Len()+len(chunk.Content) > 256<<10 {
			fail(errors.New("模型输出过长，请拆分需求"))
			return
		}
		text.WriteString(chunk.Content)
		// Hold a tail so credentials spanning multiple chunks are also redacted.
		pending += chunk.Content
		if emitPending(false) != nil {
			return
		}
	}
	payload, err := finish(strings.ReplaceAll(text.String(), run.Model.APIKey, "[已脱敏]"))
	if err != nil {
		fail(err)
		return
	}
	if emitPending(true) != nil {
		return
	}
	_ = renderSSE(c, sse.Event{Event: "done", Data: payload})
}
