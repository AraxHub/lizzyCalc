# lizzyCalc

Калькулятор с REST и gRPC API. Ниже — подробная инструкция по gRPC: что это, зачем отдельный порт, как устроен репо контрактов и кодоген, как всё работает в приложении.

---

## gRPC в lizzyCalc: подробная инструкция

### 1. Что такое gRPC и зачем он тут

**gRPC** (gRPC Remote Procedure Call) — это способ вызывать методы на сервере как обычные функции: клиент вызывает `Calculate(1, 2, "+")`, по сети летит бинарный запрос, сервер считает и возвращает бинарный ответ. Под капотом используется **HTTP/2**: одно TCP-соединение, сжатие, мультиплексирование запросов.

Отличие от REST в нашем случае:

| | REST (у нас) | gRPC (у нас) |
|---|--------------|--------------|
| Порт | 8080 | 9090 |
| Формат | JSON по HTTP/1.1 | Protocol Buffers по HTTP/2 |
| Контракт | Описан в коде (структуры Go, биндинги Gin) | Описан в `.proto` в отдельном репо |
| Кто ходит | Чаще всего браузер/фронт | Другие сервисы, мобильные приложения, тесты, CLI |

REST мы оставили: фронт и люди продолжают пользоваться JSON по 8080. gRPC добавлен для сервис-сервисного общения, для клиентов с кодогеном из одного контракта и для единого строго типизированного API.

---

### 2. Почему gRPC на отдельном порту

- **Разные протоколы.** На 8080 слушает HTTP-сервер (Gin): обычные GET/POST, JSON. На 9090 слушает gRPC-сервер: бинарный protobuf, другой формат рукопожатия (HTTP/2). Один и тот же TCP-порт не может быть «и REST, и gRPC» без прокси, который разбирает запросы по типу.
- **Разделение по ролям.** Фронт и браузерные клиенты — на 8080 (REST). Внутренние сервисы, мобильный бэкенд, инструменты (grpcurl, тесты) — на 9090 (gRPC). Так проще firewall’ы, мониторинг и понимание, кто куда ходит.
- **Можно выключить по отдельности.** В будущем можно поднимать только REST или только gRPC (например, в разных контейнерах), не трогая общий код use case’ов.

Порты задаются в конфиге: `CALCULATOR_SERVER_PORT=8080`, `CALCULATOR_GRPC_PORT=9090` (см. `.env` в `deployment/localCalc`).

---

### 3. Отдельный репо контрактов (calc-proto) и кодоген

**Почему контракты вынесены в отдельный репо**

- Один источник правды для API: описание методов и сообщений живёт в `.proto`, а не в коде конкретного сервиса.
- Несколько приложений могут использовать один и тот же контракт: lizzyCalc (Go), другой сервис на Python/Java, мобильное приложение — все генерируют клиент/сервер из одних и тех же `.proto`.
- Версионирование: в calc-proto ставят теги (v0.1.0, v1.0.0), и потребители подключают нужную версию контракта.

**Что лежит в calc-proto**

- **`api/calculator/v1/calculator.proto`** — описание сервиса `CalculatorService` с методами `Calculate` и `History` и сообщениями (request/response). Никакой Go-логики, только контракт.
- **Кодоген:** из этого файла с помощью компилятора **protoc** и плагинов **protoc-gen-go** и **protoc-gen-go-grpc** генерируются два Go-файла:
  - **`calculator.pb.go`** — структуры сообщений (CalculateRequest, CalculateResponse, HistoryItem и т.д.) и их сериализация в бинарный protobuf.
  - **`calculator_grpc.pb.go`** — интерфейс сервера `CalculatorServiceServer` (его реализует lizzyCalc) и клиент `CalculatorServiceClient` (им пользуются вызывающие).
- Генерация запускается в репо calc-proto командой **`make gen-go`**. Сгенерированный код коммитится в тот же репо (папка `gen/go/`), чтобы потребители могли делать `go get github.com/AraxHub/calc-proto` и не ставить у себя protoc.

