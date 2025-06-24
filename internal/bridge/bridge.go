package bridge

import (
	"context"
	"fmt"
	"github.com/random1st/mqtt-bridge/internal/metrics"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/random1st/mqtt-bridge/internal/config"
	"github.com/random1st/mqtt-bridge/internal/logger"
	"go.uber.org/zap"
)

var (
	mqttBridgeMessages = metrics.GetBridgeMessagesCounter()
)

type Pair struct {
	Name         string
	FromClient   mqtt.Client
	ToClient     mqtt.Client
	FromConfig   config.BrokerConfig
	ToConfig     config.BrokerConfig
	TopicPattern string
}

func startBridge(ctx context.Context, bp Pair) error {
	if err := ConnectMQTT(bp.FromClient, bp.FromConfig, bp.Name+"-FROM"); err != nil {
		logger.L().Error("Failed to connect FROM broker", zap.String("bridge", bp.Name), zap.Error(err))
		return err
	}
	logger.L().Info("Connected FROM broker", zap.String("bridge", bp.Name), zap.String("fromHost", bp.FromConfig.Host))

	if err := ConnectMQTT(bp.ToClient, bp.ToConfig, bp.Name+"-TO"); err != nil {
		logger.L().Error("Failed to connect TO broker", zap.String("bridge", bp.Name), zap.Error(err))
		DisconnectMQTT(bp.FromClient, bp.Name+"-FROM")
		return err
	}
	logger.L().Info("Connected TO broker", zap.String("bridge", bp.Name), zap.String("toHost", bp.ToConfig.Host))

	token := bp.FromClient.Subscribe(bp.TopicPattern, 0, func(client mqtt.Client, msg mqtt.Message) {
		logger.L().Debug("Received message, forwarding",
			zap.String("bridge", bp.Name),
			zap.String("topic", msg.Topic()),
			zap.Int("payloadLength", len(msg.Payload())),
		)
		bp.ToClient.Publish(msg.Topic(), 0, false, msg.Payload())
		mqttBridgeMessages.WithLabelValues(bp.Name, bp.TopicPattern).Inc()

	})
	token.Wait()
	if token.Error() != nil {
		logger.L().Error("Subscription error", zap.String("bridge", bp.Name), zap.Error(token.Error()))
		DisconnectMQTT(bp.ToClient, bp.Name+"-TO")
		DisconnectMQTT(bp.FromClient, bp.Name+"-FROM")
		return token.Error()
	}
	logger.L().Info("Subscribed to topic pattern", zap.String("bridge", bp.Name), zap.String("topicPattern", bp.TopicPattern))

	go func() {
		<-ctx.Done()
		logger.L().Info("Context canceled", zap.String("bridge", bp.Name))
		bp.FromClient.Unsubscribe(bp.TopicPattern)
		DisconnectMQTT(bp.ToClient, bp.Name+"-TO")
		DisconnectMQTT(bp.FromClient, bp.Name+"-FROM")
	}()

	return nil
}

func RunBridges(ctx context.Context, cfg *config.BridgeConfig, createClient func(cfg config.BrokerConfig, prefix string, onConnect mqtt.OnConnectHandler) mqtt.Client) error {
	var wg sync.WaitGroup

	for i, topic := range cfg.IncomingPatterns {
		logger.L().Info("Starting incoming bridge", zap.String("topicPattern", topic))
		wg.Add(1)
		go func(index int, topic string) {
			defer wg.Done()

			bridgeName := fmt.Sprintf("Incoming-%d", index)

			var toClient = createClient(cfg.LocalBroker, "local", nil)

			fromClient := createClient(cfg.RemoteBroker, "remote", func(c mqtt.Client) {
				logger.L().Info("Reconnected. Re-subscribing", zap.String("bridge", bridgeName), zap.String("topic", topic))
				token := c.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
					logger.L().Debug("Received message, forwarding",
						zap.String("bridge", bridgeName),
						zap.String("topic", msg.Topic()),
						zap.Int("payloadLength", len(msg.Payload())),
					)
					toClient.Publish(msg.Topic(), 0, false, msg.Payload())
					mqttBridgeMessages.WithLabelValues(bridgeName, topic).Inc()
				})
				token.Wait()
				if token.Error() != nil {
					logger.L().Error("Re-subscribe failed", zap.String("bridge", bridgeName), zap.Error(token.Error()))
				}
			})

			pair := Pair{
				Name:         bridgeName,
				FromClient:   fromClient,
				ToClient:     toClient,
				FromConfig:   cfg.RemoteBroker,
				ToConfig:     cfg.LocalBroker,
				TopicPattern: topic,
			}

			if err := startBridge(ctx, pair); err != nil {
				logger.L().Warn("Error starting incoming bridge", zap.Int("index", index), zap.Error(err))
			}
		}(i, topic)
	}

	for i, pattern := range cfg.OutgoingPatterns {
		logger.L().Info("Starting outgoing bridge", zap.String("topicPattern", pattern))
		wg.Add(1)

		go func(index int, topic string) {
			defer wg.Done()

			bridgeName := fmt.Sprintf("Outgoing-%d", index)

			toClient := createClient(cfg.RemoteBroker, "remote", nil)

			fromClient := createClient(cfg.LocalBroker, "local", func(c mqtt.Client) {
				logger.L().Info("Reconnected. Re-subscribing",
					zap.String("bridge", bridgeName),
					zap.String("topic", topic),
				)

				token := c.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
					logger.L().Debug("Received message, forwarding",
						zap.String("bridge", bridgeName),
						zap.String("topic", msg.Topic()),
						zap.Int("payloadLength", len(msg.Payload())),
					)
					toClient.Publish(msg.Topic(), 0, false, msg.Payload())
					mqttBridgeMessages.WithLabelValues(bridgeName, topic).Inc()
				})

				token.Wait()
				if token.Error() != nil {
					logger.L().Error("Re-subscribe failed",
						zap.String("bridge", bridgeName),
						zap.Error(token.Error()))
				}
			})

			pair := Pair{
				Name:         bridgeName,
				FromClient:   fromClient,
				ToClient:     toClient,
				FromConfig:   cfg.LocalBroker,
				ToConfig:     cfg.RemoteBroker,
				TopicPattern: topic,
			}

			err := startBridge(ctx, pair)
			if err != nil {
				logger.L().Warn("Error starting outgoing bridge", zap.Int("index", index), zap.Error(err))
			}
		}(i, pattern)
	}

	go func() {
		<-ctx.Done()
		wg.Wait()
	}()

	return nil
}
