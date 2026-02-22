# LizzyCalc

Калькулятор с полным мониторинг-стеком: Prometheus + Grafana.

---

## Мониторинг: полный ликбез

### Что такое метрики и зачем они нужны

**Метрики** — это числовые измерения состояния системы в определённый момент времени. В отличие от логов (текстовые записи событий) и трейсов (путь запроса через систему), метрики — это просто числа: сколько запросов пришло, как долго они обрабатывались, сколько памяти занято.

Зачем:
- Понять, что происходит с системой прямо сейчас
- Заметить деградацию до того, как пользователи начнут жаловаться
- Найти узкие места (bottleneck)
- Построить алерты («если latency > 500ms — разбуди дежурного»)

---

### Prometheus: что это и как работает

**Prometheus** — это time-series database (TSDB), заточенная под метрики. Главная особенность: **pull-модель**.

#### Pull vs Push

```
PUSH-модель (StatsD, Graphite):
  Приложение ──отправляет метрики──▸ Сервер метрик

PULL-модель (Prometheus):
  Prometheus ──запрашивает /metrics──▸ Приложение
```

Pull лучше тем, что:
- Prometheus сам контролирует частоту опроса
- Если приложение упало — Prometheus это сразу увидит (нет ответа)
- Не нужно настраивать firewall для исходящих соединений из приложения

#### Как Prometheus собирает метрики

1. В конфиге (`prometheus.yml`) указаны **targets** — адреса приложений
2. Каждые `scrape_interval` секунд (у нас 15s) Prometheus делает HTTP GET на `/metrics`
3. Приложение отвечает текстом в специальном формате
4. Prometheus парсит и сохраняет в свою базу

Наш конфиг:

```yaml
global:
  scrape_interval: 15s      # как часто опрашивать targets
  evaluation_interval: 15s  # как часто вычислять recording/alerting rules

scrape_configs:
  - job_name: 'calculator'
    static_configs:
      - targets: ['calculator:8080']
    metrics_path: /metrics  # эндпоинт для сбора (по умолчанию /metrics)
```

#### Формат метрик (Prometheus exposition format)

Когда делаешь `curl http://localhost:8080/metrics`, видишь что-то такое:

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/calculate",status="200"} 1547
http_requests_total{method="POST",path="/calculate",status="200"} 892
http_requests_total{method="GET",path="/calculate",status="500"} 3

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/calculate",le="0.005"} 1200
http_request_duration_seconds_bucket{method="GET",path="/calculate",le="0.01"} 1400
http_request_duration_seconds_bucket{method="GET",path="/calculate",le="0.025"} 1520
...
http_request_duration_seconds_bucket{method="GET",path="/calculate",le="+Inf"} 1547
http_request_duration_seconds_sum{method="GET",path="/calculate"} 12.847
http_request_duration_seconds_count{method="GET",path="/calculate"} 1547
```

- `# HELP` — описание метрики
- `# TYPE` — тип (counter, gauge, histogram, summary)
- `{...}` — **labels** (теги), позволяют фильтровать и группировать
- Число в конце — значение

---

### Типы метрик в Prometheus

#### 1. Counter (счётчик)

Только растёт (или сбрасывается при рестарте). Примеры:
- Количество запросов
- Количество ошибок
- Количество обработанных сообщений

```go
httpRequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)

// Использование:
httpRequestsTotal.WithLabelValues("GET", "/calculate", "200").Inc()
```

Важно: **сырое значение counter'а бесполезно**. Counter показывает «всего с момента старта» — через час это будет 10000, через день 1000000. Поэтому всегда используют `rate()` — скорость изменения.

#### 2. Gauge (измеритель)

Может расти и уменьшаться. Примеры:
- Температура CPU
- Количество активных соединений
- Размер очереди
- Количество горутин

```go
httpRequestsInFlight = promauto.NewGauge(
    prometheus.GaugeOpts{
        Name: "http_requests_in_flight",
        Help: "Number of HTTP requests currently being processed",
    },
)

// Использование:
httpRequestsInFlight.Inc()  // запрос начался
// ... обработка ...
httpRequestsInFlight.Dec()  // запрос завершился
```

#### 3. Histogram (гистограмма)

Распределение значений по бакетам (buckets). Идеально для latency.

```go
httpRequestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: prometheus.DefBuckets, // [.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]
    },
    []string{"method", "path"},
)

// Использование:
start := time.Now()
// ... обработка ...
httpRequestDuration.WithLabelValues("GET", "/calculate").Observe(time.Since(start).Seconds())
```

Histogram создаёт несколько метрик:
- `_bucket{le="X"}` — сколько запросов уложились в X секунд
- `_sum` — сумма всех значений
- `_count` — количество наблюдений

**Почему бакеты, а не точные значения?**

