package blog

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
)

type AuthService struct{}

func (s *AuthService) CreateBlogAccessToken(blogID uint, password string) (string, error) {
	var blog blogModel.Blog
	if err := global.GVA_DB.Select("id,password").First(&blog, blogID).Error; err != nil {
		return "", err
	}
	if blog.Password == nil || strings.TrimSpace(*blog.Password) == "" {
		return "", errors.New("blog is not password protected")
	}
	if *blog.Password != password {
		return "", errors.New("password mismatch")
	}
	return utils.CreateBlogToken(utils.BlogAccessPrefix+uintToString(blogID), 30*24*time.Hour)
}

func uintToString(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}
