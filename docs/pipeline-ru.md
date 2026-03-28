# Описание пайплайна сбора данных

## 1. Общая архитектура

```
┌──────────────────────────────────────────────────────────────┐
│                         Сервис                               │
│                                                              │
│  ┌─────────────┐   GenerateTasks()   ┌──────────────────┐    │
│  │  Scheduler  │ ──────────────────► │  Task Repository │    │
│  │  (чистая    │                     │  (хранит задачи  │    │
│  │   логика)   │                     │   + ScheduledAt) │    │
│  └─────────────┘                     └────────┬─────────┘    │
│                                               │              │
│                                    Dispatcher │ (фоновый     │
│                                    читает     │  процесс)    │
│                                               ▼              │
└───────────────────────────────────────────────┼──────────────┘
                                                │ публикует
                          ┌─────────────────────┼─────────────────────┐
                          │     NATS JetStream  │                     │
                          │  STATIONS.load  ◄───┘                     │
                          │  TICKETS.parse  ◄─────────────────────────┤
                          │                                           │
                          │  STATIONS.result ──────────────────────►  │
                          │  TICKETS.result  ──────────────────────►  │
                          └───────────────────────────────────────────┘
                                    ▲                        │
                              подписка                  результаты
                                    │                        ▼
                          ┌─────────────────┐     ┌──────────────────┐
                          │     Worker      │     │     Сервис       │
                          │  (процессоры)   │     │ (сохраняет в     │
                          └─────────────────┘     │  Dgraph)         │
                                                  └──────────────────┘
```

Ключевые принципы:
- **Scheduler** — чистая функция, не знает ни о NATS, ни о Dgraph
- **Dispatcher** — фоновый процесс, отправляет задачи в NATS когда `ScheduledAt <= now`
- **Worker** — подписывается на топик, выполняет задачу, публикует результат
- **Темп работы** задаётся через `ScheduledAt` на этапе генерации задач, а не на стороне воркера

---

## 2. Поток 1 — Загрузка станций

### 2.1. Генерация задачи (Scheduler)

Сервис вызывает планировщик раз в сутки:

```go
task := scheduler.GenerateStationTask(time.Now())
// StationLoadTask{ ScheduledAt: today, TopN: 100 }
repo.Save(task)
```

### 2.2. Диспетчер отправляет задачу в NATS

Фоновый процесс периодически читает из репозитория задачи, у которых `ScheduledAt <= now`, и публикует их в `STATIONS.load`.

### 2.3. Worker — `StationLoadProcessor`

```
Получить задачу из STATIONS.load
    └─► Скачать страницу Википедии «Города России» (goquery)
        └─► Распарсить HTML-таблицу → []CityData {Name, Region, Population}
            └─► Отсортировать по населению, взять топ-N
                └─► BuildPrefixGroups(cities, prefix=3, maxGroup=20)
                    └─► Для каждой префиксной группы:
                            rzd.SuggestCity(prefix)   ← один запрос на группу
                                └─► Индекс: name → RZD NodeID   (из resp.City)
                                └─► Индекс: NodeID → []Station  (из resp.Train, rail only)
                                    └─► Сопоставить каждый город группы → станции
                └─► Опубликовать StationLoadResult в STATIONS.result
```

**`StationLoadResult`:**
```json
{
  "cities": [
    {
      "name": "Москва",
      "population": 12506468,
      "stations": [
        { "external_id": "2000000", "name": "Москва Казанская", "transport_type": "rail" }
      ]
    }
  ]
}
```

**Оптимизация:** вместо 100 запросов к RZD делается ~5–15 (по числу уникальных 3-символьных префиксов).

---

## 3. Поток 2 — Загрузка билетов

### 3.1. Генерация задач (Scheduler)

```go
tasks := scheduler.GenerateTicketTasks(time.Now(), connections)
repo.SaveAll(tasks)
```

Алгоритм внутри `GenerateTicketTasks`:

```
1. Для каждой связи (origin → destination):
       Для каждой даты в [today+1 .. today+90]:
           shouldParse(connection, today, date)?  ← матрица частоты
               └─► Создать TicketParseTask{Priority: computePriority(...)}

2. sort.Slice(tasks) по Priority ascending

3. Разбить на бакеты:
       bucketSize = rand[BucketSizeMin, BucketSizeMax]   (по умолчанию 5–10)
       pause      = rand[BucketPauseMin, BucketPauseMax] (по умолчанию 15–30 сек)
       cursor     = today
       ┌──────────────────────────────────────────────────────┐
       │ Бакет 1 (7 задач)  ScheduledAt = today+00:00:00      │
       │ Бакет 2 (5 задач)  ScheduledAt = today+00:00:22      │
       │ Бакет 3 (9 задач)  ScheduledAt = today+00:00:44      │
       │ ...                                                   │
       └──────────────────────────────────────────────────────┘
```

Задачи с высоким приоритетом (A↔A, ближние даты) получают ранние `ScheduledAt`.

### 3.2. Диспетчер отправляет задачи в NATS

Фоновый процесс читает из репозитория задачи, у которых `ScheduledAt <= now`, и публикует их в `TICKETS.parse` — строго в соответствии с расписанием из шага 3.1.

### 3.3. Матрица частоты (`shouldParse`)

