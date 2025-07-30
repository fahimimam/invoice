package service

import (
	"context"
	"errors"
	"github.com/triapex/auth/internal/infra"
	"github.com/triapex/auth/internal/repo"
	"github.com/triapex/auth/logger"
	"github.com/triapex/auth/model"
	"github.com/triapex/auth/utils"
)

// InvoiceService interface
type InvoiceService interface {
	CreateInvoice(ctx context.Context, invoice *model.CreateInvoicePayload) (*model.InvoiceInfo, error)
	GetInvoiceByID(ctx context.Context, id uint) (*model.InvoiceInfo, error)
}

//Login(ctx context.Context, user *model.LoginRequest) (*model.Token, error)

// Invoice ...
type Invoice struct {
	invoiceRepo repo.InvoiceRepo
	cache       infra.KV
	log         logger.StructLogger
}

// NewInvoice ...
func NewInvoice(invoiceRepo repo.InvoiceRepo,
	cache infra.KV,
	lgr logger.StructLogger) InvoiceService {
	return &Invoice{
		cache:       cache,
		log:         lgr,
		invoiceRepo: invoiceRepo,
	}
}

// SetLogger ...
func (u *Invoice) SetLogger(l logger.StructLogger) {
	u.log = l
}

func (u *Invoice) CreateInvoice(ctx context.Context, userReq *model.CreateInvoicePayload) (*model.InvoiceInfo, error) {
	tid := utils.GetTracingID(ctx)
	u.log.Println("CreateInvoice", tid, "Request for signup from service")

	return &model.InvoiceInfo{}, errors.New("not implemented")
}
func (u *Invoice) GetInvoiceByID(ctx context.Context, id uint) (*model.InvoiceInfo, error) {
	tid := utils.GetTracingID(ctx)
	u.log.Println("CreateInvoice", tid, "Request for signup from service")

	return &model.InvoiceInfo{}, nil
}
