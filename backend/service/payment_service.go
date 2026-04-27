package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/irham/topup-backend/kafka"
	"github.com/irham/topup-backend/model"
	"github.com/irham/topup-backend/pkg"
	"github.com/irham/topup-backend/repository"
)

type PaymentService struct {
	payments *repository.PaymentRepo
	orders   *repository.OrderRepo
	midtrans *pkg.MidtransClient
	producer *kafka.Producer
}

func NewPaymentService(
	payments *repository.PaymentRepo,
	orders *repository.OrderRepo,
	midtrans *pkg.MidtransClient,
	producer *kafka.Producer,
) *PaymentService {
	return &PaymentService{
		payments: payments, orders: orders,
		midtrans: midtrans, producer: producer,
	}
}

var (
	ErrInvalidSignature = errors.New("signature midtrans tidak valid")
	ErrPaymentNotFound  = errors.New("payment tidak ditemukan")
)

// HandleWebhook processes Midtrans notification body.
// raw must be the original request bytes (for audit + signature inputs are derived from parsed fields).
func (s *PaymentService) HandleWebhook(ctx context.Context, raw json.RawMessage) error {
	var payload model.MidtransWebhookPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("parse webhook: %w", err)
	}
	if payload.OrderID == "" {
		return errors.New("order_id kosong di webhook")
	}

	statusCode := statusCodeFor(payload.TransactionStatus)
	sigOK := s.midtrans.VerifyNotificationSignature(
		payload.OrderID, statusCode, payload.GrossAmount, payload.SignatureKey,
	)
	if !sigOK {
		// tetap simpan webhook untuk audit, tapi tandai unverified
		_ = s.payments.UpdateFromWebhook(ctx,
			payload.OrderID, payload.TransactionID, payload.PaymentType,
			payload.TransactionStatus, payload.FraudStatus, false, raw,
		)
		return ErrInvalidSignature
	}

	if err := s.payments.UpdateFromWebhook(ctx,
		payload.OrderID, payload.TransactionID, payload.PaymentType,
		payload.TransactionStatus, payload.FraudStatus, true, raw,
	); err != nil {
		return err
	}

	pmt, err := s.payments.GetByMidtransOrderID(ctx, payload.OrderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrPaymentNotFound
		}
		return err
	}

	newOrderStatus, reason := mapTransactionToOrderStatus(payload.TransactionStatus, payload.FraudStatus)
	if newOrderStatus != "" {
		if err := s.orders.UpdateStatus(ctx, pmt.OrderID, newOrderStatus, reason); err != nil {
			return err
		}
	}

	_ = s.producer.Publish(ctx, kafka.TopicPayments, pmt.OrderID.String(), map[string]any{
		"event":              "midtrans_webhook",
		"order_id":           pmt.OrderID,
		"midtrans_order_id":  payload.OrderID,
		"transaction_status": payload.TransactionStatus,
		"fraud_status":       payload.FraudStatus,
		"payment_type":       payload.PaymentType,
		"gross_amount":       payload.GrossAmount,
	})

	return nil
}

// statusCodeFor returns the Midtrans status_code expected for a given transaction_status.
// Values per Midtrans docs.
func statusCodeFor(status string) string {
	switch status {
	case "capture", "settlement":
		return "200"
	case "pending":
		return "201"
	case "deny":
		return "202"
	case "cancel", "expire":
		return "202"
	default:
		return "200"
	}
}

// mapTransactionToOrderStatus converts Midtrans transaction_status to internal order_status.
// Returns (status, failureReason). status="" means do not change.
func mapTransactionToOrderStatus(status, fraud string) (string, string) {
	switch status {
	case "capture":
		if fraud == "challenge" {
			return "", "" // tetap awaiting, tunggu review
		}
		return "paid", ""
	case "settlement":
		return "paid", ""
	case "pending":
		return "awaiting_payment", ""
	case "deny":
		return "failed", "Pembayaran ditolak"
	case "cancel":
		return "cancelled", "Dibatalkan"
	case "expire":
		return "expired", "Pembayaran kedaluwarsa"
	case "refund", "partial_refund":
		return "refunded", ""
	case "failure":
		return "failed", "Transaksi gagal"
	}
	return "", ""
}

// helpers — kept here in case future webhooks send numeric fields as strings
func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

var _ = parseInt
