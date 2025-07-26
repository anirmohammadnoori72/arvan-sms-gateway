# Arvan SMS Gateway

Arvan SMS Gateway is a scalable, production-grade service designed to handle **100 million SMS messages per day (~1200 messages per second)**.  
It focuses on **availability, consistency, and flexibility** using technologies like **Kafka, PostgreSQL, Redis, and containerized Gateway/Worker Pods**.

---

## Features

- **High Throughput**: Handles large-scale traffic with Kafka as a message broker.
- **VIP/OTP Optimization**: Dedicated `sms-vip` topic for high-priority traffic, separate from `sms-normal`.
- **UUID Enforcement**:
  - Both **`message_id`** and **`user_id`** must be valid **UUIDs**.
  - Duplicate `message_id` will result in a `400 Bad Request`.
- **Configurable Processing**:
  - Sync balance checks for strict consistency.
  - Async reservation via Redis for burst load resilience.
- **Horizontally Scalable**:
  - **Two Workers**:
    - **Normal Worker**: consumes `sms-normal`.
    - **VIP Worker**: consumes `sms-vip`.
  - Gateway Pods and Worker Pods can scale independently.
- **Swagger-Documented APIs**:
  - `/send-sms`
  - `/balance/{user_id}`
  - `/message-status/{message_id}`

---

## Architecture

1. **Gateway Pods (Top Layer)**:
   - Receive API calls.
   - Validate `UUID`s, phone number, and message size.
   - Check balance (via Postgres or Redis Reservation).
   - Push to Kafka (`sms-normal` or `sms-vip`).

2. **Kafka (Middle Layer)**:
   - Acts as a buffer between producers and consumers.
   - Two topics:
     - **Normal Topic (`sms-normal`)**: standard traffic.
     - **VIP Topic (`sms-vip`)**: critical/priority traffic.

3. **Worker Pods (Bottom Layer)**:
   - **Normal Worker**: consumes from `sms-normal`.
   - **VIP Worker**: consumes from `sms-vip`.
   - Sends SMS via external provider (mocked).
   - Updates message status in Postgres.

4. **Postgres & Redis (Data Layer)**:
   - **Postgres**: Source of Truth for wallets, reservations, and message logs.
   - **Redis**: Optional caching/reservation layer for high-load scenarios.

---

## Setup

### Prerequisites
- Docker & Docker Compose
- Go 1.21+
- Kafka & Zookeeper (docker-compose ready)
- PostgreSQL
- Redis

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/arvan-sms-gateway.git
   cd arvan-sms-gateway
   ```

2. Build and run:
   ```bash
   docker-compose up -d
   go run cmd/server/main.go
   ```

3. Start workers:
   ```bash
   go run cmd/worker-normal/main.go
   go run cmd/worker-vip/main.go
   ```

4. Access Swagger UI:
   ```
   http://localhost:8081/swagger/index.html
   ```

---

## API Endpoints

### Send SMS
- **POST** `/send-sms`
- Request:
  ```json
  {
    "message_id": "uuid",
    "user_id": "uuid",
    "phone_number": "+1234567890",
    "message": "Hello, World!"
  }
  ```
- Requirements:
  - Both `message_id` and `user_id` must be **valid UUIDs**.
  - `message_id` must be unique (duplicates cause `400 Bad Request`).
  - Phone number must be valid (minimum 10 digits).
  - Message must not be empty (max 500 characters).
- Responses:
  - `200 OK`: `{"status":"pending","message_id":"uuid"}`
  - `400 Bad Request`: Invalid UUID, phone, or duplicate `message_id`.
  - `500 Internal Server Error`: Server or Kafka issue.

### Check Balance
- **GET** `/balance/{user_id}`
- Requirements:
  - `user_id` must be a valid UUID.
- Responses:
  - `200 OK`: `{"user_id":"...","balance":1000}`
  - `400 Bad Request`: Invalid UUID format.
  - `404 Not Found`: User not found.
  - `500 Internal Server Error`: Database issues.

### Check Message Status
- **GET** `/message-status/{message_id}`
- Requirements:
  - `message_id` must be a valid UUID.
- Responses:
  - `200 OK`: `{"message_id":"...","status":"pending|sent|failed|rejected"}`
  - `400 Bad Request`: Invalid UUID format.
  - `404 Not Found`: Message not found.
  - `500 Internal Server Error`: Database issues.

---

## Config (Environment Variables)

The service reads configuration from environment variables.  
Defaults are provided for local development.

```go
type Config struct {
    ServerPort          string // API server port
    ServiceName         string
    KafkaBrokers        string
    KafkaTopicNormal    string // "sms-normal"
    KafkaTopicVIP       string // "sms-vip"
    DBHost              string
    DBPort              string
    DBUser              string
    DBPassword          string
    DBName              string
    DeveloperMode       string
    ServerMetricsPort   string // ":9090" for Prometheus metrics
    DBUrl               string
    RedisAddr           string
    BatchSize           int64  // wallet batch operations
    ReservationTTL      int64  // TTL for Redis reservation (seconds)
    UseRedisReservation bool   // enable or disable Redis reservations
}
```

---

## Scaling Considerations

While the system scales horizontally via pods, **Postgres may become a bottleneck** at extreme scale.  
Mitigation strategies:
- Implement **logical sharding** (per customer or region).
- Move to **distributed SQL databases** like CockroachDB or Yugabyte.
- Use Redis to cache read-heavy paths for non-critical lookups.

---

## License
MIT License.


---

## Example Usage (Test Users)

The database includes two predefined test users (inserted via migrations):

- **VIP User**  
  `user_id = 11111111-1111-1111-1111-111111111111`  
  Balance: `10,000,000`  
  `is_vip = TRUE`

- **Normal User**  
  `user_id = 22222222-2222-2222-2222-222222222222`  
  Balance: `10,000,000`  
  `is_vip = FALSE`

### Example API Calls

#### Send SMS (VIP User)
```bash
curl -X POST http://localhost:8081/send-sms   -H "Content-Type: application/json"   -d '{
    "message_id": "31111111-1111-1111-1111-111111111111",
    "user_id": "11111111-1111-1111-1111-111111111111",
    "phone_number": "+989121234567",
    "message": "Hello VIP!"
  }'
```

#### Send SMS (Normal User)
```bash
curl -X POST http://localhost:8081/send-sms   -H "Content-Type: application/json"   -d '{
    "message_id": "32222222-2222-2222-2222-222222222222",
    "user_id": "22222222-2222-2222-2222-222222222222",
    "phone_number": "+989121234568",
    "message": "Hello Normal!"
  }'
```

#### Check Balance (VIP User)
```bash
curl -X GET http://localhost:8081/balance/11111111-1111-1111-1111-111111111111
```

#### Check Balance (Normal User)
```bash
curl -X GET http://localhost:8081/balance/22222222-2222-2222-2222-222222222222
```

