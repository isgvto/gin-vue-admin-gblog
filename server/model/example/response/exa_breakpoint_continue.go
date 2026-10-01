package response

import "github.com/isgvto/gin-vue-admin-gblog/server/model/example"

type FilePathResponse struct {
	FilePath string `json:"filePath"`
}

type FileResponse struct {
	File example.ExaFile `json:"file"`
}
