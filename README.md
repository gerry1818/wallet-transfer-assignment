# 💸 Wallet Transfer Service

A backend service to perform **wallet-to-wallet transfers** with strong guarantees:

* ✅ Idempotency (exactly-once API behavior)
* ✅ Atomicity (debit + credit succeed together)
* ✅ Consistency (ledger always balanced)
* ✅ Concurrency safety (no double spending)

---

# ⚡ Quick Start

```bash
# Start PostgreSQL
docker run --name wallet-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=wallet \
  -p 5432:5432 \
  -d postgres

# Create schema
psql -h localhost -U postgres -d wallet -f migrations/schema.sql

# insert sample data into wallet
psql -h localhost -U postgres -d wallet -f migrations/seed.sql



# Run service
go mod tidy
go run cmd/server/main.go
```

👉 Service running at: `http://localhost:8080`

---

# 🚀 Tech Stack

* Golang
* PostgreSQL
* pgxpool (DB connection pooling)
* Docker
* Zap (structured logging)
* Prometheus (metrics)

---

# 📦 Project Structure

```text
wallet-transfer-service/
│
├── cmd/server/              # App entrypoint
├── internal/
│   ├── handler/            # HTTP layer
│   ├── service/            # Business logic
│   ├── repository/         # Interfaces
│   ├── repository/postgres # DB implementation
│   ├── model/              # Request/response models
│   ├── db/                 # DB connection
│   ├── logger/             # Logging (zap)
│   └── metrics/            # Prometheus metrics
│
├── migrations/             # SQL schema
├── test/                   # Unit tests
├── go.mod
└── README.md
```

---

# 📡 API Specification

## 🔹 POST /transfers

### Request

```json
{
  "idempotencyKey": "txn-1",
  "fromWalletId": 1,
  "toWalletId": 2,
  "amount": 100
}
```

---

### ✅ Success Response

```json
{
  "transferId": 1,
  "status": "PROCESSED"
}
```

---

### ❌ Error Response

```json
{
  "error": "insufficient balance"
}
```

---

### 🔁 Idempotency Behavior

Each request is protected using:

`idempotencyKey + request hash`

---

## 🔁 Behavior

### 1. First request
- Stored in DB
- Transfer executes

---

### 2. Duplicate request (same key + same payload)
- Returns cached response
- No re-execution

---

### 3. Duplicate request (same key + different payload)
- ❌ Rejected (400/409)

---

### 4. In-progress request

```json
{
  "error": "request in progress"
}
```

---

# 🔒 Transaction Flow

All transfers execute inside a **single DB transaction**:

1. Lock wallets (`SELECT ... FOR UPDATE`)
2. Validate balance
3. Create transfer record
4. Update wallet balances
5. Insert ledger entries (DEBIT + CREDIT)
6. Update transfer state
7. Commit transaction

---

# ⚡ Concurrency Safety

* Row-level locking prevents race conditions
* Prevents double spending
* Ensures consistency under concurrent requests

---

# 🗃️ Sample Data

```sql
INSERT INTO wallets (id, balance) VALUES
(1, 1000),
(2, 500);
```

---

# 📊 Metrics

Available at:

```
GET /metrics
```

Tracks:

* transfer_success_total
* transfer_failure_total

---

⚠️ HTTP Status Codes

- 200 → Success
- 400 → Validation / business rule failure
- 409 → Request in progress / idempotency conflict
- 500 → Internal server error

---

# 🧪 Testing

Run all tests:

```bash
go test ./...
```

Covers:

* ✅ Successful transfer
* ❌ Insufficient balance
* 🔁 Idempotency behavior

---

# 🧠 Design Decisions

* Strong consistency using PostgreSQL transactions
* Pessimistic locking (SELECT FOR UPDATE)
* Double-entry ledger system
* Interface-based repository for testability
* Request hash-based idempotency

---

# ⚖️ Tradeoffs

* Strong consistency over availability
* Slight performance cost due to locking
* Simpler architecture (no distributed systems)

---

# ⚠️ Assumptions

* Integer-only transfers
* Wallet creation is out of scope
* Single PostgreSQL instance

---

# 🚀 Future Improvements

* Distributed locking (scale-out)
* Retry on deadlocks
* Rate limiting
* Authentication/authorization
* Event-driven architecture (Kafka)

---

# 📝 Logging Example

```json
{
  "level": "info",
  "msg": "transfer success",
  "transfer_id": 1
}
```

---

# 👨‍💻 Author

Girish Prajapati

---

# 🤖 AI Usage

This project used AI for:

* system design brainstorming
* concurrency strategy validation
* code structuring assistance
* test improvements
