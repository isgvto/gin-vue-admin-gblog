package global

import (
	"testing"

	"github.com/isgvto/gin-vue-admin-gblog/server/plugin/email/config"
)

func TestGetConfigTracksUpdates(t *testing.T) {
	previousProvider, previousConfig := ConfigProvider, GlobalConfig
	t.Cleanup(func() { ConfigProvider, GlobalConfig = previousProvider, previousConfig })
	GlobalConfig = &config.Email{Host: "old.invalid", Secret: "old-test-code"}
	current := config.Email{Host: "first.invalid", Secret: "first-test-code", Port: 465, IsSSL: true}
	ConfigProvider = func() config.Email { return current }
	first := GetConfig()
	current = config.Email{Host: "second.invalid", Secret: "second-test-code", Port: 587, IsLoginAuth: true}
	if GetConfig() != current {
		t.Fatal("configuration update was not observed")
	}
	if first.Host != "first.invalid" || first.Secret != "first-test-code" {
		t.Fatal("an in-flight configuration snapshot was changed")
	}
	ConfigProvider = nil
	if GetConfig() != *GlobalConfig {
		t.Fatal("standalone plugin configuration fallback failed")
	}
}
