package blog

import (
	"sort"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	blogResp "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/response"
)

type ArchiveService struct{}

func (s *ArchiveService) GetArchive(info blogReq.ArchiveSearch) (map[string]interface{}, error) {
	if info.Page <= 0 {
		info.Page = 1
	}
	if info.PageSize <= 0 {
		info.PageSize = 1000
	}

	var blogs []blogModel.Blog
	if err := global.GVA_DB.
		Where("is_published = ?", true).
		Order("create_time desc").
		Limit(info.PageSize).
		Offset((info.Page - 1) * info.PageSize).
		Find(&blogs).Error; err != nil {
		return nil, err
	}

	blogMap := map[string][]blogResp.ArchiveBlogItem{}
	for _, item := range blogs {
		key := item.CreateTime.Format("2006年01月")
		archiveItem := blogResp.ArchiveBlogItem{
			ID:         item.ID,
			Title:      item.Title,
			CreateTime: item.CreateTime,
			IsTop:      item.IsTop,
		}
		if item.Password != nil && *item.Password != "" {
			archiveItem.Privacy = true
		}
		blogMap[key] = append(blogMap[key], archiveItem)
	}

	keys := make([]string, 0, len(blogMap))
	for key := range blogMap {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		ti, errI := time.Parse("2006年01月", keys[i])
		tj, errJ := time.Parse("2006年01月", keys[j])
		if errI != nil || errJ != nil {
			return keys[i] > keys[j]
		}
		return ti.After(tj)
	})

	orderedMap := make(map[string][]blogResp.ArchiveBlogItem, len(keys))
	for _, key := range keys {
		orderedMap[key] = blogMap[key]
	}

	var count int64
	if err := global.GVA_DB.Model(&blogModel.Blog{}).Where("is_published = ?", true).Count(&count).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"blogMap": orderedMap,
		"count":   count,
	}, nil
}
