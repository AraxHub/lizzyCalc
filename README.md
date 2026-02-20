# lizzyCalc

## Kafka в lizzyCalc: топики, партиции, Kafka UI

### 1. Зачем Kafka в приложении

После каждой успешной операции калькулятор (use case) **публикует событие** в Kafka: «выполнена операция с такими-то числами и результатом». Другие сервисы или тот же lizzyCalc могут **подписаться** на этот поток и реагировать: логировать, писать в аналитику, дублировать в другой хранилище и т.д.

У нас один консьюмер — само приложение: оно читает топик и вызывает `HandleOperationEvent` (сейчас только логирует). Архитектурно продюсер и консьюмер разделены: use case не знает про Kafka, он отдаёт сообщения через порт **IProducer**; консьюмер в инфраструктуре получает сообщения и дергает **ICalculatorUseCase.HandleOperationEvent**.

---

### 2. Стек: Zookeeper + Kafka + Kafka UI

В `deployment/localCalc/docker-compose.yml` поднимаются:

| Сервис      | Образ                          | Порт на хосте | Назначение |
|------------|---------------------------------|---------------|------------|
| **Zookeeper** | confluentinc/cp-zookeeper:7.6.0 | 2181          | Координация кластера Kafka (метаданные, лидер брокера). |
| **Kafka**     | confluentinc/cp-kafka:7.6.0      | 9092, 29092   | Брокер сообщений. 9092 — с хоста, 29092 — изнутри Docker-сети. |
| **Kafka UI**  | provectuslabs/kafka-ui:latest    | 8090          | Веб-интерфейс для просмотра топиков, сообщений, consumer groups. |

- **Почему два порта у Kafka.** Внутри Docker контейнеры общаются по имени `kafka` и порту **29092** (listener `PLAINTEXT`). С хоста (IDE, grpcurl, Kafka UI с хоста) подключаются к **localhost:9092** (listener `PLAINTEXT_HOST`). Иначе с хоста по `kafka:29092` не подключиться — такого хоста на машине нет.
- В **compose** сервис calculator получает `CALCULATOR_KAFKA_BROKERS=kafka:29092` (подключение к контейнеру kafka из той же сети). Если запускаешь приложение **на хосте** (не в контейнере), в `.env` укажи `CALCULATOR_KAFKA_BROKERS=localhost:9092`.

---

### 3. Топики и партиции

**Топик** — именованный поток сообщений. У нас один топик: **`operations`** (задаётся в конфиге `CALCULATOR_KAFKA_TOPIC=operations`).

**Партиции** — топик физически разбит на партиции (партиция = упорядоченный лог на одном брокере). Сообщения с одним ключом попадают в одну партицию (порядок сохраняется по ключу); без ключа или с разными ключами — распределяются по партициям. У нас в коде продюсер может слать с ключом или без; при автосоздании топика Kafka создаёт его с одной партицией по умолчанию.

- **Replication factor** в нашем compose = 1 (один брокер). В проде обычно 2–3 для отказоустойчивости.
- **Создание топика:** включено автосоздание (`KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"`): при первой записи в `operations` топик создаётся сам. Либо вручную после подъёма стека:
  ```bash
  make kafka-create-topic
  ```
  (создаёт топик `operations` с 3 партициями и replication-factor 1; команда в Makefile вызывает `kafka-topics` внутри контейнера `lizzycalc-kafka`.)

---

### 4. Формат сообщений

В топик пишется **JSON** — сериализованная структура **domain.Operation** (number1, number2, operation, result, message, timestamp). Консьюмер читает сообщение, делает `json.Unmarshal` в `domain.Operation` и вызывает `uc.HandleOperationEvent(ctx, op)`. Невалидный JSON логируется (warn), сообщение коммитится (skip), чтобы не блокировать очередь.

---

### 5. Consumer group

Консьюмер подписан на топик в рамках **consumer group** (`CALCULATOR_KAFKA_GROUP_ID=lizzycalc-app`). Группа позволяет:

- распределять партиции между инстансами приложения (каждая партиция читается только одним консьюмером в группе);
- хранить offset по группе (до какого места в каждой партиции дочитали).

У нас один инстанс приложения — он читает все партиции топика `operations`. При падении и перезапуске чтение продолжится с последнего закоммиченного offset.

---

### 6. Переменные окружения (Kafka)

Конфиг читается из env с префиксом `CALCULATOR_KAFKA_` (см. `internal/app/config.go` и `internal/infrastructure/kafka`):

| Переменная | Пример | Описание |
|------------|--------|----------|
| `CALCULATOR_KAFKA_BROKERS` | `kafka:29092` | Брокеры через запятую. В Docker — `kafka:29092`. |
| `CALCULATOR_KAFKA_TOPIC`   | `operations`  | Имя топика для записи и чтения. |
| `CALCULATOR_KAFKA_GROUP_ID`| `lizzycalc-app` | Идентификатор consumer group. |

Если `CALCULATOR_KAFKA_BROKERS` пустой, продюсер не создаётся и консьюмер не запускается (приложение работает без Kafka).

---

### 7. Kafka UI

После `make backend-up` или `make backend-from-zero` открой в браузере:

**http://localhost:8090**

- **Clusters** — кластер `lizzycalc` (подключён к Zookeeper и брокеру внутри Docker).
- **Topics** — список топиков; выбери `operations`, посмотри партиции, количество сообщений, offset’ы.
- **Messages** — просмотр сообщений в топике (по партиции, offset), тело в JSON.
- **Consumers** — consumer groups; группа `lizzycalc-app`, лаг (lag) по партициям.

Удобно проверять, что после вызова REST Calculate в топике появилось новое сообщение и консьюмер его обработал (лаг не растёт).

---

### 8. Краткая схема потока Kafka

```
  REST Calculate
          |
          v
  usecase/calculator (Calculate)
          | после успешного расчёта
          v
  IProducer.Send(ctx, key, value)  ——>  Kafka topic "operations"
          |                                    |
  (реализация: kafka.Producer)                 v
                                        Consumer (kafka.Reader)
                                        json.Unmarshal -> domain.Operation
                                                |
                                                v
                                        uc.HandleOperationEvent(ctx, op)
                                        (логирование / будущая обработка)
```
