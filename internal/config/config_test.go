package config

import (
	"reflect"
	"sync"
	"testing"

	"github.com/knadh/koanf"
)

func resetGlobalConfig() {
	once = sync.Once{}
	cfg = BridgeConfig{}
	koanfCfg = koanf.New(".")
}

func TestLoadConfigBasic(t *testing.T) {
	resetGlobalConfig()

	t.Setenv("REMOTE_BROKER__HOST", "remote.example.com")
	t.Setenv("REMOTE_BROKER__PORT", "8883")
	t.Setenv("REMOTE_BROKER__USER", "remUser")
	t.Setenv("REMOTE_BROKER__PASS", "remPass")
	t.Setenv("REMOTE_BROKER__TLS", "true")
	t.Setenv("LOCAL_BROKER__HOST", "local.example.com")
	t.Setenv("LOCAL_BROKER__PORT", "1883")
	t.Setenv("LOCAL_BROKER__USER", "locUser")
	t.Setenv("LOCAL_BROKER__PASS", "locPass")
	t.Setenv("LOCAL_BROKER__TLS", "false")
	t.Setenv("LOGGING_LEVEL", "debug")
	t.Setenv("INCOMING_PATTERNS", "dev/status,dev/foo")
	t.Setenv("OUTGOING_PATTERNS", "dev/+/cmd")

	config := LoadConfig()

	if config.RemoteBroker.Host != "remote.example.com" {
		t.Errorf("RemoteBroker.Host = %q, want %q", config.RemoteBroker.Host, "remote.example.com")
	}
	if config.RemoteBroker.Port != "8883" {
		t.Errorf("RemoteBroker.Port = %q, want %q", config.RemoteBroker.Port, "8883")
	}
	if config.RemoteBroker.User != "remUser" {
		t.Errorf("RemoteBroker.User = %q, want %q", config.RemoteBroker.User, "remUser")
	}
	if config.RemoteBroker.Pass != "remPass" {
		t.Errorf("RemoteBroker.Pass = %q, want %q", config.RemoteBroker.Pass, "remPass")
	}
	if config.RemoteBroker.TLS != true {
		t.Errorf("RemoteBroker.TLS = %v, want true", config.RemoteBroker.TLS)
	}
	if config.LocalBroker.Host != "local.example.com" {
		t.Errorf("LocalBroker.Host = %q, want %q", config.LocalBroker.Host, "local.example.com")
	}
	if config.LocalBroker.Port != "1883" {
		t.Errorf("LocalBroker.Port = %q, want %q", config.LocalBroker.Port, "1883")
	}
	if config.LocalBroker.User != "locUser" {
		t.Errorf("LocalBroker.User = %q, want %q", config.LocalBroker.User, "locUser")
	}
	if config.LocalBroker.Pass != "locPass" {
		t.Errorf("LocalBroker.Pass = %q, want %q", config.LocalBroker.Pass, "locPass")
	}
	if config.LocalBroker.TLS != false {
		t.Errorf("LocalBroker.TLS = %v, want false", config.LocalBroker.TLS)
	}
	if config.LoggingLevel != "debug" {
		t.Errorf("LoggingLevel = %q, want %q", config.LoggingLevel, "debug")
	}

	wantIncoming := []string{"dev/status", "dev/foo", HealthCheckIncoming}
	if !reflect.DeepEqual(config.IncomingPatterns, wantIncoming) {
		t.Errorf("IncomingPatterns = %+v, want %+v", config.IncomingPatterns, wantIncoming)
	}
	wantOutgoing := []string{"dev/+/cmd", HealthCheckOutgoing}
	if !reflect.DeepEqual(config.OutgoingPatterns, wantOutgoing) {
		t.Errorf("OutgoingPatterns = %+v, want %+v", config.OutgoingPatterns, wantOutgoing)
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	resetGlobalConfig()

	config := LoadConfig()

	if config.RemoteBroker.Host != "" {
		t.Errorf("RemoteBroker.Host = %q, want empty", config.RemoteBroker.Host)
	}
	if len(config.IncomingPatterns) != 1 || config.IncomingPatterns[0] != HealthCheckIncoming {
		t.Errorf("IncomingPatterns = %+v, want just [%q]", config.IncomingPatterns, HealthCheckIncoming)
	}
	if len(config.OutgoingPatterns) != 1 || config.OutgoingPatterns[0] != HealthCheckOutgoing {
		t.Errorf("OutgoingPatterns = %+v, want just [%q]", config.OutgoingPatterns, HealthCheckOutgoing)
	}
}
