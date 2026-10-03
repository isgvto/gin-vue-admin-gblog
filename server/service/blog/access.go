package blog

import (
	"errors"
	"strconv"
	"strings"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	systemModel "github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
)

type blogAccessContext struct {
	IsAdmin  bool
	Username string
	BlogID   *uint
}

var ErrPasswordProtected = errors.New("password protected")

func parseBlogAccessToken(rawToken string) (blogAccessContext, error) {
	claims, err := utils.ParseBlogToken(rawToken)
	if err != nil {
		return blogAccessContext{}, err
	}

	if strings.HasPrefix(claims.Subject, utils.BlogAdminPrefix) {
		return blogAccessContext{
			IsAdmin:  true,
			Username: strings.TrimPrefix(claims.Subject, utils.BlogAdminPrefix),
		}, nil
	}

	if strings.HasPrefix(claims.Subject, utils.BlogAccessPrefix) {
		blogIDStr := strings.TrimPrefix(claims.Subject, utils.BlogAccessPrefix)
		parsedID, parseErr := strconv.ParseUint(blogIDStr, 10, 64)
		if parseErr != nil {
			return blogAccessContext{}, parseErr
		}
		blogID := uint(parsedID)
		return blogAccessContext{BlogID: &blogID}, nil
	}

	return blogAccessContext{}, errors.New("invalid blog token")
}

func ensureBlogReadable(entity blogModel.Blog, rawToken string) (blogAccessContext, error) {
	if entity.Password == nil || strings.TrimSpace(*entity.Password) == "" {
		return blogAccessContext{}, nil
	}

	access, err := parseBlogAccessToken(rawToken)
	if err != nil {
		return blogAccessContext{}, ErrPasswordProtected
	}
	if access.IsAdmin {
		return access, nil
	}
	if access.BlogID != nil && *access.BlogID == entity.ID {
		return access, nil
	}
	return blogAccessContext{}, ErrPasswordProtected
}

func loadBlogAdminUser(username string) (*systemModel.SysUser, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("empty username")
	}
	var user systemModel.SysUser
	if err := global.GVA_DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
