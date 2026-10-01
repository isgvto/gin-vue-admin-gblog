package example

import "github.com/isgvto/gin-vue-admin-gblog/server/service"

type ApiGroup struct {
	CustomerApi

	AttachmentCategoryApi
	FileUploadAndDownloadApi
}

var (
	customerService = service.ServiceGroupApp.ExampleServiceGroup.CustomerService

	attachmentCategoryService    = service.ServiceGroupApp.ExampleServiceGroup.AttachmentCategoryService
	fileUploadAndDownloadService = service.ServiceGroupApp.ExampleServiceGroup.FileUploadAndDownloadService
)
