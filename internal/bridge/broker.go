package bridge

import (
	"crypto/tls"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"

	"github.com/random1st/mqtt-bridge/internal/config"
	"github.com/random1st/mqtt-bridge/internal/logger"
	"go.uber.org/zap"
)

func CreateMQTTClient(broker config.BrokerConfig, prefix string) mqtt.Client {
	scheme := "tcp"
	if broker.TLS {
		scheme = "tcps"
	}
	url := scheme + "://" + broker.Host + ":" + broker.Port

	var tlsConfig *tls.Config
	if broker.TLS {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(url)
	opts.SetClientID(prefix + "-" + uuid.New().String())
	opts.SetUsername(broker.User)
	opts.SetPassword(broker.Pass)
	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}
	opts.SetKeepAlive(60 * time.Second)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)

	return mqtt.NewClient(opts)
}

func maskPassword(pass string) string {
	if len(pass) < 1 {
		return "****"
	}
	return "********"
}

func ConnectMQTT(client mqtt.Client, broker config.BrokerConfig, name string) error {
	scheme := "tcp"
	if broker.TLS {
		scheme = "tcps"
	}
	logger.L().Info("Connecting to broker",
		zap.String("client", name),
		zap.String("scheme", scheme),
		zap.String("host", broker.Host),
		zap.String("port", broker.Port),
		zap.String("user", broker.User),
		zap.String("pass", maskPassword(broker.Pass)),
	)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		logger.L().Error("Failed to connect",
			zap.String("client", name),
			zap.Error(token.Error()),
		)
		return token.Error()
	}
	logger.L().Info("Connected OK", zap.String("client", name))
	return nil
}

func DisconnectMQTT(client mqtt.Client, name string) {
	if client.IsConnected() {
		logger.L().Info("Disconnecting...", zap.String("client", name))
		client.Disconnect(250)
		logger.L().Info("Disconnected", zap.String("client", name))
	}
}
