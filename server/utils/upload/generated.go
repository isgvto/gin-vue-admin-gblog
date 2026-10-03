package upload

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"time"
)

// UploadGenerated reuses the configured OSS implementation without an HTTP
// loopback or exposing storage credentials to either browser or model.
func UploadGenerated(ctx context.Context, name string, data []byte) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return "", "", err
	}
	if _, err = part.Write(data); err != nil {
		return "", "", err
	}
	if err = writer.Close(); err != nil {
		return "", "", err
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", "http://upload.invalid", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err = req.ParseMultipartForm(12 << 20); err != nil {
		return "", "", err
	}
	defer req.MultipartForm.RemoveAll()
	file := req.MultipartForm.File["file"][0]
	file.Header.Set("Content-Type", "image/png")
	oss := NewOss()
	if contextual, ok := oss.(interface {
		UploadFileWithContext(context.Context, *multipart.FileHeader) (string, string, error)
	}); ok {
		return contextual.UploadFileWithContext(ctx, file)
	}
	return oss.UploadFile(file)
}
