package repo

import (
	"context"
	"github.com/triapex/auth/config"
	"github.com/triapex/auth/internal/infra"
	"github.com/triapex/auth/model"
)

// InvoiceRepo returns brand repo
type InvoiceRepo interface {
	GetInvoiceById(ctx context.Context, id uint) (*model.InvoiceInfo, error)
	CreateInvoice(ctx context.Context, invoice *model.InvoiceInfo) (*model.InvoiceInfo, error)
	DeleteInvoice(ctx context.Context, id uint) error
	UpdateInvoice(ctx context.Context, invoice *model.InvoiceInfo) (*model.InvoiceInfo, error)
}

// PostgresInvoice brand repo
type PostgresInvoice struct {
	table *config.Table
	db    infra.DB
}

// NewInvoice returns new brand repo
func NewInvoice(table *config.Table, db infra.DB) InvoiceRepo {
	return &PostgresInvoice{
		table: table,
		db:    db,
	}
}

// GetInvoiceById ....
func (p *PostgresInvoice) GetInvoiceById(ctx context.Context, id uint) (*model.InvoiceInfo, error) {
	user := &model.InvoiceInfo{}
	if err := p.db.FindOne(ctx, p.table.Invoice, infra.DbQuery{
		"id": id,
	}, []string{"UserOrgMap"}, user); err != nil {
		return nil, err
	}
	return user, nil
}

// CreateInvoice ....
func (p *PostgresInvoice) CreateInvoice(ctx context.Context, user *model.InvoiceInfo) (*model.InvoiceInfo, error) {
	err := p.db.Insert(ctx, p.table.Invoice, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (p *PostgresInvoice) DeleteInvoice(ctx context.Context, id uint) error {
	return nil
}

func (p *PostgresInvoice) UpdateInvoice(ctx context.Context, invoice *model.InvoiceInfo) (*model.InvoiceInfo, error) {
	return nil, nil
}
