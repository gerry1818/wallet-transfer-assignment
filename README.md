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
docker exec -i wallet-db psql -U postgres -d wallet < migrations/schema.sql

# insert sample data into wallet
docker exec -i wallet-db psql -U postgres -d wallet < migrations/seed.sql


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

* Each request must include a unique `idempotencyKey`
* Duplicate requests:

  * Return cached response
  * Do NOT re-execute transaction

If request is still processing:

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
3. Update wallet balances
4. Insert ledger entries (DEBIT + CREDIT)
5. Create transfer record
6. Commit transaction

---

# ⚡ Concurrency Safety

* Uses row-level locking
* Prevents double spending
* Ensures serialized updates per wallet

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

* success count
* failure count
* idempotency hits

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

* Strong consistency using DB transactions
* Pessimistic locking for correctness
* Double-entry ledger system
* Interface-based repository for testability

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