| Маршрут | 1–14 дн. | 15–45 дн. | 46–90 дн. |
|---------|----------|-----------|-----------|
| A ↔ A  | 1        | 2         | 4         |
| A ↔ B  | 1        | 3         | 7         |
| A ↔ C  | 2        | 5         | 14        |
| A ↔ D  | 3        | 7         | 21        |
| B ↔ B  | 1        | 4         | 10        |
| B ↔ C  | 2        | 7         | 14        |
| B ↔ D  | 3        | 10        | 21        |
| C ↔ C  | 3        | 10        | 21        |
| C ↔ D  | 5        | 14        | 30        |
| D ↔ D  | 7        | 21        | 60        |

Тиры: A (>1М чел.), B (500К–1М), C (100К–500К), D (<100К).
Спящий режим: связь не использовалась в маршрутах >30 дней → интервал перекрывается до 21 дня.

### 3.4. Worker — `TicketParseProcessor`

```
Получить задачу из TICKETS.parse
    └─► ProxyManager.GetClient()       ← HTTP-клиент с прокси
        └─► rzd.ParseTrains(origin, dest, date)
            ├─► Успех (200 OK)
            │       └─► Опубликовать TicketParseResult в TICKETS.result
            │           └─► msg.Ack()
            │
            ├─► Бан прокси (429 / 403)
            │       └─► ProxyManager.Ban(httpClient)
            │           └─► msg.Nak(delay=1 мин)
            │
            └─► Сетевая ошибка / таймаут
                    └─► msg.Nak(delay=30 сек)
```

**`TicketParseResult`:**
```json
{
  "origin_code": "2000000",
  "destination_code": "2060600",
  "departure_date": "2026-04-10T00:00:00Z",
  "trains": [
    { "external_id": "016А", "departure_at": "...", "arrival_at": "...", "min_price": 1850.0 }
  ]
}
```

---

## 4. Жизненный цикл задачи

```
Scheduler        Repository       Dispatcher       NATS           Worker
────────────────────────────────────────────────────────────────────────────
GenerateTasks()
──────────────►  SaveAll(tasks)
                 (с ScheduledAt)

                 [каждые N сек]
                 WHERE scheduled_at <= now
                 ◄──────────────
                                  publish(msg) ──► [In Queue]
                                                   [In Progress] ◄── Run()
                                                                      process()
                                                   [Done] ◄────────── Ack()     ✓
                                                   [+30s] ◄────────── Nak(30s)  ✗ сеть
                                                   [+1m]  ◄────────── Nak(1min) ✗ бан
                                                   [Drop] ◄────────── Ack()     ✗ битый payload
```

---

## 5. Gateway — контракт воркера

```go
type Gateway interface {
    Subscribe(ctx context.Context, topic string,
        handler func(ctx context.Context, payload []byte) error) error
    Publish(ctx context.Context, topic string, data []byte) error
}
```

Конкретная реализация (NATS JetStream) инжектируется снаружи. Обработка Ack/Nak:
- `nil` → Ack
- `*NakError{Delay}` → Nak с заданной задержкой
- другая ошибка → Nak с дефолтной задержкой (30 сек)

---

## 6. Поиск маршрутов (HTTP API)

```
GET /route?from=Москва&to=Сыктывкар&date=2026-04-10
    └─► DgraphItineraryBuilder.Build()
            └─► Обход графа: City → Station → Trip → Station → City
                └─► Вернуть []RoutePoint { Name, StartAt, EndAt, Details }
```

Каждый раз, когда алгоритм использует связь между станциями, обновляется `LastUsedAt` — именно это поле определяет, перейдёт ли связь в спящий режим.

---

## 7. Схема данных в Dgraph

```
City
  ├── city.name        : string  (индекс term)
  ├── city.population  : int
  └── has_station ──►  Station
                         ├── station.name           : string
                         ├── station.transport_type : string
                         ├── station.external_id    : string  (7-значный код РЖД)
                         └── departs ──►  Trip
                                           ├── trip.external_id   : string
                                           ├── trip.price         : float
                                           ├── trip.departure_at  : datetime
                                           ├── trip.arrival_at    : datetime
                                           └── destination ──►  Station
```

---

## 8. Конфигурация планировщика

| Переменная | По умолчанию | Описание |
|---|---|---|
| `SCHEDULER_TOP_N` | `100` | Топ-N городов по населению |
| `SCHEDULER_DAYS_AHEAD` | `90` | Горизонт дат для парсинга |
| `SCHEDULER_BUCKET_SIZE_MIN` | `5` | Минимум задач в одном бакете |
| `SCHEDULER_BUCKET_SIZE_MAX` | `10` | Максимум задач в одном бакете |
| `SCHEDULER_BUCKET_PAUSE_MIN` | `15s` | Минимальная пауза между бакетами |
| `SCHEDULER_BUCKET_PAUSE_MAX` | `30s` | Максимальная пауза между бакетами |

---

## 9. Инфраструктура

| Компонент | Технология | Роль |
|---|---|---|
| Граф БД | Dgraph | Хранение городов, станций, поездок |
| Очередь | NATS JetStream | Доставка задач воркерам, публикация результатов |
| Сервис | Go + chi | HTTP API, scheduler, dispatcher, сохранение результатов |
| Воркер | Go | Процессоры задач (station, ticket) |
| Наблюдаемость | OpenTelemetry + Zap | Трейсы, метрики, логи |