**Как lizzyCalc использует calc-proto**

- В **go.mod** добавлена зависимость: `require github.com/AraxHub/calc-proto ...` и при локальной разработке — `replace github.com/AraxHub/calc-proto => ../calc-proto`, чтобы брать код из соседней папки.
- lizzyCalc **не генерирует** ничего из proto сам: он импортирует уже сгенерированный пакет `github.com/AraxHub/calc-proto/gen/go/calculator/v1` и реализует интерфейс `CalculatorServiceServer`, вызывая свои use case’ы (те же, что и для REST).

Итого: контракт и кодоген живут в calc-proto; lizzyCalc только подключает готовый модуль и «навешивает» на него свою бизнес-логику.

---

### 4. Как это работает в реальном приложении

**Старт приложения**

1. Читается конфиг (в т.ч. `CALCULATOR_GRPC_PORT=9090`).
2. Поднимаются БД и Redis, создаются репозиторий, кэш и use case калькулятора (общие для REST и gRPC).
3. Запускается **gRPC-сервер** в отдельной горутине: слушает порт 9090, регистрирует реализацию `CalculatorService` (наш handler в `internal/api/grpc/calculator`).
4. Запускается **HTTP-сервер** (Gin): слушает 8080, отдаёт REST (POST /api/v1/calculate, GET /api/v1/history).
5. При получении SIGINT/SIGTERM сначала останавливается HTTP, затем gRPC (graceful shutdown).

**Обработка gRPC-запроса**

1. Клиент (другой сервис, grpcurl, тест) подключается к `localhost:9090` и вызывает, например, `CalculatorService.Calculate` с телом `{ number1: 1, number2: 2, operation: "+" }` (в бинарном protobuf).
2. gRPC-библиотека десериализует запрос в `CalculateRequest`, находит зарегистрированный handler и вызывает метод **Calculate** нашей реализации (`internal/api/grpc/calculator/server.go`).
3. Там мы вызываем **тот же use case**, что и в REST: `uc.Calculate(ctx, req.GetNumber1(), req.GetNumber2(), req.GetOperation())`.
4. Use case считает (или достаёт из кэша), пишет в БД при необходимости и возвращает `*domain.Operation`.
5. Handler переводит результат в `CalculateResponse` и отдаёт клиенту по gRPC (бинарный ответ).

То есть **логика одна** (use case), **транспортов два**: REST (контроллер в `internal/api/http/controllers/calculator`) и gRPC (Server в `internal/api/grpc/calculator`). Дублируется только слой «запрос → use case → ответ».

**Проверка с хоста (grpcurl)**

