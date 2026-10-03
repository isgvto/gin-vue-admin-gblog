package system

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	aiService "github.com/isgvto/gin-vue-admin-gblog/server/service/ai"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Recovery is driven by reads/submissions and also handles tasks interrupted by restart.
func recoverExpiredErrorAnalysis(ctx context.Context) error {
	return global.GVA_DB.WithContext(ctx).Model(&system.SysError{}).
		Where("status = ? AND (analysis_started_at IS NULL OR analysis_started_at < ?)", "处理中", time.Now().Add(-200*time.Second)).
		Updates(map[string]any{"status": "处理失败", "analysis_error": "分析任务超时或服务中断，请重新分析", "analysis_completed_at": time.Now(), "analysis_task": ""}).Error
}

func (s *SysErrorService) GetSysErrorSolution(ctx context.Context, id string) error {
	if err := recoverExpiredErrorAnalysis(ctx); err != nil {
		return err
	}
	var record system.SysError
	if err := global.GVA_DB.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("错误日志不存在")
		}
		return err
	}
	if record.Status == "处理中" {
		return errors.New("该日志正在分析，请等待结果")
	}
	svc := aiService.ErrorAnalysisService{}
	cfg, err := svc.Config()
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return errors.New("错误日志 AI 分析未启用，请在 AI 模型配置中启用")
	}
	model, err := svc.Resolve(cfg)
	if err != nil {
		return err
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	token := hex.EncodeToString(tokenBytes)
	claimed := global.GVA_DB.WithContext(ctx).Model(&system.SysError{}).
		Where("id = ? AND (status IS NULL OR status <> ?)", record.ID, "处理中").
		Updates(map[string]any{"status": "处理中", "analysis_error": "", "analysis_task": token,
			"analysis_started_at": time.Now(), "analysis_completed_at": nil,
			"analysis_model_id": model.ID, "analysis_model": model.Name + " / " + model.Model})
	if claimed.Error != nil {
		return claimed.Error
	}
	if claimed.RowsAffected != 1 {
		return errors.New("该日志已被删除或正在分析")
	}
	go runErrorAnalysis(record, cfg, *model, token)
	return nil
}

func runErrorAnalysis(record system.SysError, cfg aiModel.ErrorAnalysisConfig, model aiModel.AiModelConfig, token string) {
	var result string
	var analysisErr error
	defer func() {
		if recover() != nil {
			analysisErr = errors.New("分析任务异常中断，请重试")
		}
		updates := map[string]any{"analysis_completed_at": time.Now(), "analysis_task": ""}
		if analysisErr != nil {
			updates["status"] = "处理失败"
			updates["analysis_error"] = aiService.SanitizeAnalysisText(analysisErr.Error())
		} else {
			updates["status"] = "处理完成"
			updates["solution"] = result
			updates["solution_model"] = model.Name + " / " + model.Model
			updates["solution_generated_at"] = time.Now()
			updates["analysis_error"] = ""
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Token matching prevents a late response from overwriting a newer analysis.
		if err := global.GVA_DB.WithContext(ctx).Model(&system.SysError{}).
			Where("id = ? AND analysis_task = ?", record.ID, token).Updates(updates).Error; err != nil {
			zap.L().Error("保存错误日志分析结果失败", zap.Uint("id", record.ID), zap.Error(err))
		}
	}()
	var source, info string
	if record.Form != nil {
		source = *record.Form
	}
	if record.Info != nil {
		info = *record.Info
	}
	result, analysisErr = (aiService.ErrorAnalysisService{}).Analyze(context.Background(), cfg, &model, source, info)
}
