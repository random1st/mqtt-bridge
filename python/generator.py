#!/usr/bin/env python3

import ssl
import os
import time
import json
import uuid
import random
import paho.mqtt.client as mqtt

# Настройки MQTT из вашего сообщения
# NEW_MQTT_HOST = os.environ.get("NEW_MQTT_HOST", "koloni-dev.infra.locker-os.com")
# NEW_MQTT_PORT = int(os.environ.get("NEW_MQTT_PORT", 8883))
# NEW_MQTT_USER = os.environ.get("NEW_MQTT_USER", "mqtt_user")
# NEW_MQTT_PASS = os.environ.get("NEW_MQTT_PASS", "Du9sIUp2nLn8lDUQ")

NEW_MQTT_HOST = os.environ.get(
    "OLD_MQTT_HOST", "kolonidclock-vzzlrx.a03.euc1.aws.hivemq.cloud"
)
NEW_MQTT_PORT = int(os.environ.get("OLD_MQTT_PORT", 8883))
NEW_MQTT_USER = os.environ.get("OLD_MQTT_USER", "hivemq.client.1731960296977")
NEW_MQTT_PASS = os.environ.get("OLD_MQTT_PASS", "%BD<af1.6SqT0p#C9mYd")
# Какой топик используем для публикации?
# (Можно заменить на любой другой, например "dev/status" или "test/messages")
MQTT_TOPIC = "qa-2/status"


def generate_random_message() -> dict:
    """
    Генерирует структуру:
    {
      "id_org": <случайный UUID>,
      "payload": [
         "<device_id>_<box_num>_<status_code>_<product_status>",
         ...
      ]
    }
    Пример: {
      "id_org": "6e6a801c-8e5e-4906-8ec3-3febeb67395d",
      "payload": [
        "c68c5a6e9760634d_17_1_1",
        "c68c5a6e9760634d_23_1_1"
      ]
    }
    """
    # Случайный UUID для id_org
    id_org = str(uuid.uuid4())

    # Случайное количество элементов в payload
    count = random.randint(1, 5)

    payload_list = []
    for _ in range(count):
        device_id = "".join(random.choices("0123456789abcdef", k=16))

        box_num = random.randint(1, 60)

        status_code = random.randint(0, 1)

        product_status = 1

        entry = f"{device_id}_{box_num}_{status_code}_{product_status}"
        payload_list.append(entry)

    return {"id_org": id_org, "payload": payload_list}


def on_connect(client, userdata, flags, rc):
    if rc == 0:
        print("[INFO] Connected to MQTT broker successfully.")
    else:
        print(f"[ERROR] Connection failed with code {rc}")


def main():
    client = mqtt.Client(client_id="random_generator")
    client.username_pw_set(NEW_MQTT_USER, NEW_MQTT_PASS)

    client.tls_set(tls_version=ssl.PROTOCOL_TLSv1_2)
    client.tls_insecure_set(True)

    client.on_connect = on_connect

    # Подключаемся
    print(f"Connecting to MQTT broker {NEW_MQTT_HOST}:{NEW_MQTT_PORT}")
    client.connect(NEW_MQTT_HOST, NEW_MQTT_PORT, keepalive=60)

    # Запускаем loop в фоне
    client.loop_start()

    try:
        while True:
            # Генерируем сообщение
            msg_dict = generate_random_message()
            msg_json = json.dumps(msg_dict)

            # Публикуем
            client.publish(MQTT_TOPIC, payload=msg_json)
            print(f"[PUBLISHED] topic='{MQTT_TOPIC}' => {msg_json}")

            # Ждем 1 секунду (можно увеличить/уменьшить)
            time.sleep(1)
    except KeyboardInterrupt:
        print("Stopping random message generator...")

    client.loop_stop()
    client.disconnect()


if __name__ == "__main__":
    main()
