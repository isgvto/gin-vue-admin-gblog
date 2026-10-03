package system

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	systemReq "github.com/isgvto/gin-vue-admin-gblog/server/model/system/request"
	"gorm.io/gorm"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubBindingOnlyUpdatesCurrentUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	previous := global.GVA_DB
	global.GVA_DB = db
	defer func() { global.GVA_DB = previous }()
	if err = db.AutoMigrate(&system.SysUser{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&system.SysUser{GVA_MODEL: global.GVA_MODEL{ID: 1}, NickName: "keep", GitHubUsername: "old"})
	db.Create(&system.SysUser{GVA_MODEL: global.GVA_MODEL{ID: 2}, GitHubUsername: "other"})
	invoke := func(body string) int {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/user/github", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("claims", &systemReq.CustomClaims{BaseClaims: systemReq.BaseClaims{ID: 1}})
		(&BaseApi{}).SetGitHubProfile(c)
		var output struct{ Code int }
		if err := json.Unmarshal(w.Body.Bytes(), &output); err != nil {
			t.Fatal(err)
		}
		return output.Code
	}
	if invoke(`{"id":2,"username":"demo"}`) != 0 {
		t.Fatal("binding failed")
	}
	var current, other system.SysUser
	db.First(&current, 1)
	db.First(&other, 2)
	if current.GitHubUsername != "demo" || current.NickName != "keep" || other.GitHubUsername != "other" {
		t.Fatal("unrelated account data changed")
	}
	if invoke(`{"username":"../bad"}`) == 0 {
		t.Fatal("invalid login accepted")
	}
	if invoke(`{"username":""}`) != 0 {
		t.Fatal("unbind failed")
	}
	db.First(&current, 1)
	if current.GitHubUsername != "" {
		t.Fatal("unbind did not persist empty value")
	}
}
