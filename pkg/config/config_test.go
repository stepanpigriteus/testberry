package config

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMustAtoi(t *testing.T) {
	if err := os.Setenv("TEST_INT", "42"); err != nil {
		t.Fatalf("failed to set TEST_INT: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_INT"); err != nil {
			t.Fatalf("failed to unset TEST_INT: %v", err)
		}
	}()

	got := mustAtoi("TEST_INT", 5)
	if got != 42 {
		t.Errorf("Expected 42, got %d", got)
	}

	if err := os.Unsetenv("TEST_INT"); err != nil {
		t.Fatalf("failed to unset TEST_INT: %v", err)
	}
	got = mustAtoi("TEST_INT", 5)
	if got != 5 {
		t.Errorf("Expected default 5, got %d", got)
	}

	if err := os.Setenv("TEST_INT", "not-an-int"); err != nil {
		t.Fatalf("failed to set TEST_INT: %v", err)
	}
	got = mustAtoi("TEST_INT", 7)
	if got != 7 {
		t.Errorf("Expected fallback 7, got %d", got)
	}
}

func TestMustParseBool(t *testing.T) {
	if err := os.Setenv("TEST_BOOL", "true"); err != nil {
		t.Fatalf("failed to set TEST_BOOL: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_BOOL"); err != nil {
			t.Fatalf("failed to unset TEST_BOOL: %v", err)
		}
	}()

	if !mustParseBool("TEST_BOOL", false) {
		t.Errorf("Expected true")
	}

	if err := os.Setenv("TEST_BOOL", "not-a-bool"); err != nil {
		t.Fatalf("failed to set TEST_BOOL: %v", err)
	}
	if mustParseBool("TEST_BOOL", true) != true {
		t.Errorf("Expected fallback true")
	}

	if err := os.Unsetenv("TEST_BOOL"); err != nil {
		t.Fatalf("failed to unset TEST_BOOL: %v", err)
	}
	if mustParseBool("TEST_BOOL", false) {
		t.Errorf("Expected fallback false")
	}
}

func TestMustParseDuration(t *testing.T) {
	if err := os.Setenv("TEST_DURATION", "2s"); err != nil {
		t.Fatalf("failed to set TEST_DURATION: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_DURATION"); err != nil {
			t.Fatalf("failed to unset TEST_DURATION: %v", err)
		}
	}()

	got := mustParseDuration("TEST_DURATION", 1*time.Second)
	if got != 2*time.Second {
		t.Errorf("Expected 2s, got %v", got)
	}

	if err := os.Setenv("TEST_DURATION", "not-a-duration"); err != nil {
		t.Fatalf("failed to set TEST_DURATION: %v", err)
	}
	got = mustParseDuration("TEST_DURATION", 3*time.Second)
	if got != 3*time.Second {
		t.Errorf("Expected fallback 3s, got %v", got)
	}

	if err := os.Unsetenv("TEST_DURATION"); err != nil {
		t.Fatalf("failed to unset TEST_DURATION: %v", err)
	}
	got = mustParseDuration("TEST_DURATION", 5*time.Second)
	if got != 5*time.Second {
		t.Errorf("Expected fallback 5s, got %v", got)
	}
}

func TestMustParseStringSlice(t *testing.T) {
	if err := os.Setenv("TEST_SLICE", "one, two ,three"); err != nil {
		t.Fatalf("failed to set TEST_SLICE: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_SLICE"); err != nil {
			t.Fatalf("failed to unset TEST_SLICE: %v", err)
		}
	}()

	expected := []string{"one", "two", "three"}
	got := mustParseStringSlice("TEST_SLICE", []string{"default"})
	if len(got) != len(expected) {
		t.Fatalf("Expected slice of length %d, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], got[i])
		}
	}

	if err := os.Setenv("TEST_SLICE", ""); err != nil {
		t.Fatalf("failed to set TEST_SLICE: %v", err)
	}
	got = mustParseStringSlice("TEST_SLICE", []string{"default"})
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("Expected default slice")
	}
}

func TestGetEnv(t *testing.T) {
	if err := os.Setenv("TEST_ENV", "value"); err != nil {
		t.Fatalf("failed to set TEST_ENV: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_ENV"); err != nil {
			t.Fatalf("failed to unset TEST_ENV: %v", err)
		}
	}()

	got := getEnv("TEST_ENV")
	if got != "value" {
		t.Errorf("Expected value, got %s", got)
	}

	if err := os.Unsetenv("TEST_ENV"); err != nil {
		t.Fatalf("failed to unset TEST_ENV: %v", err)
	}
	got = getEnv("TEST_ENV")
	if got != "" {
		t.Errorf("Expected empty string, got %s", got)
	}
}

func TestGetEnvWithDefault(t *testing.T) {
	if err := os.Setenv("TEST_ENV_DEF", "realvalue"); err != nil {
		t.Fatalf("failed to set TEST_ENV_DEF: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_ENV_DEF"); err != nil {
			t.Fatalf("failed to unset TEST_ENV_DEF: %v", err)
		}
	}()

	got := getEnvWithDefault("TEST_ENV_DEF", "default")
	if got != "realvalue" {
		t.Errorf("Expected realvalue, got %s", got)
	}

	if err := os.Unsetenv("TEST_ENV_DEF"); err != nil {
		t.Fatalf("failed to unset TEST_ENV_DEF: %v", err)
	}
	got = getEnvWithDefault("TEST_ENV_DEF", "default")
	if got != "default" {
		t.Errorf("Expected default, got %s", got)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	os.Clearenv()
	cfg := LoadConfig()
	fmt.Println(cfg)

	if cfg.Redis.Port != 6379 {
		t.Errorf("Expected default Redis port 6379, got %d", cfg.Redis.Port)
	}
	if cfg.Kafka.Topic != "orders" {
		t.Errorf("Expected default Kafka topic 'orders', got %s", cfg.Kafka.Topic)
	}
	if cfg.Kafka.ConsumerGroup != "my-consumer-group" {
		t.Errorf("Expected default Kafka group, got %s", cfg.Kafka.ConsumerGroup)
	}
	if cfg.Redis.DialTimeout != 5*time.Second {
		t.Errorf("Expected default DialTimeout 5s, got %v", cfg.Redis.DialTimeout)
	}
}
