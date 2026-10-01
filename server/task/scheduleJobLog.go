package task

import (
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
)

func RecordScheduleJobLog(jobID uint, beanName, methodName, params string, fn func() error) error {
	start := time.Now()
	err := fn()

	status := int8(1)
	var errText *string
	if err != nil {
		status = 0
		text := err.Error()
		errText = &text
	}

	now := time.Now()
	log := blogModel.ScheduleJobLog{
		JobID:      jobID,
		BeanName:   stringPtr(beanName),
		MethodName: stringPtr(methodName),
		Params:     stringPtr(params),
		Status:     status,
		Error:      errText,
		Times:      int(now.Sub(start).Milliseconds()),
		CreateTime: &now,
	}
	if logErr := global.GVA_DB.Create(&log).Error; logErr != nil {
		if err != nil {
			global.GVA_LOG.Sugar().Errorf("record schedule job log failed: %v", logErr)
			return err
		}
		return logErr
	}
	return err
}

func stringPtr(value string) *string {
	return &value
}
