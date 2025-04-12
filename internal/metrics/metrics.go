package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var mqttBridgeMessages *prometheus.CounterVec
var onceMetrics sync.Once

func GetBridgeMessagesCounter() *prometheus.CounterVec {
	onceMetrics.Do(func() {
		mqttBridgeMessages = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mqtt_bridge_messages_total",
				Help: "Total number of MQTT messages bridged",
			},
			[]string{"bridgeName", "topicPattern"},
		)
		prometheus.MustRegister(mqttBridgeMessages)
	})
	return mqttBridgeMessages
}