Хранить каждое значение дорого. Бакеты — компромисс: мы теряем точность, но экономим место. Зная распределение по бакетам, можем вычислить перцентили.

#### 4. Summary (резюме)

Похож на histogram, но вычисляет квантили на стороне приложения. Минус: нельзя агрегировать между инстансами. В 99% случаев лучше использовать histogram.

---

### Latency (задержка): что это и почему важно

**Latency** — время от получения запроса до отправки ответа. Измеряется в миллисекундах или секундах.

```
Клиент ──запрос──▸ [Сервер обрабатывает] ──ответ──▸ Клиент
         │◀────────── latency ──────────▶│
```

Latency складывается из:
- Парсинг запроса
- Бизнес-логика
- Запросы в БД
- Сериализация ответа

**Почему нельзя смотреть только на среднее (average)?**

Допустим, 99 запросов выполнились за 10ms, а 1 запрос — за 10 секунд.
- Среднее: (99×10 + 1×10000) / 100 = 109ms — выглядит нормально
- Реальность: каждый 100-й пользователь ждёт 10 секунд

Среднее скрывает проблемы. Поэтому используют **перцентили**.

---

### Перцентили (percentiles): p50, p95, p99

**Перцентиль X** — значение, ниже которого находится X% наблюдений.

- **p50 (медиана)** — 50% запросов быстрее этого значения, 50% медленнее
- **p95** — 95% запросов быстрее, только 5% медленнее
- **p99** — 99% запросов быстрее, только 1% медленнее

Пример:
```
p50 = 15ms   → половина запросов укладывается в 15ms
p95 = 80ms   → 95% запросов укладывается в 80ms
p99 = 500ms  → 99% запросов укладывается в 500ms, но 1% ждёт дольше
```

**Какой перцентиль смотреть?**

- **p50** — типичный пользовательский опыт
- **p95** — «почти все пользователи» (для SLA обычно берут его)
- **p99** — worst case (важно для финтеха, биржи)

На нашем дашборде показан **p95** — это золотая середина.

#### Как Prometheus считает перцентили из histogram

Функция `histogram_quantile()`:

```promql
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
```

Разберём:
1. `http_request_duration_seconds_bucket` — сырые бакеты
2. `rate(...[5m])` — скорость заполнения бакетов за последние 5 минут
3. `sum(...) by (le)` — суммируем по всем label'ам, кроме `le` (граница бакета)
4. `histogram_quantile(0.95, ...)` — интерполирует 95-й перцентиль

---

### PromQL: язык запросов Prometheus

#### Базовые операции

```promql
# Просто значение метрики
http_requests_total

# Фильтрация по label
http_requests_total{status="500"}
http_requests_total{status=~"5.."}  # regex: любой 5xx

# rate — скорость изменения counter'а (запросов в секунду)
rate(http_requests_total[1m])

# increase — абсолютный прирост за период
increase(http_requests_total[1h])

# sum — агрегация
sum(rate(http_requests_total[1m]))

# sum by — агрегация с группировкой
sum(rate(http_requests_total[1m])) by (path)
sum(rate(http_requests_total[1m])) by (status)
```

#### [1m], [5m] — что это?

Это **range vector** — окно времени для вычисления rate/increase.

- `[1m]` — rate за последнюю минуту (более чувствительный к пикам)
- `[5m]` — rate за 5 минут (более сглаженный)

Правило: range должен быть >= 4× scrape_interval. У нас scrape_interval=15s, значит минимум [1m].

---

### Grafana: визуализация метрик

**Grafana** — UI для построения дашбордов. Подключается к Prometheus как datasource и выполняет PromQL-запросы.

#### Datasource

Настроен в `grafana/provisioning/datasources/datasources.yml`:

```yaml
datasources:
  - name: Prometheus
    type: prometheus
    uid: prometheus
    url: http://prometheus:9090  # Grafana обращается к Prometheus по Docker-сети
    isDefault: true
```

#### Dashboard provisioning

Дашборды можно создавать руками в UI или **провиженить** из JSON-файлов.

`grafana/provisioning/dashboards/dashboards.yml` говорит Grafana:
```yaml
providers:
  - name: 'default'
    type: file
    options:
      path: /etc/grafana/provisioning/dashboards  # искать JSON-дашборды здесь
```

При старте Grafana читает `calculator.json` и создаёт дашборд автоматически.

---

### Наш дашборд: что показывает каждый график

#### 1. Requests/sec (Stat panel)

```promql
sum(rate(http_requests_total[1m]))
```

**Что показывает**: общий RPS (requests per second) — сколько запросов в секунду обрабатывает сервер.

**Пороги (thresholds)**:
- Зелёный: норма
- Жёлтый: > 100 RPS — высокая нагрузка
- Красный: > 500 RPS — возможно перегруз

#### 2. P95 Latency (Stat panel)

```promql
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
```

