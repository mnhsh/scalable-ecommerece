# GoCart

A microservices e-commerce backend built with Go. Features async stock updates via RabbitMQ, JWT auth with role-based access, and a CLI demo tool.

Built while learning backend development through [boot.dev](https://boot.dev).

## Table of Contents

- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [How It Works](#how-it-works)
- [Contributing](#contributing)

## Architecture

```
                                    ┌─────────────┐
                                    │     CLI     │
                                    └──────┬──────┘
                                           │
                                           ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                            API Gateway (:8080)                                │
│                    JWT Auth · RBAC · Request Routing                          │
└──────────────────────────────────────────────────────────────────────────────┘
         │                    │                    │                    │
         ▼                    ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  User Service   │  │ Product Service │  │  Cart Service   │  │  Order Service  │
│    (:8081)      │  │    (:8082)      │  │    (:8083)      │  │    (:8084)      │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ • Registration  │  │ • Product CRUD  │  │ • Add/remove    │  │ • Create order  │
│ • JWT Auth      │  │ • Stock mgmt    │  │ • Update qty    │  │ • Cancel order  │
│ • Token refresh │  │ • RabbitMQ      │  │ • Price lookup  │  │ • RabbitMQ      │
│                 │  │   consumer      │  │                 │  │   publisher     │
└────────┬────────┘  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘
         │                    │                    │                    │
         ▼                    ▼                    ▼                    ▼
    [PostgreSQL]         [PostgreSQL]        [PostgreSQL]         [PostgreSQL]
      :5433                :5434               :5435                :5436

                    ┌───────────────────────────────────────┐
                    │              RabbitMQ                 │
                    │          (:5672 / :15672)             │
                    ├───────────────────────────────────────┤
                    │  Exchange: orders                     │
                    │  • order.created  → decrement stock   │
                    │  • order.cancelled → restore stock    │
                    └───────────────────────────────────────┘
                              ▲                   │
                              │                   │
                   publish ───┘                   └─── consume
                (order-service)               (product-service)
```

## Tech Stack

- **Go 1.22** - All services
- **PostgreSQL 16** - One database per service
- **RabbitMQ** - Async messaging between services
- **SQLC** - Type-safe SQL code generation
- **Goose** - Database migrations
- **JWT + Argon2id** - Auth and password hashing
- **Docker Compose** - Container orchestration

## Quick Start

### 1. Start everything

```bash
docker compose up -d
```

### 2. Run migrations

```bash
goose -dir services/user-service/sql/schema postgres \
  "postgres://postgres:postgres@localhost:5433/users?sslmode=disable" up

goose -dir services/product-service/sql/schema postgres \
  "postgres://postgres:postgres@localhost:5434/products?sslmode=disable" up

goose -dir services/cart-service/sql/schema postgres \
  "postgres://postgres:postgres@localhost:5435/carts?sslmode=disable" up

goose -dir services/order-service/sql/schema postgres \
  "postgres://postgres:postgres@localhost:5436/orders?sslmode=disable" up
```

### 3. Build and run CLI

```bash
cd services/cli
go build -o cli .
./cli
```

### Admin Login

| Email | Password |
|-------|----------|
| `admin@example.com` | `admin123` |

## Usage

### CLI Demo

The CLI lets you test the full flow:

```
========================================
       ECOM CLI - E-Commerce Store     
========================================

--- Welcome ---
1. Login
2. Register
3. Exit
```

**As a customer:** Register → Browse products → Add to cart → Checkout → View/cancel orders

**As admin:** Login with admin credentials → Manage Products → Add/delete products

### API Examples

```bash
# Register
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# Get products
curl http://localhost:8080/api/products

# Add to cart (need token from login)
curl -X POST http://localhost:8080/api/cart/items \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"product_id": "<uuid>", "quantity": 2}'

# Create order
curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer <token>"

# Cancel order
curl -X DELETE http://localhost:8080/api/orders/<order-id> \
  -H "Authorization: Bearer <token>"
```

## API Endpoints

### Public

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/users` | Register |
| POST | `/api/login` | Login |
| POST | `/api/refresh` | Refresh token |
| POST | `/api/revoke` | Logout |
| GET | `/api/products` | List products |
| GET | `/api/products/{id}` | Get product |

### Protected (Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/me` | Current user |
| GET | `/api/cart` | Get cart |
| POST | `/api/cart/items` | Add to cart |
| PATCH | `/api/cart/items/{id}` | Update quantity |
| DELETE | `/api/cart/items/{id}` | Remove item |
| DELETE | `/api/cart` | Clear cart |
| POST | `/api/orders` | Create order |
| GET | `/api/orders` | List orders |
| GET | `/api/orders/{id}` | Get order |
| DELETE | `/api/orders/{id}` | Cancel order |

### Admin Only

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/admin/products` | Create product |
| PATCH | `/admin/products/{id}` | Update product |
| DELETE | `/admin/products/{id}` | Delete product |

## How It Works

The interesting part is how stock updates happen asynchronously via RabbitMQ:

```
┌──────────────┐                        ┌─────────────────┐
│Order Service │ ── order.created ────▶ │ Product Service │
│  (publisher) │                        │   (consumer)    │
│              │ ── order.cancelled ──▶ │                 │
└──────────────┘                        └─────────────────┘
                                               │
                                               ▼
                                        [Update stock]
```

1. User creates order → order-service publishes `order.created`
2. product-service consumes it → decrements stock
3. User cancels order → order-service publishes `order.cancelled`
4. product-service consumes it → restores stock

This keeps the services decoupled. Order-service doesn't need to know how stock updates work - it just fires events and moves on.

## Contributing

### Clone and setup

```bash
git clone https://github.com/mnhsh/scalable-ecommerce.git
cd scalable-ecommerce
docker compose up -d
```

### Run migrations

```bash
goose -dir services/user-service/sql/schema postgres \
  "postgres://postgres:postgres@localhost:5433/users?sslmode=disable" up

goose -dir services/product-service/sql/schema postgres \
  "postgres://postgres:postgres@localhost:5434/products?sslmode=disable" up

goose -dir services/cart-service/sql/schema postgres \
  "postgres://postgres:postgres@localhost:5435/carts?sslmode=disable" up

goose -dir services/order-service/sql/schema postgres \
  "postgres://postgres:postgres@localhost:5436/orders?sslmode=disable" up
```

### Build CLI

```bash
cd services/cli
go build -o cli .
```

### Run tests

```bash
go test ./services/...
```

### Regenerate database code

```bash
cd services/<service-name> && sqlc generate
```

### Submit a pull request

Fork the repo and open a PR to `main`.
