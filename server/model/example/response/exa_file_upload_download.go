package response

import "github.com/isgvto/gin-vue-admin-gblog/server/model/example"

type ExaFileResponse struct {
	File example.ExaFileUploadAndDownload `json:"file"`
}