**Что показывает**: 95-й перцентиль latency. «95% запросов выполняются быстрее этого значения».

**Пороги**:
- Зелёный: норма
- Жёлтый: > 100ms — подтормаживает
- Красный: > 500ms — проблема

#### 3. In-Flight Requests (Stat panel)

```promql
http_requests_in_flight
```

**Что показывает**: сколько запросов прямо сейчас в обработке (начались, но ещё не завершились).

Если это число растёт и не падает — запросы копятся, сервер не справляется.

#### 4. Requests by Endpoint (Time series)

```promql
sum(rate(http_requests_total[1m])) by (path)
```

**Что показывает**: RPS в разбивке по эндпоинтам. Видно, какие ручки нагружены больше всего.

#### 5. Latency Distribution (Time series)

```promql
histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))
```

**Что показывает**: p50, p95, p99 latency по каждому эндпоинту. Видно:
- Типичное время ответа (p50)
- Время для большинства пользователей (p95)
- Worst case (p99)

Если p99 сильно отличается от p95 — есть редкие, но очень медленные запросы.

#### 6. Response Codes (Time series, stacked bars)

```promql
sum(increase(http_requests_total[1m])) by (status)
```

**Что показывает**: количество ответов по HTTP-кодам (200, 400, 500...).

**Цвета**:
- Зелёный: 2xx (успех)
- Красный: 4xx, 5xx (ошибки)

Резкий рост красного — что-то сломалось.

#### 7. Error Rate (Time series)

```promql
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
sum(rate(http_requests_total{status=~"4.."}[5m])) / sum(rate(http_requests_total[5m]))
```

**Что показывает**: процент ошибок (5xx и 4xx отдельно).

- 5xx — ошибки сервера (баги, таймауты БД)
- 4xx — ошибки клиента (невалидные запросы)

Норма: error rate < 1%. Если 5xx > 5% — инцидент.

---

### Как добавить/изменить дашборд руками

1. Открой Grafana: http://localhost:3001 (admin/admin)
2. Dashboards → LizzyCalc HTTP Metrics
3. Жми Edit на любой панели
4. В поле Query пиши PromQL
5. Save dashboard

Или экспортируй JSON (Dashboard settings → JSON Model) и положи в `grafana/provisioning/dashboards/`.

---

### Задержка при сборе метрик

**scrape_interval = 15s** означает:
- Prometheus опрашивает `/metrics` каждые 15 секунд
- Между scrape'ами данные не обновляются
- На графике точки появляются с шагом 15 секунд

**Что это значит на практике**:
- Если сервер упал и поднялся за 10 секунд — Prometheus может это не заметить
- Пики нагрузки короче 15 секунд могут быть сглажены
- `rate()` за [1m] использует ~4 точки (60s / 15s)

**Уменьшить scrape_interval?**
- 5s — более детальная картина, но больше нагрузка на Prometheus и приложение
- 1s — для очень критичных систем, требует много ресурсов
- 15s — золотая середина для большинства случаев

---

### Как работает сбор метрик в нашем приложении

Middleware `PrometheusMetrics` (`internal/api/http/middlewares/metrics.go`):

```go
func PrometheusMetrics(c *gin.Context) {
    // Не считаем сам /metrics
    if c.Request.URL.Path == "/metrics" {
        c.Next()
        return
    }

    httpRequestsInFlight.Inc()    // +1 активный запрос
    start := time.Now()

    c.Next()                       // выполняем handler

    duration := time.Since(start).Seconds()
    status := strconv.Itoa(c.Writer.Status())
    path := c.FullPath()

    httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
    httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
    httpRequestsInFlight.Dec()    // -1 активный запрос
}
```

Порядок:
1. Запрос приходит → `Inc()` на in_flight
2. Handler обрабатывает
3. Запрос завершён → записываем counter и histogram, `Dec()` на in_flight

Эндпоинт `/metrics` (system controller) отдаёт все накопленные метрики через `promhttp.Handler()`.

---

### Полезные PromQL-запросы

```promql
# Топ-5 самых медленных эндпоинтов
topk(5, histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path)))

# Эндпоинты с ошибками
sum(rate(http_requests_total{status=~"5.."}[5m])) by (path) > 0

# Общее количество запросов за последний час
sum(increase(http_requests_total[1h]))

# Средний RPS за день
avg_over_time(sum(rate(http_requests_total[5m]))[1d:5m])
```

---

### Доступы

| Сервис | URL | Логин |
|--------|-----|-------|
| Grafana | http://localhost:3001 | admin / admin |
| Prometheus | http://localhost:9091 | — |
| Метрики приложения | http://localhost:8080/metrics | — |

---

### Запуск

```bash
cd deployment/localCalc
docker-compose up -d
```

После старта Grafana автоматически загрузит дашборд **LizzyCalc HTTP Metrics**.
