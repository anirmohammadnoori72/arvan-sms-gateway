
# Arvan SMS Gateway

Arvan SMS Gateway is a scalable, production-grade service designed to handle **100 million SMS messages per day (~1200 messages per second)**.  
It focuses on **availability, consistency, and flexibility** using technologies like **Kafka, PostgreSQL, Redis, and containerized Gateway/Worker Pods**.

---

## Features

- **High Throughput**: Handles large-scale traffic with Kafka as a message broker.
- **VIP/OTP Optimization**: Prioritizes critical messages using dedicated VIP topics.
- **Configurable Processing**:
  - Sync balance checks for instant consistency.
  - Async balance checks for burst traffic resilience.
- **Wallet Reservation (Optional)**:
  - Uses Redis for reserving tokens to reduce Postgres lock contention.
  - Can be disabled for stricter consistency.
- **Horizontally Scalable**:
  - Gateway Pods and Worker Pods can scale independently.
- **Swagger-Documented APIs**:
  - `/send-sms`
  - `/balance/{user_id}`
  - `/message-status/{message_id}`

---

## Architecture

1. **Gateway Pods (Top Layer)**:
   - Receive API calls.
   - Validate and log messages.
   - Perform balance checks (via Postgres or Redis Reservation).
   - Push messages to Kafka (Normal or VIP Topic).

2. **Kafka (Middle Layer)**:
   - Acts as a buffer for decoupling producers and consumers.
   - Two topics:
     - Normal Topic (Standard traffic).
     - VIP/Express Topic (Priority traffic).

3. **Worker Pods (Bottom Layer)**:
   - Consume messages from Kafka.
   - Send SMS via external provider (mocked).
   - Update message status in Postgres.
   - For VIP, balance deduction can be **Sync or Async**.

4. **Postgres & Redis (Right Side)**:
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

3. Access Swagger UI:
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
- Responses:
  - `200 OK`: `{"status":"pending","message_id":"uuid"}`
  - `400 Bad Request`: Invalid phone, duplicate message ID, or insufficient balance.
  - `500 Internal Server Error`: Server or Kafka issue.

### Check Balance
- **GET** `/balance/{user_id}`
- Responses:
  - `200 OK`: `{"balance":1000}`
  - `400 Bad Request`: Invalid user ID.
  - `500 Internal Server Error`: DB issues.

### Check Message Status
- **GET** `/message-status/{message_id}`
- Responses:
  - `200 OK`: `{"status":"pending|sent|failed|rejected"}`
  - `400 Bad Request`: Invalid message ID.
  - `500 Internal Server Error`: DB issues.

---

## Scaling Considerations

While the system scales horizontally via pods, **Postgres may become a bottleneck** under extreme growth.  
Possible solutions:
- Implement **logical sharding** by customer.
- Use **NewSQL / distributed SQL databases** (CockroachDB, Yugabyte).
- Cache-heavy read paths via Redis for non-critical reads.

---

## License
MIT License.
