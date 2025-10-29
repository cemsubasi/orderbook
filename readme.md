# OrderBook Project

A real-time order book application with Go backend, React frontend, Kafka for event streaming, and PostgreSQL for storage. Supports live updates over WebSockets.

### Features
- Real-time order book updates using WebSockets
- Place buy/sell orders
-	Event-driven architecture with Kafka
-	Persistent storage with PostgreSQL
-	Dockerized for easy deployment

### Prerequisites
- Git
-	Docker & Docker Compose

### Setup
1.	Clone the repository:
```bash
git clone https://github.com/cemsubasi/orderbook.git
cd orderbook
```

2.	Open .env and fill in your variables or export environments:

```bash
vi .env
```
or
```bash
export BE_PORT=8080
```

3.	Build and start the containers:

```bash
docker-compose build --no-cache
docker-compose up -d
```

### Running
- Backend: http://localhost:8080
- Frontend: http://localhost:3000

### Docker Compose Services
- postgres: PostgreSQL database
- zookeeper: Kafka Zookeeper
- kafka: Kafka broker
- orderbook: Go backend
- frontend: React frontend

### Environment Variables
```env
# Backend config
BE_PORT

# Frontend config
FE_PORT

# PostgreSQL config
PG_USER
PG_PASS
PG_DB
PG_HOST

# Kafka config
KAFKA_HOST
KAFKA_PORT
```

### Database Migrations

The project includes three scripts for managing database migrations: create_migration.sh, apply_migrations.sh and drop_migrations.sh. These scripts use the Go Migrate tool.

To create a new migration, run:
```sh
./create_migration.sh <migration_name>
```
This generates new migration files with the specified name.

To apply pending migrations to the database, run:
```sh
./apply_migrations.sh
```
To revert applied migrations, run:
```sh
./drop_migrations.sh <migration_name>
```

The orderbook container automatically runs apply_migrations.sh after each build to ensure the database is up-to-date.

### Architecture & Design Notes

Overview

The system is built as a modular monolith consisting of several internal packages, each responsible for a single concern.
It provides an order matching engine with persistence and real-time updates via Kafka and WebSocket.

Modules
- engine/ → Core order matching logic. Manages in-memory order books and emits events (e.g., order_added, trade_executed) through Kafka.
- event/ → Handles Kafka integration.
  - KafkaPublisher: publishes order/trade events.
  - KafkaConsumers: listen to events and perform side effects (DB persistence).
- db/ → Database layer using pgxpool. Provides InitPostgres and RetrieveOrderBooks.
- api/ → HTTP controllers for handling external REST requests.
- ws/ → Real-time WebSocket hub for broadcasting snapshot updates.
- main.go → Application entrypoint. Initializes dependencies, starts Kafka consumers/publisher, loads state from DB, and runs the HTTP server.

Data Flow
1.	HTTP Request → API layer sends orders to engine.
2.	Engine → Processes and emits events to Kafka.
3.	Kafka Consumers → 
- Persist to DB (order_consumer, trade_consumer)
- Broadcast via WebSocket (ws_consumer)
4.	Web Clients → Receive real-time updates.

Key Design Choices
- **Engine independence:** The `engine` module has no external dependencies — it operates purely in-memory and only interacts with Kafka through an abstracted publisher interface. (except for one utility package `google/uuid` used due to project time constraints)
- **Decoupling via Kafka:** Producers (engine) and consumers (DB/ws) are independent; system remains resilient even if some consumers are temporarily down.
- **In-memory state recovery:** On startup, order books are reconstructed from the DB. *(Note: should reconcile with Kafka events to double-check and ensure engine state consistency.)*
- **Graceful shutdown:** Context cancellation propagates stop signals to the engine and Kafka components.
