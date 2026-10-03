package response

import (
	"time"

	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
)

type SearchBlogItem struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ArchiveBlogItem struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	CreateTime time.Time `json:"createTime"`
	IsTop      bool      `json:"top"`
	Password   string    `json:"password"`
	Privacy    bool      `json:"privacy"`
}

type BlogInfoItem struct {
	ID          uint                `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Preview     string              `json:"preview"`
	CreateTime  time.Time           `json:"createTime"`
	Views       int                 `json:"views"`
	Words       int                 `json:"words"`
	ReadTime    int                 `json:"readTime"`
	Top         bool                `json:"top"`
	Password    string              `json:"password"`
	Privacy     bool                `json:"privacy"`
	Category    *blogModel.Category `json:"category,omitempty"`
	Tags        []blogModel.Tag     `json:"tags"`
}

type BlogDetail struct {
	ID             uint                `json:"id"`
	Title          string              `json:"title"`
	FirstPicture   string              `json:"firstPicture"`
	Content        string              `json:"content"`
	Description    string              `json:"description"`
	Published      bool                `json:"published"`
	Recommend      bool                `json:"recommend"`
	Appreciation   bool                `json:"appreciation"`
	CommentEnabled bool                `json:"commentEnabled"`
	Top            bool                `json:"top"`
	CreateTime     time.Time           `json:"createTime"`
	UpdateTime     time.Time           `json:"updateTime"`
	Views          int                 `json:"views"`
	Words          int                 `json:"words"`
	ReadTime       int                 `json:"readTime"`
	Password       string              `json:"password"`
	Privacy        bool                `json:"privacy"`
	Category       *blogModel.Category `json:"category,omitempty"`
	Tags           []blogModel.Tag     `json:"tags"`
}

type AdminArticleListItem struct {
	ID             uint                `json:"id"`
	Title          string              `json:"title"`
	FirstPicture   string              `json:"firstPicture"`
	Description    string              `json:"description"`
	Published      bool                `json:"published"`
	Recommend      bool                `json:"recommend"`
	Appreciation   bool                `json:"appreciation"`
	CommentEnabled bool                `json:"commentEnabled"`
	Top            bool                `json:"top"`
	CreateTime     time.Time           `json:"createTime"`
	UpdateTime     time.Time           `json:"updateTime"`
	Views          int                 `json:"views"`
	Words          int                 `json:"words"`
	ReadTime       int                 `json:"readTime"`
	Password       string              `json:"password"`
	Category       *blogModel.Category `json:"category,omitempty"`
	CategoryID     uint                `json:"categoryId"`
}

type AdminArticleDetail struct {
	ID             uint                `json:"id"`
	Title          string              `json:"title"`
	FirstPicture   string              `json:"firstPicture"`
	Content        string              `json:"content"`
	Description    string              `json:"description"`
	Published      bool                `json:"published"`
	Recommend      bool                `json:"recommend"`
	Appreciation   bool                `json:"appreciation"`
	CommentEnabled bool                `json:"commentEnabled"`
	Top            bool                `json:"top"`
	CreateTime     time.Time           `json:"createTime"`
	UpdateTime     time.Time           `json:"updateTime"`
	Views          int                 `json:"views"`
	Words          int                 `json:"words"`
	ReadTime       int                 `json:"readTime"`
	Password       string              `json:"password"`
	Category       *blogModel.Category `json:"category,omitempty"`
	Tags           []blogModel.Tag     `json:"tags"`
	Cate           any                 `json:"cate"`
	TagList        []any               `json:"tagList"`
}
