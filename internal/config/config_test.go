package config

import "testing"

func TestLoadRejectsPublicBindWithoutToken(t *testing.T) {
	t.Setenv("VPS_INSPECTOR_ADDR", "0.0.0.0:8719")
	t.Setenv("VPS_INSPECTOR_AUTH_TOKEN", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected public bind without token to fail")
	}
}

func TestLoadAllowsLoopbackWithoutToken(t *testing.T) {
	t.Setenv("VPS_INSPECTOR_ADDR", "127.0.0.1:8719")
	t.Setenv("VPS_INSPECTOR_AUTH_TOKEN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected loopback config to load: %v", err)
	}
	if cfg.Addr != "127.0.0.1:8719" {
		t.Fatalf("unexpected addr: %s", cfg.Addr)
	}
}
