package blog

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"gorm.io/gorm"
)

func cleanupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&blogModel.SiteSetting{}, &blogModel.About{}); err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	t.Cleanup(func() { _ = conn.Close() })
	return db
}

func cleanupRow(name, value string, kind int) blogModel.SiteSetting {
	return blogModel.SiteSetting{NameEn: &name, Value: &value, Type: &kind}
}

func TestRetiredSiteSettingCleanupAndLegacyWrites(t *testing.T) {
	db := cleanupTestDB(t)
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })
	rows := []blogModel.SiteSetting{}
	for _, name := range retiredSiteSettings {
		rows = append(rows, cleanupRow(name, "old", 1))
	}
	rows = append(rows, cleanupRow("malfunctionText", "漫漫迷途，终有归途", 1), cleanupRow("footerImgUrl", "/img/qr.png", 1), cleanupRow("netease", "profile link", 2), cleanupRow("badge", `{"color":"blue","subject":"GBlog","title":"blog","url":"/","value":"Open Source","extra":"keep"}`, 3), cleanupRow("badge", "invalid legacy json", 3))
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	about := []blogModel.About{{NameEn: "musicId", Value: "123"}, {NameEn: "title", Value: "关于我"}, {NameEn: "content", Value: "正文"}}
	if err := db.Create(&about).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := CleanupRetiredSiteSettings(db); err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	db.Model(&blogModel.SiteSetting{}).Where("name_en IN ?", retiredSiteSettings).Count(&count)
	if count != 0 {
		t.Fatal("retired settings remain")
	}
	db.Model(&blogModel.About{}).Where("name_en = ?", "musicId").Count(&count)
	if count != 0 {
		t.Fatal("retired about player remains")
	}
	var badge blogModel.SiteSetting
	badgeID := rows[len(rows)-2].ID
	if err := db.First(&badge, badgeID).Error; err != nil {
		t.Fatal(err)
	}
	var fields map[string]string
	if err := json.Unmarshal([]byte(*badge.Value), &fields); err != nil {
		t.Fatal(err)
	}
	if _, exists := fields["color"]; exists {
		t.Fatal("color remains")
	}
	if fields["subject"] != "GBlog" || fields["extra"] != "keep" || fields["url"] != "/" {
		t.Fatalf("badge content changed: %v", fields)
	}
	var malformed blogModel.SiteSetting
	if err := db.First(&malformed, rows[len(rows)-1].ID).Error; err != nil || *malformed.Value != "invalid legacy json" {
		t.Fatal("malformed legacy badge was destroyed")
	}
	grouped, err := (&SiteSettingService{}).GetGrouped()
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range grouped {
		for _, row := range group {
			if isRetiredSiteSetting(row.NameEn) {
				t.Fatal("retired default was recreated")
			}
		}
	}
	oldSetting := cleanupRow("bg1", "should not return", 1)
	oldBadge := cleanupRow("badge", `{"color":"red","subject":"Edited","value":"Value"}`, 3)
	if err := (&SiteSettingService{}).UpdateAll(blogReq.SiteSettingBatchUpdate{Settings: []blogReq.SiteSettingUpsert{{NameEn: oldSetting.NameEn, Value: oldSetting.Value, Type: oldSetting.Type}, {ID: badgeID, NameEn: oldBadge.NameEn, Value: oldBadge.Value, Type: oldBadge.Type}}}); err != nil {
		t.Fatal(err)
	}
	db.Model(&blogModel.SiteSetting{}).Where("name_en = ?", "bg1").Count(&count)
	if count != 0 {
		t.Fatal("legacy request reintroduced a background setting")
	}
	if err := db.First(&badge, badgeID).Error; err != nil {
		t.Fatal(err)
	}
	fields = nil
	_ = json.Unmarshal([]byte(*badge.Value), &fields)
	if _, exists := fields["color"]; exists || fields["subject"] != "Edited" {
		t.Fatal("legacy badge write was not sanitized")
	}
	if err := (&AboutService{}).UpdateValues(map[string]string{"musicId": "999", "title": "新标题"}); err != nil {
		t.Fatal(err)
	}
	db.Model(&blogModel.About{}).Where("name_en = ?", "musicId").Count(&count)
	if count != 0 {
		t.Fatal("legacy request restored about music")
	}
}

func TestRetiredSiteSettingCleanupRollsBack(t *testing.T) {
	db := cleanupTestDB(t)
	rows := []blogModel.SiteSetting{cleanupRow("bg1", "old", 1), cleanupRow("badge", `{"color":"blue","subject":"keep"}`, 3)}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	failure := errors.New("simulated badge update failure")
	if err := db.Callback().Update().Before("gorm:update").Register("fail_cleanup_update", func(tx *gorm.DB) { tx.AddError(failure) }); err != nil {
		t.Fatal(err)
	}
	if err := CleanupRetiredSiteSettings(db); !errors.Is(err, failure) {
		t.Fatalf("expected rollback failure, got %v", err)
	}
	var count int64
	db.Model(&blogModel.SiteSetting{}).Where("name_en = ?", "bg1").Count(&count)
	if count != 1 {
		t.Fatal("partial deletion committed after failure")
	}
}

func TestActiveSiteSettingsDoesNotMutateStoredRows(t *testing.T) {
	rows := []blogModel.SiteSetting{cleanupRow("playlistId", "old", 1), cleanupRow("badge", `{"color":"blue","subject":"keep"}`, 3), cleanupRow("malfunctionText", "保留", 1)}
	active := activeSiteSettings(rows)
	if len(active) != 2 || *active[1].Value != "保留" {
		t.Fatal("wrong fields filtered")
	}
	if *rows[1].Value != `{"color":"blue","subject":"keep"}` {
		t.Fatal("response cleanup mutated source")
	}
	var fields map[string]string
	_ = json.Unmarshal([]byte(*active[0].Value), &fields)
	if _, exists := fields["color"]; exists {
		t.Fatal("public response exposes badge color")
	}
}
