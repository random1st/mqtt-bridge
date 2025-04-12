package main

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/random1st/mqtt-bridge/internal/bridge"
	"github.com/random1st/mqtt-bridge/internal/config"
	"github.com/random1st/mqtt-bridge/internal/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(strings.ToLower(cfg.LoggingLevel))
	defer logger.Sync()

	logger.L().Info("Starting MQTT bridge",
		zap.String("remote_host", cfg.RemoteBroker.Host),
		zap.String("local_host", cfg.LocalBroker.Host),
		zap.String("log_level", cfg.LoggingLevel),
	)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":2112", nil)
		if err != nil {
			log.Printf("Prometheus metrics server failed: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	if err := bridge.RunBridges(ctx, cfg, bridge.CreateMQTTClient); err != nil {
		log.Printf("Warning: RunBridges returned error: %v", err)
	}
	logger.L().Info("MQTT bridge started. Press Ctrl+C to stop...")

	if err := runBridgeTest(cfg); err != nil {
		logger.L().Fatal("Health check failed", zap.Error(err))
	} else {
		logger.L().Info("Health check  test passed")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.L().Info("Received stop signal, shutting down gracefully...")
	cancel()
	logger.L().Info("MQTT bridge stopped. Goodbye.")
}

func runBridgeTest(cfg *config.BridgeConfig) error {
	logger.L().Info("Starting bridge integration test...")

	localClient := bridge.CreateMQTTClient(cfg.LocalBroker, "test-local")
	remoteClient := bridge.CreateMQTTClient(cfg.RemoteBroker, "test-remote")

	if err := bridge.ConnectMQTT(localClient, cfg.LocalBroker, "test-local"); err != nil {
		return fmt.Errorf("failed to connect local broker: %w", err)
	}
	defer bridge.DisconnectMQTT(localClient, "test-local")

	if err := bridge.ConnectMQTT(remoteClient, cfg.RemoteBroker, "test-remote"); err != nil {
		return fmt.Errorf("failed to connect remote broker: %w", err)
	}
	defer bridge.DisconnectMQTT(remoteClient, "test-remote")

	if err := testMessageFlow(remoteClient, localClient, "dev/status", "Hello from Remote -> Local"); err != nil {
		return fmt.Errorf("test Remote->Local failed: %w", err)
	}
	logger.L().Info("Test Remote->Local passed")

	if err := testMessageFlow(localClient, remoteClient, "dev/test/cmd", "Hello from Local -> Remote"); err != nil {
		return fmt.Errorf("test Local->Remote failed: %w", err)
	}
	logger.L().Info("Test Local->Remote passed")

	logger.L().Info("All integration tests passed!")
	return nil
}

func testMessageFlow(pubClient, subClient mqtt.Client, topic, payload string) error {
	resultChan := make(chan string, 1)

	token := subClient.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		resultChan <- string(msg.Payload())
	})
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("subscribe error: %v", token.Error())
	}
	defer subClient.Unsubscribe(topic)

	pubClient.Publish(topic, 0, false, payload)

	select {
	case res := <-resultChan:
		if res == payload {
			return nil
		}
		return fmt.Errorf("expected payload '%s', got '%s'", payload, res)
	case <-time.After(5 * time.Second):
		return fmt.Errorf("did not receive message on topic '%s' within 5s", topic)
	}
}