- Установи [grpcurl](https://github.com/fullstorydev/grpcurl): `brew install grpcurl`.
- Запусти lizzyCalc. Наш gRPC-сервер **не отдаёт reflection** (мы не регистрируем Server Reflection API), поэтому grpcurl не может сам узнать список сервисов и методов. Нужно передать схему через `.proto` из репо calc-proto:

  ```bash
  # Путь к calc-proto — замени на свой (например от корня репо: ../calc-proto).
  CALC_PROTO="/Users/admin/liz education/calc-proto"

  # History (без тела запроса)
  grpcurl -plaintext -import-path "$CALC_PROTO" -proto api/calculator/v1/calculator.proto \
    localhost:9090 calculator.v1.CalculatorService/History

  # Calculate (с телом)
  grpcurl -plaintext -import-path "$CALC_PROTO" -proto api/calculator/v1/calculator.proto \
    -d '{"number1":10,"number2":5,"operation":"-"}' \
    localhost:9090 calculator.v1.CalculatorService/Calculate
  ```

**Что значит «не отдаёт reflection»**

- **gRPC reflection** — это когда сервер по специальному RPC отдаёт клиенту описание своих сервисов и методов (дескрипторы). Тогда grpcurl/Postman могут без `.proto` показать список методов и подставить поля.
- У нас reflection **не включён** (в коде нет `reflection.Register(grpcServer)`), поэтому клиент не может «спросить у сервера» схему и выдаёт *server does not support the reflection API*.
- Итог: при вызове с хоста нужно указывать `-proto` и `-import-path` на папку calc-proto, чтобы grpcurl знал контракт. Для «родных» клиентов (другой сервис с сгенерированным кодом из того же proto) это не нужно — у них схема уже в коде.

---

### 5. Интерцепторы (аналог HTTP middleware)

**Что это такое**

**Интерцепторы** в gRPC — это обёртки вокруг вызова RPC. Они выполняются до и после твоего handler’а (метода `Calculate`, `History` и т.д.) и позволяют делать общую логику без дублирования в каждом методе: логирование, метрики, аутентификация, трейсинг.

В HTTP это делают middleware (например `r.Use(middlewares.RequestLogger)` в Gin). В gRPC ту же роль играют **UnaryServerInterceptor** для обычных запросов (request → response) и **StreamServerInterceptor** для стримов.

**Зачем нужны**

- **Логирование** — каждый RPC логируем: имя метода, длительность, код ответа (OK / InvalidArgument / Internal). У нас это `LoggingUnaryInterceptor` в `internal/api/grpc/interceptors`: пишет в slog до и после вызова handler’а.
- **Метрики** — считать количество вызовов, латентность по методам.
- **Аутентификация/авторизация** — проверить токен в контексте, вернуть ошибку до вызова бизнес-логики.
- **Трейсинг** — прокинуть span ID, записать в trace.

Без интерцепторов пришлось бы в начале и конце каждого метода сервиса вручную вызывать логгер/метрики; с интерцептором эта логика в одном месте.

**Как работает низкоуровнево**

1. При регистрации сервиса gRPC сохраняет не только handler (твой `Calculate`/`History`), но и **цепочку интерцепторов**, переданную в `grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor1, interceptor2, ...))`.

2. Когда приходит unary-запрос, сервер не вызывает handler напрямую. Он вызывает **первый интерцептор**, передавая ему:
   - `ctx`, десериализованный `req`, описание метода (`*grpc.UnaryServerInfo`: имя RPC и т.д.) и **функцию `handler`** — это либо следующий интерцептор в цепочке, либо в конце сама твоя реализация (например `calculator.Server.Calculate`).

3. Сигнатура интерцептора:
   ```go
   func(ctx, req, info, handler) (resp, err)
   ```
   Интерцептор может: замерить время, залогировать вход, вызвать `handler(ctx, req)` (это передаёт управление дальше по цепочке или в твой код), залогировать выход и ошибку, вернуть результат. Если интерцептор не вызовет `handler`, твой метод сервиса не выполнится (например, auth-интерцептор вернёт ошибку и не вызовет handler).

4. Цепочка: запрос входит → Interceptor1 → Interceptor2 → … → твой `Calculate`/`History` → ответ поднимается обратно через те же интерцепторы (после `handler(...)` каждый может дописать лог/метрику и вернуть resp/err).

В нашем коде один интерцептор — логирующий; он вызывается для каждого unary RPC, измеряет время, после `handler(ctx, req)` пишет в лог метод, `latency_ms` и `grpc_code` (или ошибку).

---

### 6. Краткая схема по слоям

```
Клиент REST (браузер)          Клиент gRPC (сервис / grpcurl)
        |                                    |
        v                                    v
   :8080 (HTTP/JSON)                    :9090 (HTTP/2 + protobuf)
        |                                    |
        v                                    v
  api/http (Gin)                    api/grpc (grpc.Server)
  controllers/calculator           calculator/server.go
        |                                    |
        +--------------+-------------------+
                       v
              ports.CalculatorUseCase
                       |
                       v
              usecase/calculator (Calculate, History)
                       |
                       v
              repository + cache, domain
```

Контракт для gRPC живёт в репо **calc-proto** (`.proto` + сгенерированный Go); lizzyCalc только реализует этот контракт и поднимает второй сервер на отдельном порту.
