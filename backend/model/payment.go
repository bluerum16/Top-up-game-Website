package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Cerminan tabel payment_transactions
type PaymentTransaction struct {
    ID                uuid.UUID        `db:"id"                   json:"id"`
    OrderID           uuid.UUID        `db:"order_id"             json:"order_id"`
    Env               string           `db:"env"                  json:"env"`
    MidtransOrderID   string           `db:"midtrans_order_id"    json:"midtrans_order_id"`
    TransactionID     string           `db:"transaction_id"       json:"transaction_id,omitempty"`
    PaymentType       string           `db:"payment_type"         json:"payment_type,omitempty"`
    GrossAmount       int64            `db:"gross_amount"         json:"gross_amount"`
    TransactionStatus string           `db:"transaction_status"   json:"transaction_status"`
    FraudStatus       string           `db:"fraud_status"         json:"fraud_status,omitempty"`
    SnapToken         string           `db:"snap_token"           json:"snap_token,omitempty"`
    SnapRedirectURL   string           `db:"snap_redirect_url"    json:"snap_redirect_url,omitempty"`
    VANumber          string           `db:"va_number"            json:"va_number,omitempty"`
    Bank              string           `db:"bank"                 json:"bank,omitempty"`
    QRCodeURL         string           `db:"qr_code_url"          json:"qr_code_url,omitempty"`
    RawNotification   json.RawMessage  `db:"raw_notification"     json:"-"` // tidak di-expose
    SignatureVerified  bool            `db:"signature_verified"   json:"signature_verified"`
    TransactionTime   *time.Time       `db:"transaction_time"     json:"transaction_time,omitempty"`
    SettlementTime    *time.Time       `db:"settlement_time"      json:"settlement_time,omitempty"`
    ExpiryTime        *time.Time       `db:"expiry_time"          json:"expiry_time,omitempty"`
    CreatedAt         time.Time        `db:"created_at"           json:"created_at"`
}

// Payload webhook dari Midtrans — bukan dari DB, tapi dari HTTP body
type MidtransWebhookPayload struct {
    OrderID           string      `json:"order_id"`
    TransactionStatus string      `json:"transaction_status"`
    FraudStatus       string      `json:"fraud_status"`
    GrossAmount       string      `json:"gross_amount"` // string di Midtrans!
    SignatureKey      string      `json:"signature_key"`
    PaymentType       string      `json:"payment_type"`
    TransactionID     string      `json:"transaction_id"`
    TransactionTime   string      `json:"transaction_time"`
    SettlementTime    string      `json:"settlement_time,omitempty"`
    VANumbers         []VANumber  `json:"va_numbers,omitempty"`
    Actions           []Action    `json:"actions,omitempty"` // untuk GoPay deeplink
}

type VANumber struct {
    Bank     string `json:"bank"`
    VANumber string `json:"va_number"`
}

type Action struct {
    Name   string `json:"name"`
    Method string `json:"method"`
    URL    string `json:"url"`
}

// Config gateway — di-load dari payment_gateway_config saat startup
type GatewayConfig struct {
    Env       string `db:"env"`
    Provider  string `db:"provider"`
    ServerKey string `db:"server_key"`
    ClientKey string `db:"client_key"`
    BaseURL   string `db:"base_url"`
    SnapURL   string `db:"snap_url"`
    IsActive  bool   `db:"is_active"`
}