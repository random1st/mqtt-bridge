package config

import (
	"log"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
)

var (
	cfg      BridgeConfig
	once     sync.Once
	koanfCfg = koanf.New(".")
)

type BrokerConfig struct {
	Host string `koanf:"host"`
	Port string `koanf:"port"`
	User string `koanf:"user"`
	Pass string `koanf:"pass"`
	TLS  bool   `koanf:"tls"`
}

type BridgeConfig struct {
	RemoteBroker        BrokerConfig `koanf:"remote_broker"`
	LocalBroker         BrokerConfig `koanf:"local_broker"`
	LoggingLevel        string       `koanf:"logging_level"`
	IncomingPatternsRaw string       `koanf:"incoming_patterns"`
	OutgoingPatternsRaw string       `koanf:"outgoing_patterns"`
	IncomingPatterns    []string     `koanf:"-"`
	OutgoingPatterns    []string     `koanf:"-"`
}

var HealthCheckIncoming = "__healthcheck__incoming__"
var HealthCheckOutgoing = "__healthcheck__outgoing__"

func LoadConfig() *BridgeConfig {
	once.Do(func() {
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("No .env file found or error reading it: %v", err)
		}
		if err := koanfCfg.Load(env.Provider("", ".", func(s string) string {
			s = strings.ToLower(s)
			return strings.ReplaceAll(s, "__", ".")
		}), nil); err != nil {
			log.Fatalf("Error loading environment variables: %v", err)
		}
		if err := koanfCfg.Unmarshal("", &cfg); err != nil {
			log.Fatalf("Error unmarshalling config into BridgeConfig: %v", err)
		}
		if cfg.IncomingPatternsRaw != "" {
			cfg.IncomingPatterns = strings.Split(cfg.IncomingPatternsRaw, ",")
		}
		if cfg.OutgoingPatternsRaw != "" {
			cfg.OutgoingPatterns = strings.Split(cfg.OutgoingPatternsRaw, ",")
		}
		cfg.IncomingPatterns = append(cfg.IncomingPatterns, HealthCheckIncoming)
		cfg.OutgoingPatterns = append(cfg.OutgoingPatterns, HealthCheckOutgoing)
		log.Printf("Loaded config: %+v", cfg)
	})
	return &cfg
}
