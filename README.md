# 🎮 GameTopUp Platform — Database Schema

> PostgreSQL 15+ database schema for a game top-up marketplace inspired by Itemku.
> Includes full e-commerce flow, Midtrans payment integration (sandbox & production),
> Publisher API integration, and ML-powered behavioral tracking.

---

## 📋 Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Business Logic](#business-logic)
- [Database Relations](#database-relations)
- [Feature Breakdown](#feature-breakdown)
- [ML Tracking System](#ml-tracking-system)
- [Payment Integration (Midtrans)](#payment-integration-midtrans)
- [Publisher API Integration](#publisher-api-integration)
- [Getting Started](#getting-started)
- [Useful Queries](#useful-queries)

---

## Overview

Platform ini adalah marketplace top-up game yang memungkinkan:
- **Buyer** membeli top-up, item, akun, gift card, dan voucher game
- **Seller** membuka toko dan menjual produk mereka sendiri
- **Admin** mengelola platform, publisher API, dan payout

Schema dirancang untuk skala produksi dengan 37 tabel, partisi untuk tabel ML bervolume tinggi, dan integrasi dua environment (sandbox & production) untuk payment dan publisher API.

---

## Tech Stack

| Layer | Teknologi |
|---|---|
| Database | PostgreSQL 15+ |
| Extensions | `uuid-ossp`, `pg_trgm`, `btree_gin`, `pgcrypto` |
| Payment | Midtrans Snap (sandbox & production) |
| Publisher | Direct API / Aggregator (Digiflazz, UniPin, dll) |
| ML Pipeline | External (Python/Spark) — data disimpan di DB |
| Backend | GO/GIN |
| Frontend | NextJs |
| Environtment | Docker |
| Hosting | VPS |

---

## Business Logic

### Alur Top-Up End-to-End

```
User buka game page
        │
        ▼
Input Player ID + Server ID
        │
        ▼ (hit publisher verify API, cached 24 jam)
Tampilkan nama akun game user
        │
        ▼
Pilih nominal (product_variant)
        │
        ▼
Pilih metode pembayaran
        │
        ▼
Buat order (status: pending)
        │
        ▼ (Midtrans create transaction → dapat snap_token)
User bayar di halaman Midtrans
        │
        ▼ (Midtrans kirim webhook ke server)
Verifikasi signature webhook
        │
        ▼ (order status: paid → processing)
Server kirim request ke Publisher API
        │
        ├─── sukses ──▶ order status: success
        │                    │
        │                    ▼
        │             ml_user_topup_affinity terupdate (trigger)
        │
        └─── gagal  ──▶ order status: failed → refund otomatis
```

### Model Bisnis Multi-Seller

Platform menggunakan model **marketplace** dimana:

1. **Platform products** — produk yang dikelola langsung oleh admin (seller_id = NULL), terhubung langsung ke Publisher API
2. **Seller products** — produk yang didaftarkan seller independen, harga kompetitif antar seller untuk produk game yang sama
3. **Margin** dihitung dari `price` (harga jual) vs `cost_price` (COGS/harga beli dari publisher) di tabel `product_variants`

### Order Status Flow

```
pending
    │
    ▼
awaiting_payment ──(expired)──▶ expired
    │
    ▼ (webhook settlement)
paid
    │
    ▼
processing ──(publisher error)──▶ failed ──▶ refunded
    │
    ▼
success
```

Setiap perubahan status dicatat otomatis di `order_status_logs` via trigger.

### Voucher & Diskon

- Voucher bisa bersifat **global** (semua game) atau **spesifik per game**
- Bisa dibatasi per user tertentu atau publik
- Tipe diskon: **persentase** atau **nominal tetap**
- `max_discount` membatasi nilai diskon maksimum untuk tipe persentase
- Penggunaan dicatat di `voucher_usages` (UNIQUE per order, mencegah double pakai)

---

## Database Relations

### Diagram Relasi Utama

```
users
 ├── user_sessions        (1:N) — device & browser tracking
 ├── user_addresses       (1:N)
 ├── user_wallets         (1:1) — saldo platform
 │    └── wallet_transactions (1:N)
 ├── sellers              (1:1) — jika user jadi seller
 └── orders               (1:N)

games
 ├── categories           (N:1)
 ├── game_tags            (1:N)
 ├── game_topup_config    (1:1) — field dinamis per game (Player ID, Server ID, dll)
 ├── products             (1:N)
 │    ├── product_variants    (1:N) — setiap nominal/denomination
 │    ├── product_images      (1:N)
 │    └── product_reviews     (1:N)
 └── publisher_integrations  (1:N, per env)
      └── publisher_product_mapping (1:N) — mapping SKU ke kode publisher

orders
 ├── users                (N:1)
 ├── products             (N:1)
 ├── product_variants     (N:1)
 ├── payment_transactions (1:1)
 ├── order_status_logs    (1:N) — audit trail status
 ├── publisher_api_logs   (1:N) — setiap request ke publisher
 └── publisher_delivery_status (1:1) — untuk publisher async
```

### Tabel Kunci & Foreign Keys

| Tabel | FK Utama | Keterangan |
|---|---|---|
| `orders` | `user_id`, `product_id`, `variant_id` | Inti transaksi |
| `payment_transactions` | `order_id` | 1:1 dengan order |
| `product_variants` | `product_id` | Varian/nominal per produk |
| `publisher_product_mapping` | `variant_id`, `integration_id` | Mapping SKU |
| `publisher_api_logs` | `order_id`, `integration_id` | Log setiap API call |
| `ml_user_topup_affinity` | `user_id`, `game_id`, `variant_id` | UNIQUE constraint |
| `voucher_usages` | `voucher_id`, `order_id` | UNIQUE, cegah double pakai |

---

## Feature Breakdown

### 👤 Users & Auth

| Fitur | Tabel | Detail |
|---|---|---|
| Registrasi & login | `users` | Email + phone, password bcrypt |
| Multi-device session | `user_sessions` | IP, user-agent, device type, OS |
| Verifikasi email/phone | `users` | `email_verified_at`, `phone_verified_at` |
| Referral system | `users` | `referral_code`, `referred_by` |
| Saldo/wallet | `user_wallets` | Balance + locked amount |
| Histori wallet | `wallet_transactions` | Debit/credit/hold/release |
| Alamat | `user_addresses` | Multi-alamat, default flag |

### 🎮 Games & Produk

| Fitur | Tabel | Detail |
|---|---|---|
| Katalog game | `games` | Multi-platform, multi-genre |
| Kategori | `categories` | Hierarki flat |
| Tag game | `game_tags` | "popular", "new", "sale" |
| Multi-tipe produk | `products` | top_up, item, account, gift_card, dll |
| Varian/nominal | `product_variants` | Tiap nominal = 1 baris, ada `cost_price` |
| Galeri gambar | `product_images` | Multiple images per produk |
| Review terverifikasi | `product_reviews` | `is_verified` = hanya pembeli yang bisa review |
| Multi-seller | `sellers` | Rating, total_sales, verified badge |

### 🔑 Top-Up Specific

| Fitur | Tabel | Detail |
|---|---|---|
| Form dinamis per game | `game_topup_config` | JSONB fields config (Player ID, Server ID, dll) |
| Verifikasi Player ID | `player_id_cache` | Cache 24 jam, hindari spam publisher API |

**Contoh `game_topup_config.fields`:**
```json
[
  {"key": "player_id", "label": "Player ID", "type": "text", "required": true},
  {"key": "server_id", "label": "Server", "type": "select", "required": true,
   "options": [{"label": "Asia", "value": "1"}, {"label": "North America", "value": "2"}]}
]
```

### 📦 Orders

| Fitur | Tabel | Detail |
|---|---|---|
| Order management | `orders` | Lengkap dengan pricing breakdown |
| Status tracking | `orders` | 9 status: pending → success/failed/refunded |
| Audit trail status | `order_status_logs` | Trigger otomatis setiap status berubah |
| Auto order number | Trigger | Format `TXN-20240301-XXXXXXXX` |
| Voucher | `vouchers`, `voucher_usages` | Persentase/fixed, per-game, per-user |

### 🔔 Notifikasi

| Fitur | Tabel | Detail |
|---|---|---|
| In-app notifications | `notifications` | Order success/failed, promo, system |
| Read tracking | `notifications` | `is_read`, `read_at` |

### 💰 Seller Payout

| Fitur | Tabel | Detail |
|---|---|---|
| Payout management | `seller_payouts` | Bank transfer, status tracking |

### 🔍 Audit

| Fitur | Tabel | Detail |
|---|---|---|
| Full audit log | `audit_logs` | Semua aksi penting tercatat + IP |

---

## ML Tracking System

Sistem ML dirancang untuk menjawab tiga pertanyaan bisnis:

### 1. Orang lebih sering mengetik apa?

**Tabel:** `ml_search_events` → diproses ke `ml_search_aggregates`

```
User mengetik di search bar
        │
        ▼
ml_search_events (raw, setiap keystroke/submit)
  - query text
  - apakah hasil di-klik? (clicked_result)
  - posisi klik (clicked_position)
  - context: homepage/game_page/global
        │
        ▼ (scheduled job harian)
ml_search_aggregates (agregat per hari)
  - total_searches
  - unique_users
  - conversion (led to purchase)
        │
        ▼
ml_search_suggestions (autocomplete, pre-computed ML)
  - fuzzy match via pg_trgm
```

**Query autocomplete (pg_trgm):**
```sql
SELECT query_normalized, total_searches
FROM ml_search_aggregates
WHERE query_normalized % 'mobile leg'   -- similarity match
   OR query_normalized ILIKE 'mobile leg%'
ORDER BY total_searches DESC
LIMIT 8;
```

### 2. Orang lebih sering melakukan top-up apa?

**Tabel:** `ml_user_topup_affinity` — **diupdate otomatis via trigger** setiap order berhasil

```sql
-- Trigger fn_update_topup_affinity() otomatis jalan
-- ketika order.status berubah menjadi 'success'
INSERT INTO ml_user_topup_affinity (user_id, game_id, variant_id, ...)
ON CONFLICT DO UPDATE SET
    purchase_count = purchase_count + 1,
    total_spent    = total_spent + amount;
```

Data ini digunakan untuk:
- **Upsell**: "Kamu biasa beli 100 Diamond, mau coba 500?"
- **Rekomendasi**: "User seperti kamu sering beli ini"
- **Email marketing**: segmentasi berdasarkan game favorit

### 3. Orang lebih sering buka top-up game yang mana?

**Tabel:** `ml_page_views` → diproses ke `ml_game_popularity`

```
ml_page_views (raw events)
  - page_type: 'game', 'product', 'home'
  - referrer_type: search/banner/direct/social
  - time_on_page_ms
  - scroll_depth (0–100%)
        │
        ▼ (ML pipeline harian)
ml_game_popularity
  - popularity_score (computed)
  - trend_direction: 'up'/'down'/'stable'
        │
        ▼
ml_recommendations
  - personalized per user
  - trending global
  - similar games
```

### ML Tables Summary

| Tabel | Fungsi | Update Frequency |
|---|---|---|
| `ml_search_events` | Raw search log | Real-time |
| `ml_page_views` | Raw page view log | Real-time |
| `ml_user_events` | Raw click/hover/cart | Real-time |
| `ml_user_topup_affinity` | Frekuensi top-up per user | Trigger (auto) |
| `ml_search_aggregates` | Agregat pencarian harian | Scheduled (daily) |
| `ml_game_popularity` | Skor popularitas game | Scheduled (daily) |
| `ml_recommendations` | Output model ML | Scheduled (6 jam) |
| `ml_search_suggestions` | Autocomplete suggestions | Scheduled (harian) |

### Pre-built Views

```sql
-- Top pencarian minggu ini
SELECT * FROM v_top_searches_7d LIMIT 10;

-- Game trending bulan ini
SELECT * FROM v_trending_games LIMIT 10;

-- Game favorit seorang user
SELECT * FROM v_user_favorite_games WHERE user_id = '<uuid>';

-- Nominal yang sering dibeli user
SELECT * FROM v_user_favorite_variants WHERE user_id = '<uuid>';
```

---

## Payment Integration (Midtrans)

### Dua Environment

```sql
-- Config tersimpan di payment_gateway_config
-- sandbox  → https://api.sandbox.midtrans.com
-- production → https://api.midtrans.com
```

### Flow Midtrans Snap

```
1. Backend buat transaksi → POST /v1/transactions (Midtrans API)
   Payload: order_id, gross_amount, customer_details, item_details

2. Midtrans return snap_token + redirect_url
   → disimpan di payment_transactions.snap_token

3. Frontend load Snap.js dengan snap_token
   → popup halaman bayar Midtrans

4. User bayar (QRIS / e-wallet / VA / kartu)

5. Midtrans kirim webhook POST ke endpoint kamu
   → disimpan di payment_transactions.raw_notification

6. Verifikasi signature:
   SHA512( order_id + status_code + gross_amount + server_key )
   → payment_transactions.signature_verified = true

7. Update order status → paid → proses top-up ke publisher
```

### Field Midtrans di `payment_transactions`

| Field | Keterangan |
|---|---|
| `midtrans_order_id` | Order ID yang dikirim ke Midtrans |
| `snap_token` | Token untuk Snap.js popup |
| `transaction_status` | `settlement`, `capture`, `pending`, `expire`, `cancel` |
| `fraud_status` | `accept`, `challenge`, `deny` |
| `va_number` | Nomor Virtual Account |
| `qr_code_url` | URL QR Code untuk QRIS |
| `raw_notification` | Full JSON payload dari webhook |
| `signature_verified` | Boolean hasil verifikasi HMAC |
| `env` | `sandbox` / `production` |

### Mengganti Environment

Cukup ubah satu kolom di `orders`:
```sql
-- Development / testing
UPDATE orders SET payment_env = 'sandbox' WHERE id = '<order_id>';

-- Production
UPDATE orders SET payment_env = 'production' WHERE id = '<order_id>';
```

Backend membaca config dari `payment_gateway_config` berdasarkan `env` ini.

---

## Publisher API Integration

### Dua Model Integrasi

**1. Direct API** — daftar langsung ke publisher game (Garena, Moonton, Supercell)
- Margin lebih besar
- Perlu verifikasi bisnis & deposit
- Cocok untuk game populer dengan volume tinggi

**2. Aggregator** — via Digiflazz, UniPin, VIP Reseller, Codashop API
- Setup cepat, ratusan produk sekaligus
- Margin lebih kecil
- Direkomendasikan untuk MVP / tahap awal

### Config per Game per Environment

```sql
INSERT INTO publisher_integrations
  (game_id, env, provider, api_key, base_url, auth_type, is_async)
VALUES
  ('<game_uuid>', 'sandbox',    'digiflazz', 'xxx', 'https://api.digiflazz.com', 'hmac', false),
  ('<game_uuid>', 'production', 'garena',    'xxx', 'https://api.garena.com',    'hmac', false);
```

### Mapping SKU

```sql
INSERT INTO publisher_product_mapping
  (variant_id, integration_id, publisher_product_id, publisher_price)
VALUES
  ('<variant_uuid>', '<integration_uuid>', 'ML-86-DIAMOND', 14000);
--   ^ variant "86 Diamond" kamu       ^ kode produk di publisher ^ harga beli
```

### Retry & Async Support

- `max_retries` — berapa kali retry jika publisher API timeout
- `is_async = true` — publisher tidak langsung return sukses/gagal; sistem polling via `publisher_delivery_status`
- Semua request dicatat di `publisher_api_logs` beserta latency (`response_ms`) dan nomor attempt

---

## Getting Started

### Prerequisites

```bash
PostgreSQL 15+
psql atau client GUI (DBeaver, TablePlus, pgAdmin)
```

### Setup

```bash
# 1. Buat database
createdb gametopup_db

# 2. Jalankan schema
psql -d gametopup_db -f topup_schema.sql

# 3. Verifikasi
psql -d gametopup_db -c "\dt"
# Harus muncul 37 tabel
```

### Konfigurasi Midtrans Sandbox

```sql
-- Ganti dengan Midtrans Sandbox keys kamu
-- Dashboard: https://dashboard.sandbox.midtrans.com
UPDATE payment_gateway_config
SET
  server_key = 'SB-Mid-server-XXXXXXXXXXXXXXXXXXXXXXXX',
  client_key = 'SB-Mid-client-XXXXXXXXXXXXXXXXXXXXXXXX'
WHERE env = 'sandbox';
```

### Environment Variables (rekomendasi backend)

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=gametopup_db
DB_USER=postgres
DB_PASS=yourpassword

MIDTRANS_ENV=sandbox
MIDTRANS_SERVER_KEY=SB-Mid-server-xxx
MIDTRANS_CLIENT_KEY=SB-Mid-client-xxx

PUBLISHER_ENV=sandbox
```

---

## Useful Queries

```sql
-- 10 pencarian terpopuler minggu ini
SELECT query, total_searches, conversion_rate
FROM v_top_searches_7d
LIMIT 10;

-- Game trending bulan ini
SELECT name, page_views, total_orders, trend
FROM v_trending_games
LIMIT 10;

-- Riwayat top-up seorang user
SELECT game_name, total_purchases, total_spent
FROM v_user_favorite_games
WHERE user_id = '<user_uuid>';

-- Monitor transaksi Midtrans sandbox
SELECT o.order_number, o.total_amount, pt.payment_type,
       pt.transaction_status, pt.env, pt.created_at
FROM payment_transactions pt
JOIN orders o ON o.id = pt.order_id
WHERE pt.env = 'sandbox'
ORDER BY pt.created_at DESC
LIMIT 50;

-- Revenue per game bulan ini
SELECT g.name, COUNT(o.id) AS orders, SUM(o.total_amount) AS revenue
FROM orders o
JOIN products p ON p.id = o.product_id
JOIN games g ON g.id = p.game_id
WHERE o.status = 'success'
  AND o.created_at >= DATE_TRUNC('month', NOW())
GROUP BY g.name
ORDER BY revenue DESC;

-- Konversi gap: user sering lihat tapi tidak pernah beli
SELECT pv.game_id, g.name, COUNT(*) AS views, COUNT(DISTINCT o.id) AS purchases
FROM ml_page_views pv
JOIN games g ON g.id = pv.game_id
LEFT JOIN orders o
  ON o.user_id = pv.user_id
  AND o.product_id IN (SELECT id FROM products WHERE game_id = pv.game_id)
  AND o.status = 'success'
WHERE pv.user_id = '<user_uuid>'
GROUP BY pv.game_id, g.name
HAVING COUNT(DISTINCT o.id) = 0;
```

---

## Schema Stats

| Kategori | Jumlah |
|---|---|
| Total tabel | 37 |
| Total index | 30+ |
| Views | 4 |
| Triggers | 8 |
| Custom functions | 4 |
| ENUM types | 11 |

---

## Roadmap

- [ ] Table partitioning untuk `ml_search_events` dan `ml_page_views` (> 10M rows)
- [ ] `pgvector` extension untuk embedding-based recommendations
- [ ] Read replica config untuk ML query workload
- [ ] Row-level security (RLS) untuk seller isolation
- [ ] TimescaleDB untuk time-series analytics

---

## License

MIT