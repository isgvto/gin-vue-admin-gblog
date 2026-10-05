package blog

import (
	"context"

	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderBlogSiteSetting = system.InitOrderInternal + 100

type initBlogSiteSetting struct{}

// auto run
func init() {
	system.RegisterInit(initOrderBlogSiteSetting, &initBlogSiteSetting{})
}

func (i *initBlogSiteSetting) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&blogModel.SiteSetting{})
}

func (i *initBlogSiteSetting) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&blogModel.SiteSetting{})
}

func (i *initBlogSiteSetting) InitializerName() string {
	return blogModel.SiteSetting{}.TableName()
}

func (i *initBlogSiteSetting) InitializeData(ctx context.Context) (next context.Context, err error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}

	entities := []blogModel.SiteSetting{
		{NameEn: ptrString("blogName"), NameZh: ptrString("博客名称"), Value: ptrString("Gvto's Blog"), Type: ptrInt(1)},
		{NameEn: ptrString("webTitleSuffix"), NameZh: ptrString("网页标题后缀"), Value: ptrString("- Gvto's Blog"), Type: ptrInt(1)},
		{NameEn: ptrString("footerImgTitle"), NameZh: ptrString("页脚图片标题"), Value: ptrString("扫码手机访问"), Type: ptrInt(1)},
		{NameEn: ptrString("footerImgUrl"), NameZh: ptrString("页脚图片路径"), Value: ptrString("/img/qr.png"), Type: ptrInt(1)},
		{NameEn: ptrString("copyright"), NameZh: ptrString("Copyright"), Value: ptrString("{\"title\":\"Copyright © 2019 - 2024\",\"siteName\":\"Gvto'S BLOG\"}"), Type: ptrInt(1)},
		{NameEn: ptrString("beian"), NameZh: ptrString("ICP备案号"), Value: ptrString("豫ICP备2023001942号-1"), Type: ptrInt(1)},
		{NameEn: ptrString("reward"), NameZh: ptrString("赞赏码"), Value: ptrString("/img/reward.jpg"), Type: ptrInt(1)},
		{NameEn: ptrString("commentAdminFlag"), NameZh: ptrString("博主评论标识"), Value: ptrString("归途"), Type: ptrInt(1)},
		{NameEn: ptrString("avatar"), NameZh: ptrString("头像"), Value: ptrString("/img/avatar.jpg"), Type: ptrInt(2)},
		{NameEn: ptrString("name"), NameZh: ptrString("昵称"), Value: ptrString("Gvto"), Type: ptrInt(2)},
		{NameEn: ptrString("rollText"), NameZh: ptrString("滚动个签"), Value: ptrString("\"惶惶二十载，书剑两无成\",\"天地一逆旅，同悲万古尘\",\"漫漫迷途终有归途\""), Type: ptrInt(2)},
		{NameEn: ptrString("github"), NameZh: ptrString("GitHub"), Value: ptrString("https://github.com/LZMclear"), Type: ptrInt(2)},
		{NameEn: ptrString("telegram"), NameZh: ptrString("Telegram"), Value: ptrString("sss"), Type: ptrInt(2)},
		{NameEn: ptrString("qq"), NameZh: ptrString("QQ"), Value: ptrString("http://wpa.qq.com/msgrd?V=1&uin=3135679861&Menu=no"), Type: ptrInt(2)},
		{NameEn: ptrString("bilibili"), NameZh: ptrString("bilibili"), Value: ptrString("https://space.bilibili.com/473358514?spm_id_from=333.1007.0.0"), Type: ptrInt(2)},
		{NameEn: ptrString("netease"), NameZh: ptrString("网易云音乐"), Value: ptrString("https://music.163.com/#/user/home?id=1881835523"), Type: ptrInt(2)},
		{NameEn: ptrString("email"), NameZh: ptrString("email"), Value: ptrString("3135679861@qq.com"), Type: ptrInt(2)},
		{NameEn: ptrString("favorite"), NameZh: ptrString("自定义"), Value: ptrString("{\"title\":\"sasdad\",\"content\":\"sada\"}"), Type: ptrInt(2)},
		{NameEn: ptrString("badge"), NameZh: ptrString("徽标"), Value: ptrString("{\"subject\":\"asfas\",\"title\":\"asda\",\"url\":\"qwfqwa\",\"value\":\"safas\"}"), Type: ptrInt(3)},
		{NameEn: ptrString("badge"), NameZh: ptrString("徽标"), Value: ptrString("{\"subject\":\"ASDVa\",\"title\":\"asfca\",\"url\":\"ewgwae\",\"value\":\"SDVAS\"}"), Type: ptrInt(3)},
		{NameEn: ptrString("malfunctionText"), NameZh: ptrString("首页故障风文字"), Value: ptrString("Gvto's Blog"), Type: ptrInt(1)},
	}
	if len(entities) == 0 {
		return context.WithValue(ctx, i.InitializerName(), entities), nil
	}
	if err = db.Create(&entities).Error; err != nil {
		return ctx, errors.Wrap(err, blogModel.SiteSetting{}.TableName()+" table data initialize failed")
	}
	next = context.WithValue(ctx, i.InitializerName(), entities)
	return next, nil
}

func (i *initBlogSiteSetting) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	var count int64
	if err := db.Model(&blogModel.SiteSetting{}).Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

func ptrString(value string) *string {
	return &value
}

func ptrInt(value int) *int {
	return &value
}
