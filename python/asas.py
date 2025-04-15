#!/usr/bin/env python3

import os
import ssl
import time
import logging
import uuid

import paho.mqtt.client as mqtt

logging.basicConfig(level=logging.INFO)

handler = logging.StreamHandler()
handler.setLevel(logging.INFO)
formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
handler.setFormatter(formatter)
logger = logging.getLogger(__name__)
logger.addHandler(handler)

ENV = os.environ.get("ENV", "qa-2")

OLD_MQTT_HOST = os.environ.get(
    "OLD_MQTT_HOST", "kolonidclock-vzzlrx.a03.euc1.aws.hivemq.cloud"
)
OLD_MQTT_PORT = int(os.environ.get("OLD_MQTT_PORT", 8883))
OLD_MQTT_USER = os.environ.get("OLD_MQTT_USER", "hivemq.client.1731960296977")
OLD_MQTT_PASS = os.environ.get("OLD_MQTT_PASS", "%BD<af1.6SqT0p#C9mYd")

NEW_MQTT_HOST = os.environ.get("NEW_MQTT_HOST", "koloni-qa-2.infra.locker-os.com")
NEW_MQTT_PORT = int(os.environ.get("NEW_MQTT_PORT", 8883))
NEW_MQTT_USER = os.environ.get("NEW_MQTT_USER", "mqtt_user")
NEW_MQTT_PASS = os.environ.get("NEW_MQTT_PASS", "Du9sIUp2nLn8lDUQ")

# logger.info configuration
logger.info("ENV: %s", ENV)
logger.info("MQTT configuration:")
logger.info("OLD_MQTT_HOST:%s", OLD_MQTT_HOST)
logger.info("OLD_MQTT_PORT: %s", OLD_MQTT_PORT)
logger.info("OLD_MQTT_USER: %s", OLD_MQTT_USER)
logger.info("OLD_MQTT_PASS: %s", OLD_MQTT_PASS)
logger.info("NEW_MQTT_HOST:%s", NEW_MQTT_HOST)
logger.info("NEW_MQTT_PORT: %s", NEW_MQTT_PORT)
logger.info("NEW_MQTT_USER: %s", NEW_MQTT_USER)
logger.info("NEW_MQTT_PASS: %s", NEW_MQTT_PASS)


if ENV == "production":
    OLD_SUBSCRIBE_TOPIC = "/status"
    NEW_SUBSCRIBE_TOPIC = "+/cmd"
else:
    OLD_SUBSCRIBE_TOPIC = f"{ENV}/status"
    NEW_SUBSCRIBE_TOPIC = f"{ENV}/+/cmd"

logger.info("MQTT topics:")
logger.info("OLD_SUBSCRIBE_TOPIC: %s", OLD_MQTT_HOST)
logger.info("NEW_SUBSCRIBE_TOPIC: %s", NEW_MQTT_HOST)


client_old = mqtt.Client(client_id=str(uuid.uuid4()))
client_old.username_pw_set(OLD_MQTT_USER, OLD_MQTT_PASS)
client_old.tls_set(tls_version=ssl.PROTOCOL_TLSv1_2)
client_old.tls_insecure_set(True)

client_new = mqtt.Client(client_id=str(uuid.uuid4()))
client_new.username_pw_set(NEW_MQTT_USER, NEW_MQTT_PASS)
client_new.tls_set(tls_version=ssl.PROTOCOL_TLSv1_2)
client_new.tls_insecure_set(True)


def on_message_old(client, userdata, msg):
    logger.info("[OLD->NEW] topic=%s => forwarding to NEW", msg.topic)
    client_new.publish(msg.topic, payload=msg.payload)


def on_message_new(client, userdata, msg):
    logger.info("[NEW->OLD] topic=%s => forwarding to OLD", msg.topic)
    client_old.publish(msg.topic, payload=msg.payload)


def on_connect_old(client, userdata, flags, rc):
    if rc == 0:
        logger.info("[OLD] Connected OK, subscribing to: %s", OLD_SUBSCRIBE_TOPIC)
        client.subscribe(OLD_SUBSCRIBE_TOPIC)
    else:
        logger.info("[OLD] Connection error: %s", rc)


def on_connect_new(client, userdata, flags, rc):
    if rc == 0:
        logger.info("[NEW] Connected OK, subscribing to: %s", NEW_SUBSCRIBE_TOPIC)
        client.subscribe(NEW_SUBSCRIBE_TOPIC)
    else:
        logger.info("[NEW] Connection error: %s", rc)


def main():
    client_old.on_connect = on_connect_old
    client_old.on_message = on_message_old

    client_new.on_connect = on_connect_new
    client_new.on_message = on_message_new

    logger.info("Connecting OLD => %s:%s", OLD_MQTT_HOST, OLD_MQTT_PORT)
    client_old.connect(OLD_MQTT_HOST, OLD_MQTT_PORT, keepalive=60)

    logger.info("Connecting NEW => %s:%s", NEW_MQTT_HOST, NEW_MQTT_PORT)
    client_new.connect(NEW_MQTT_HOST, NEW_MQTT_PORT, keepalive=60)

    client_old.loop_start()
    client_new.loop_start()

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        logger.info("Stopping MQTT bridge...")

    client_old.loop_stop()
    client_new.loop_stop()

    client_old.disconnect()
    client_new.disconnect()


if __name__ == "__main__":
    main()
