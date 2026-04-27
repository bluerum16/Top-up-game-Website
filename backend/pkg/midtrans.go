package pkg

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/irham/topup-backend/config"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransClient struct {
	cfg    config.MidtransConfig
	snap   snap.Client
	envTag string
}

func NewMidtransClient(cfg config.MidtransConfig) *MidtransClient {
	env := midtrans.Sandbox
	envTag := "sandbox"
	if cfg.IsProduction() {
		env = midtrans.Production
		envTag = "production"
	}

	var s snap.Client
	s.New(cfg.ServerKey, env)

	return &MidtransClient{cfg: cfg, snap: s, envTag: envTag}
}

func (m *MidtransClient) Env() string { return m.envTag }

type SnapRequest struct {
	OrderID     string
	GrossAmount int64
	CustomerName string
	CustomerEmail string
	CustomerPhone string
	ItemName    string
}

type SnapResult struct {
	Token       string
	RedirectURL string
}

func (m *MidtransClient) CreateSnapTransaction(req SnapRequest) (*SnapResult, error) {
	body := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.OrderID,
			GrossAmt: req.GrossAmount,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: req.CustomerName,
			Email: req.CustomerEmail,
			Phone: req.CustomerPhone,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    req.OrderID,
				Name:  req.ItemName,
				Price: req.GrossAmount,
				Qty:   1,
			},
		},
	}

	resp, err := m.snap.CreateTransaction(body)
	if err != nil {
		return nil, fmt.Errorf("midtrans CreateTransaction: %w", err)
	}
	return &SnapResult{Token: resp.Token, RedirectURL: resp.RedirectURL}, nil
}

// VerifyNotificationSignature checks SHA512(orderID+statusCode+grossAmount+serverKey).
// grossAmount must match the raw string from Midtrans payload (e.g. "10000.00").
func (m *MidtransClient) VerifyNotificationSignature(orderID, statusCode, grossAmount, signature string) bool {
	raw := orderID + statusCode + grossAmount + m.cfg.ServerKey
	sum := sha512.Sum512([]byte(raw))
	return hex.EncodeToString(sum[:]) == signature
}
