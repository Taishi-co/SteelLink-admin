package database

import (
	"strings"
	"testing"
)

func TestLoadMySQLConfigAllowsLocalDefaultHost(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("GIN_MODE", "")

	originalContainerRuntime := isContainerRuntime
	isContainerRuntime = func() bool { return false }
	t.Cleanup(func() {
		isContainerRuntime = originalContainerRuntime
	})

	cfg, err := loadMySQLConfig()
	if err != nil {
		t.Fatalf("loadMySQLConfig returned error: %v", err)
	}

	if cfg.host != "localhost" {
		t.Fatalf("host = %q, want localhost", cfg.host)
	}
}

func TestLoadMySQLConfigRequiresHostInReleaseMode(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("GIN_MODE", "release")

	originalContainerRuntime := isContainerRuntime
	isContainerRuntime = func() bool { return false }
	t.Cleanup(func() {
		isContainerRuntime = originalContainerRuntime
	})

	_, err := loadMySQLConfig()
	if err == nil {
		t.Fatal("loadMySQLConfig returned nil error, want missing DB_HOST error")
	}

	if !strings.Contains(err.Error(), "DB_HOST is required") {
		t.Fatalf("error = %q, want DB_HOST requirement", err.Error())
	}
}

func TestLoadMySQLConfigRequiresHostInContainer(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("GIN_MODE", "")

	originalContainerRuntime := isContainerRuntime
	isContainerRuntime = func() bool { return true }
	t.Cleanup(func() {
		isContainerRuntime = originalContainerRuntime
	})

	_, err := loadMySQLConfig()
	if err == nil {
		t.Fatal("loadMySQLConfig returned nil error, want missing DB_HOST error")
	}
}

func TestLoadMySQLConfigUsesExplicitHostInReleaseMode(t *testing.T) {
	t.Setenv("DB_HOST", "172.16.0.6")
	t.Setenv("GIN_MODE", "release")

	originalContainerRuntime := isContainerRuntime
	isContainerRuntime = func() bool { return false }
	t.Cleanup(func() {
		isContainerRuntime = originalContainerRuntime
	})

	cfg, err := loadMySQLConfig()
	if err != nil {
		t.Fatalf("loadMySQLConfig returned error: %v", err)
	}

	if cfg.host != "172.16.0.6" {
		t.Fatalf("host = %q, want 172.16.0.6", cfg.host)
	}
}
