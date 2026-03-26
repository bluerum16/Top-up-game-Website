# 🎮 GameTopUp Platform

> Marketplace top-up game terinspirasi Itemku. Full-stack monorepo dengan arsitektur
> dual-database, event streaming via Kafka, dan ML behavioral tracking.

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-14-black?logo=next.js)](https://nextjs.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?logo=postgresql)](https://postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://docker.com)
[![Kafka](https://img.shields.io/badge/Kafka-7.5-231F20?logo=apachekafka)](https://kafka.apache.org)
[![Python](https://img.shields.io/badge/Python-3.11-3776AB?logo=python)](https://python.org)

---

## 📋 Table of Contents

- [Arsitektur Sistem](#arsitektur-sistem)
- [Tech Stack](#tech-stack)
- [Struktur Project](#struktur-project)
- [Setup Development](#setup-development)
- [Setup VPS Production](#setup-vps-production)
- [Kafka: Dual Database](#kafka-dual-database)
- [Business Logic](#business-logic)
- [Database Relations](#database-relations)
- [Feature Breakdown](#feature-breakdown)
- [ML Tracking System](#ml-tracking-system)
- [Payment Integration (Midtrans)](#payment-integration-midtrans)
- [Publisher API Integration](#publisher-api-integration)
- [Useful Queries](#useful-queries)

---

## Arsitektur Sistem

```
VPS (Ubuntu 22.04)
└── Docker Compose
    ├── Nginx              → Reverse proxy, SSL, routing
    ├── frontend           → Next.js 14 + Tailwind (user site)
    ├── dashboard          → Next.js 14 + Tailwind (admin)
    ├── backend            → Golang + Gin (REST API)
    ├── ml_service         → Python + FastAPI (ML & recommendations)
    ├── kafka              → Event streaming main DB → dashboard DB
    ├── zookeeper          → Kafka dependency
    ├── postgres_main      → Database utama (38 tabel)
    ├── postgres_dashboard → Database analytics/dashboard
    └── redis              → Cache + session storage
```

**Kenapa dual database + Kafka?**
Database utama (`postgres_main`) menangani semua transaksi live — order, payment, user, top-up. Database dashboard (`postgres_dashboard`) khusus untuk analytics dan reporting. Dengan memisahkan keduanya, query berat dari dashboard tidak mengganggu performa transaksi live. Kafka menjadi jembatan: setiap event penting di main DB dikirim sebagai message ke Kafka, lalu dikonsumsi oleh service yang mengupdate dashboard DB.

---

## Tech Stack

| Layer | Teknologi | Versi |
|---|---|---|
| Frontend (user) | Next.js + Tailwind CSS | 14 / 3.4 |
| Frontend (admin) | Next.js + Tailwind CSS | 14 / 3.4 |
| Backend API | Golang + Gin | 1.22 / 1.9 |
| ML Service | Python + FastAPI | 3.11 / 0.110 |
| Database utama | PostgreSQL | 15 |
| Database dashboard | PostgreSQL | 15 |
| Cache / Session | Redis | 7 |
| Event streaming | Apache Kafka | 7.5 (Confluent) |
| Reverse proxy | Nginx | Alpine |
| Container | Docker + Docker Compose | 25+ |
| Hosting | VPS (Ubuntu 22.04) | — |
| Payment | Midtrans Snap | sandbox & production |
| Publisher | Direct API / Aggregator | — |
| ML Libraries | scikit-learn, pandas, numpy | — |

---

## Struktur Project

```
topup-platform/
├── docker-compose.yml
├── .env.example
├── .env                     ← jangan di-commit!
│
├── nginx/
│   ├── nginx.conf
│   └── ssl/                 ← SSL certificates
│
├── backend/                 ← Golang + Gin
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── database/        ← koneksi postgres, redis
│   │   ├── kafka/           ← producer
│   │   ├── router/
│   │   ├── middleware/      ← JWT, CORS, rate limit
│   │   ├── handler/         ← HTTP handlers per domain
│   │   ├── service/         ← business logic
│   │   └── repository/      ← DB queries (pgx)
│   ├── pkg/
│   │   ├── midtrans/
│   │   ├── publisher/
│   │   └── mailer/          ← SMTP
│   ├── Dockerfile
│   └── go.mod
│
├── frontend/                ← Next.js + Tailwind
│   ├── app/
│   │   ├── page.tsx
│   │   ├── game/[slug]/
│   │   ├── order/
│   │   └── contact/
│   ├── components/
│   ├── tailwind.config.ts
│   ├── Dockerfile
│   └── package.json
│
├── dashboard/               ← Next.js + Tailwind (admin)
│   ├── app/
│   │   ├── dashboard/
│   │   ├── orders/
│   │   ├── games/
│   │   ├── users/
│   │   └── contacts/
│   ├── Dockerfile
│   └── package.json
│
├── ml_service/              ← Python + FastAPI
│   ├── main.py
│   ├── routers/
│   ├── services/
│   ├── kafka/               ← consumer
│   ├── models/              ← trained .pkl files
│   ├── Dockerfile
│   └── requirements.txt
│
└── db/
    ├── main/
    │   ├── init.sql         ← schema utama (38 tabel)
    │   └── contact_emails.sql
    └── dashboard/
        └── init.sql         ← dashboard schema (8 tabel)
```

---

## Setup Development

### 1. Prerequisites

```bash
# Install Docker & Docker Compose
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# Verifikasi
docker --version          # Docker 25+
docker compose version    # v2.24+
```

### 2. Clone & Configure

```bash
git clone https://github.com/yourusername/topup-platform.git
cd topup-platform

cp .env.example .env
nano .env
```

Isi minimal ini di `.env`:

```env
APP_ENV=development
JWT_SECRET=ganti_dengan_random_string_panjang_min_32_karakter
POSTGRES_MAIN_PASS=password_kuat_1
POSTGRES_DASH_PASS=password_kuat_2
REDIS_PASS=password_redis
MIDTRANS_ENV=sandbox
MIDTRANS_SERVER_KEY=SB-Mid-server-xxxx
MIDTRANS_CLIENT_KEY=SB-Mid-client-xxxx
```

### 3. Jalankan semua service

```bash
# Development mode (termasuk Kafka UI di port 8090)
docker compose --profile dev up -d

# Cek semua container
docker compose ps
```

Output yang diharapkan:

```
NAME                  STATUS
topup_nginx           Up
topup_frontend        Up
topup_dashboard       Up
topup_backend         Up (healthy)
topup_ml              Up
topup_pg_main         Up (healthy)
topup_pg_dashboard    Up (healthy)
topup_redis           Up (healthy)
topup_kafka           Up
topup_zookeeper       Up
topup_kafka_ui        Up   ← dev only, akses di localhost:8090
```

### 4. Verifikasi endpoint

```bash
curl http://localhost:8080/health    # → {"status":"ok"}
curl http://localhost:8000/health    # → {"status":"ok"}

# Buka di browser
# http://localhost:3000        → frontend user
# http://localhost:3001        → dashboard admin (atau subdomain admin.)
# http://localhost:8090        → Kafka UI
```

### 5. Reset database

```bash
# Hapus semua data dan mulai dari awal
docker compose down -v
docker compose up -d
```

---

## Setup VPS Production

### 1. Persiapan server

```bash
# Update & install Docker
sudo apt update && sudo apt upgrade -y
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# Install Certbot
sudo apt install certbot -y
```

### 2. Point domain ke IP VPS

Tambahkan DNS records di provider domain kamu:

```
A    yourdomain.com        → IP_VPS
A    www.yourdomain.com    → IP_VPS
A    admin.yourdomain.com  → IP_VPS
```

### 3. Generate SSL (Let's Encrypt gratis)

```bash
sudo certbot certonly --standalone \
  -d yourdomain.com \
  -d www.yourdomain.com \
  -d admin.yourdomain.com

sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem   nginx/ssl/
sudo chown $USER:$USER nginx/ssl/*
```

### 4. Konfigurasi production

```bash
nano .env
```

```env
APP_ENV=production
MIDTRANS_ENV=production
MIDTRANS_SERVER_KEY=Mid-server-xxxx     # key production (tanpa SB-)
MIDTRANS_CLIENT_KEY=Mid-client-xxxx
PUBLISHER_ENV=production
```

### 5. Build & deploy

```bash
# Production (Kafka UI tidak ikut jalan)
docker compose up -d --build

docker compose ps    # semua harus Up
```

### 6. Auto-renew SSL (cron)

```bash
crontab -e
# Tambahkan baris ini:
0 3 * * * certbot renew --quiet && \
  cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem /path/project/nginx/ssl/ && \
  cp /etc/letsencrypt/live/yourdomain.com/privkey.pem /path/project/nginx/ssl/ && \
  docker compose -f /path/project/docker-compose.yml restart nginx
```

### 7. Perintah maintenance VPS

```bash
# Restart satu service
docker compose restart backend

# Lihat log real-time
docker compose logs -f backend
docker compose logs -f kafka

# Update setelah git pull
git pull
docker compose up -d --build backend frontend dashboard ml_service

# Backup database
docker exec topup_pg_main pg_dump \
  -U topup_user gametopup_main > backup_main_$(date +%Y%m%d).sql

docker exec topup_pg_dashboard pg_dump \
  -U dash_user gametopup_dashboard > backup_dash_$(date +%Y%m%d).sql
```

---

## Kafka: Dual Database

Kafka menghubungkan `postgres_main` dan `postgres_dashboard` secara async.

```
Go Backend (producer)
    │
    ├── order sukses    ──▶ topup.orders    ──▶ Dashboard consumer → dash_revenue_daily
    ├── payment settle  ──▶ topup.payments  ──▶ Dashboard consumer → dash_payment_stats_daily
    ├── user baru       ──▶ topup.users     ──▶ Dashboard consumer → dash_user_growth_daily
    ├── page view       ──▶ topup.pageviews ──▶ ML consumer (Python) → ml_page_views
    ├── search event    ──▶ topup.search    ──▶ ML consumer (Python) → ml_search_events
    └── contact form    ──▶ topup.contacts  ──▶ Go consumer → SMTP email ke admin
```

### Kafka Topics

| Topic | Producer | Consumer | Tujuan |
|---|---|---|---|
| `topup.orders` | Go | Dashboard svc | Revenue stats |
| `topup.payments` | Go | Dashboard svc | Payment breakdown |
| `topup.users` | Go | Dashboard svc | User growth |
| `topup.pageviews` | Go | Python ML | ml_page_views |
| `topup.search` | Go | Python ML | ml_search_events |
| `topup.contacts` | Go | Go mailer | Email notif admin |

---

## Business Logic

### Alur Top-Up End-to-End

```
User buka game page
        │
        ▼
Input Player ID + Server ID
        │
        ▼  (verify API publisher → cached Redis 24 jam)
Tampilkan nama akun game
        │
        ▼
Pilih nominal + metode bayar
        │
        ▼
Buat order (status: pending)
        │
        ▼  (Gin → Midtrans create transaction → snap_token)
User bayar di Midtrans Snap popup
        │
        ▼  (Midtrans webhook POST → Gin handler)
Verifikasi HMAC signature
        │
        ▼  order: paid → processing
        │  Kafka produce → topup.orders
Go → Publisher API
        │
        ├── sukses → order: success
        │                 │
        │                 ▼
        │           ml_user_topup_affinity += 1 (DB trigger)
        │           Kafka → dashboard consumer
        │
        └── gagal  → order: failed → refund
```

### Order Status Flow

```
pending → awaiting_payment → paid → processing → success
                │                       │
                ▼                       ▼
             expired                  failed → refunded
```

### Contact Email Flow

```
User isi form → POST /api/contact (Gin)
    → INSERT contact_emails
    → Kafka topup.contacts
    → Go consumer → SMTP → email notif ke admin
```

---

## Database Relations

```
users
 ├── user_sessions        (1:N)
 ├── user_wallets         (1:1)
 │    └── wallet_transactions (1:N)
 ├── sellers              (1:1)
 ├── orders               (1:N)
 └── contact_emails       (1:N) — nullable

games
 ├── categories           (N:1)
 ├── game_topup_config    (1:1)
 ├── products             (1:N)
 │    └── product_variants (1:N)
 └── publisher_integrations (1:N per env)
      └── publisher_product_mapping (1:N)

orders
 ├── payment_transactions (1:1) — Midtrans data
 ├── order_status_logs    (1:N) — trigger otomatis
 └── publisher_api_logs   (1:N)

postgres_main ──[Kafka]──▶ postgres_dashboard
```

### Tabel Kunci

| Tabel | FK Utama | Keterangan |
|---|---|---|
| `orders` | `user_id`, `product_id`, `variant_id` | Inti transaksi |
| `payment_transactions` | `order_id` | 1:1, field Midtrans lengkap |
| `publisher_product_mapping` | `variant_id`, `integration_id` | Mapping SKU |
| `ml_user_topup_affinity` | `user_id`, `game_id`, `variant_id` | UNIQUE, auto-update |
| `contact_emails` | `user_id?`, `order_id?` | Form contact |

---

## Feature Breakdown

### 👤 Users & Auth
Registrasi/login email+phone, JWT+Redis session, verifikasi email/phone, referral system, saldo/wallet, multi-alamat.

### 🎮 Games & Produk
Katalog multi-platform, multi-tipe produk (top_up, item, account, gift_card, dll), varian/nominal dengan `cost_price` untuk tracking margin, multi-seller marketplace, review untuk pembeli terverifikasi saja.

### 🔑 Top-Up Specific
Form dinamis per game via JSONB config. Verifikasi Player ID real-time di-cache Redis 24 jam.

### 📦 Orders
9 status dengan audit trail otomatis (trigger). Auto-generate order number `TXN-YYYYMMDD-XXXXXXXX`. Voucher persentase/fixed per-game atau per-user.

### ✉️ Contact Email
Form kontak dengan 5 kategori, assignment ke admin, status penanganan, notifikasi SMTP via Kafka.

### 📊 Dashboard (database terpisah)
Revenue harian/bulanan, game performance, payment breakdown, user growth, search trends, order funnel — semua diisi via Kafka dari main DB.

### 🤖 ML Recommendations
Rekomendasi personalisasi, trending games, autocomplete search — diproses Python FastAPI dan disimpan di main DB.

---

## ML Tracking System

| Pertanyaan | Tabel Raw | Tabel Agregat | Update |
|---|---|---|---|
| Apa yang sering diketik? | `ml_search_events` | `ml_search_aggregates` | Daily job (Python) |
| Top-up apa yang sering dilakukan? | — | `ml_user_topup_affinity` | DB Trigger (otomatis) |
| Game mana yang sering dibuka? | `ml_page_views` | `ml_game_popularity` | Daily job (Python) |

Output model disimpan di `ml_recommendations` dan `ml_search_suggestions`, di-serve oleh Python FastAPI ke Go backend.

---

## Payment Integration (Midtrans)

### Flow Midtrans Snap

```
1. POST /api/order       → Gin buat order + hit Midtrans API → dapat snap_token
2. Frontend load Snap.js → popup bayar
3. User bayar (QRIS / e-wallet / VA / kartu)
4. Midtrans POST webhook → /webhook/midtrans (Gin)
5. Gin verifikasi: SHA512(order_id + status_code + gross_amount + server_key)
6. Update order → proses top-up ke publisher
```

### Ganti Environment

```env
MIDTRANS_ENV=sandbox      # development
MIDTRANS_ENV=production   # live
```

Backend otomatis membaca config dari tabel `payment_gateway_config` berdasarkan nilai ini.

---

## Publisher API Integration

**Direct API** — daftar ke Garena/Moonton/Supercell. Margin besar, butuh verifikasi bisnis.

**Aggregator** — via Digiflazz/UniPin/Codashop API. Setup cepat, cocok untuk MVP.

```env
PUBLISHER_ENV=sandbox      # pakai sandbox aggregator
PUBLISHER_ENV=production   # pakai publisher live
```

---

## Useful Queries

```sql
-- Top 10 pencarian minggu ini
SELECT * FROM v_top_searches_7d LIMIT 10;

-- Game trending bulan ini
SELECT * FROM v_trending_games LIMIT 10;

-- Favorite games seorang user
SELECT * FROM v_user_favorite_games WHERE user_id = '<uuid>';

-- Monitor transaksi Midtrans sandbox
SELECT o.order_number, o.total_amount, pt.payment_type,
       pt.transaction_status, pt.created_at
FROM payment_transactions pt
JOIN orders o ON o.id = pt.order_id
WHERE pt.env = 'sandbox'
ORDER BY pt.created_at DESC LIMIT 50;

-- Revenue per game bulan ini
SELECT g.name, COUNT(o.id) AS orders, SUM(o.total_amount) AS revenue
FROM orders o
JOIN products p ON p.id = o.product_id
JOIN games g ON g.id = p.game_id
WHERE o.status = 'success'
  AND o.created_at >= DATE_TRUNC('month', NOW())
GROUP BY g.name ORDER BY revenue DESC;

-- Contact emails belum dibalas
SELECT name, email, subject, category, created_at
FROM contact_emails
WHERE status IN ('new', 'read')
ORDER BY created_at ASC;

-- Dashboard: Revenue hari ini (dari dashboard DB)
SELECT gross_revenue, net_revenue, total_orders, success_orders
FROM dash_revenue_daily
WHERE date = CURRENT_DATE;
```

---

## Schema Stats

| Kategori | Jumlah |
|---|---|
| Tabel main DB | 38 (37 + contact_emails) |
| Tabel dashboard DB | 8 |
| Indexes | 35+ |
| Views | 4 |
| Triggers | 9 |
| Kafka topics | 6 |
| Docker services | 10 |

---

## Roadmap

- [ ] Table partitioning `ml_search_events` & `ml_page_views` (> 10M rows)
- [ ] `pgvector` untuk embedding-based recommendations
- [ ] GitHub Actions CI/CD → VPS via SSH
- [ ] Prometheus + Grafana untuk monitoring
- [ ] Row-level security (RLS) untuk seller isolation

---

## License

MIT