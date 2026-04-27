package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/irham/topup-backend/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepo struct {
	db *pgxpool.Pool
}

func NewPaymentRepo(db *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) Create(ctx context.Context, p *model.PaymentTransaction) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO payment_transactions (
			order_id, env, midtrans_order_id,
			gross_amount, transaction_status,
			snap_token, snap_redirect_url
		) VALUES ($1, $2::payment_env, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''))
		RETURNING id, created_at
	`,
		p.OrderID, p.Env, p.MidtransOrderID,
		p.GrossAmount, p.TransactionStatus,
		p.SnapToken, p.SnapRedirectURL,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return fmt.Errorf("payment create: %w", err)
	}
	return nil
}

func (r *PaymentRepo) GetByMidtransOrderID(ctx context.Context, midtransOrderID string) (*model.PaymentTransaction, error) {
	var p model.PaymentTransaction
	err := r.db.QueryRow(ctx, `
		SELECT id, order_id, env::text, midtrans_order_id,
		       COALESCE(transaction_id, ''), COALESCE(payment_type, ''),
		       gross_amount, COALESCE(transaction_status, ''), COALESCE(fraud_status, ''),
		       COALESCE(snap_token, ''), COALESCE(snap_redirect_url, ''),
		       COALESCE(va_number, ''), COALESCE(bank, ''), COALESCE(qr_code_url, ''),
		       signature_verified, transaction_time, settlement_time, expiry_time, created_at
		FROM payment_transactions
		WHERE midtrans_order_id = $1
	`, midtransOrderID).Scan(
		&p.ID, &p.OrderID, &p.Env, &p.MidtransOrderID,
		&p.TransactionID, &p.PaymentType,
		&p.GrossAmount, &p.TransactionStatus, &p.FraudStatus,
		&p.SnapToken, &p.SnapRedirectURL,
		&p.VANumber, &p.Bank, &p.QRCodeURL,
		&p.SignatureVerified, &p.TransactionTime, &p.SettlementTime, &p.ExpiryTime, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("payment get: %w", err)
	}
	return &p, nil
}

func (r *PaymentRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.PaymentTransaction, error) {
	var p model.PaymentTransaction
	err := r.db.QueryRow(ctx, `
		SELECT id, order_id, env::text, midtrans_order_id,
		       COALESCE(transaction_id, ''), COALESCE(payment_type, ''),
		       gross_amount, COALESCE(transaction_status, ''), COALESCE(fraud_status, ''),
		       COALESCE(snap_token, ''), COALESCE(snap_redirect_url, ''),
		       COALESCE(va_number, ''), COALESCE(bank, ''), COALESCE(qr_code_url, ''),
		       signature_verified, transaction_time, settlement_time, expiry_time, created_at
		FROM payment_transactions
		WHERE order_id = $1
	`, orderID).Scan(
		&p.ID, &p.OrderID, &p.Env, &p.MidtransOrderID,
		&p.TransactionID, &p.PaymentType,
		&p.GrossAmount, &p.TransactionStatus, &p.FraudStatus,
		&p.SnapToken, &p.SnapRedirectURL,
		&p.VANumber, &p.Bank, &p.QRCodeURL,
		&p.SignatureVerified, &p.TransactionTime, &p.SettlementTime, &p.ExpiryTime, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("payment get by order: %w", err)
	}
	return &p, nil
}

// UpdateFromWebhook persists fields received in the Midtrans webhook payload.
// raw should be the original notification body for audit.
func (r *PaymentRepo) UpdateFromWebhook(
	ctx context.Context,
	midtransOrderID string,
	transactionID string,
	paymentType string,
	transactionStatus string,
	fraudStatus string,
	signatureVerified bool,
	raw json.RawMessage,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE payment_transactions
		SET transaction_id      = NULLIF($2, ''),
		    payment_type        = NULLIF($3, ''),
		    transaction_status  = NULLIF($4, ''),
		    fraud_status        = NULLIF($5, ''),
		    signature_verified  = $6,
		    raw_notification    = $7,
		    webhook_received_at = NOW(),
		    settlement_time     = CASE WHEN $4 = 'settlement' THEN NOW() ELSE settlement_time END
		WHERE midtrans_order_id = $1
	`, midtransOrderID, transactionID, paymentType, transactionStatus, fraudStatus, signatureVerified, raw)
	if err != nil {
		return fmt.Errorf("payment update from webhook: %w", err)
	}
	return nil
}
