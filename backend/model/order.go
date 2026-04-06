package model

import (
	"time"

	"github.com/google/uuid"
)

// Struct DB — cerminan langsung tabel orders
type Order struct {
    ID                uuid.UUID  `db:"id"                  json:"id"`
    OrderNumber       string     `db:"order_number"        json:"order_number"`
    UserID            uuid.UUID  `db:"user_id"             json:"user_id"`
    ProductID         uuid.UUID  `db:"product_id"          json:"product_id"`
    VariantID         uuid.UUID  `db:"variant_id"          json:"variant_id"`
    PlayerID          string     `db:"player_id"           json:"player_id"`
    ServerID          string     `db:"server_id"           json:"server_id"`
    PlayerDisplayName string     `db:"player_display_name" json:"player_display_name"`
    Quantity          int        `db:"quantity"            json:"quantity"`
    UnitPrice         int64      `db:"unit_price"          json:"unit_price"`
    Subtotal          int64      `db:"subtotal"            json:"subtotal"`
    DiscountAmount    int64      `db:"discount_amount"     json:"discount_amount"`
    VoucherCode       string     `db:"voucher_code"        json:"voucher_code,omitempty"`
    PlatformFee       int64      `db:"platform_fee"        json:"platform_fee"`
    TotalAmount       int64      `db:"total_amount"        json:"total_amount"`
    Status            string     `db:"status"              json:"status"`
    PaymentMethod     string     `db:"payment_method"      json:"payment_method,omitempty"`
    PaymentEnv        string     `db:"payment_env"         json:"payment_env"`
    PaidAt            *time.Time `db:"paid_at"             json:"paid_at,omitempty"`
    CompletedAt       *time.Time `db:"completed_at"        json:"completed_at,omitempty"`
    ExpiredAt         *time.Time `db:"expired_at"          json:"expired_at,omitempty"`
    FailureReason     string     `db:"failure_reason"      json:"failure_reason,omitempty"`
    CreatedAt         time.Time  `db:"created_at"          json:"created_at"`
    UpdatedAt         time.Time  `db:"updated_at"          json:"updated_at"`
}

// Input dari frontend saat buat order baru
type CreateOrderRequest struct {
    VariantID     string `json:"variant_id"     validate:"required,uuid"`
    PlayerID      string `json:"player_id"      validate:"required"`
    ServerID      string `json:"server_id"`
    PaymentMethod string `json:"payment_method" validate:"required"`
    VoucherCode   string `json:"voucher_code"`
}

// Response detail order — JOIN dari beberapa tabel
type OrderDetail struct {
    Order                      // embed semua field Order
    GameName     string        `db:"game_name"      json:"game_name"`
    GameSlug     string        `db:"game_slug"      json:"game_slug"`
    GameThumb    string        `db:"game_thumbnail" json:"game_thumbnail"`
    ProductName  string        `db:"product_name"   json:"product_name"`
    VariantName  string        `db:"variant_name"   json:"variant_name"`
    CurrencyUnit string        `db:"currency_unit"  json:"currency_unit"`
    SnapToken    string        `db:"snap_token"     json:"snap_token,omitempty"`
    PaymentType  string        `db:"payment_type"   json:"payment_type,omitempty"`
    StatusLogs   []StatusLog   `db:"-"              json:"status_logs,omitempty"`
}

// Untuk list order (lebih ringkas dari OrderDetail)
type OrderListItem struct {
    ID          uuid.UUID `db:"id"           json:"id"`
    OrderNumber string    `db:"order_number" json:"order_number"`
    GameName    string    `db:"game_name"    json:"game_name"`
    GameThumb   string    `db:"game_thumb"   json:"game_thumb"`
    VariantName string    `db:"variant_name" json:"variant_name"`
    TotalAmount int64     `db:"total_amount" json:"total_amount"`
    Status      string    `db:"status"       json:"status"`
    CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}

// Dari tabel order_status_logs
type StatusLog struct {
    FromStatus string    `db:"from_status" json:"from_status"`
    ToStatus   string    `db:"to_status"   json:"to_status"`
    Note       string    `db:"note"        json:"note,omitempty"`
    CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}