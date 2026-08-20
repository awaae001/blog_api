package config

import (
	"blog_api/src/model"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadLegacyConfigDefaultsPublicAPIsToDisabled(t *testing.T) {
	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "system_config.json"), []byte(`{"system_conf":{"safe_conf":{}}}`), 0o600); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}
	t.Setenv("CONFIG_PATH", configDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load legacy config: %v", err)
	}
	if len(cfg.Safe.EnabledPublicAPIs) != 0 {
		t.Fatalf("enabled public APIs = %v, want empty", cfg.Safe.EnabledPublicAPIs)
	}
}

func TestNormalizeEnabledPublicAPIs(t *testing.T) {
	got := normalizeEnabledPublicAPIs([]string{
		model.PublicAPIFriend,
		model.PublicAPIFriend,
		"frinde",
		" unknown ",
		model.PublicAPIEmail,
	})
	want := []string{model.PublicAPIFriend, model.PublicAPIEmail}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalized keys = %v, want %v", got, want)
	}
}
