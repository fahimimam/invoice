package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fahimimam/invoice/internal/network"
	"github.com/fahimimam/invoice/logger"
	localModel "github.com/fahimimam/invoice/model"
	"github.com/unidoc/unipdf/v4/creator"
	"strconv"
)

const (
	CreateInvoice = "CreateInvoice"
	ReadInvoice   = "ReadInvoice"
	UpdateInvoice = "UpdateInvoice"
	DeleteInvoice = "DeleteInvoice"
)

// InvoiceServiceInterface ......
type InvoiceServiceInterface interface {
	GetInvoiceById(ctx context.Context, id string) (*localModel.InvoiceInfo, error)
	CreateInvoice(ctx context.Context, invoice *localModel.InvoiceInfo) error
	DeleteInvoice(ctx context.Context, id string) error
	UpdateInvoice(ctx context.Context, invoice *localModel.InvoiceInfo) error
	SetInvoiceIdentityGateway(certPem, keyPem, mspId string) error
	CloseInvoiceIdentityGateway() error
	GeneratePDFInvoice(invoice *localModel.InvoiceInfo) ([]byte, error)
}

type InvoiceService struct {
	invoiceNetwork network.NetworkInterface
	lgr            logger.StructLogger
}

// NewInvoiceService ...
func NewInvoiceService(invoiceNetwork network.NetworkInterface, lgr logger.StructLogger) InvoiceServiceInterface {
	return &InvoiceService{
		invoiceNetwork: invoiceNetwork,
		lgr:            lgr,
	}
}

// SetLogger ...
func (is *InvoiceService) SetLogger(lgr logger.StructLogger) {
	is.lgr = lgr
}

func (is *InvoiceService) CreateInvoice(ctx context.Context, invoiceInfo *localModel.InvoiceInfo) error {
	assetJSON, err := json.Marshal(invoiceInfo)
	if err != nil {
		return err
	}

	if err := is.invoiceNetwork.SubmitTransaction(CreateInvoice, string(assetJSON)); err != nil {
		return err
	}

	return nil
}

func (is *InvoiceService) GetInvoiceById(ctx context.Context, id string) (*localModel.InvoiceInfo, error) {
	result, err := is.invoiceNetwork.EvaluateTransaction(ReadInvoice, id)
	if err != nil {
		return nil, err
	}

	var invoice localModel.InvoiceInfo
	if err = json.Unmarshal(result, &invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

func (is *InvoiceService) DeleteInvoice(ctx context.Context, id string) error {
	return is.invoiceNetwork.SubmitTransaction(DeleteInvoice, id)
}

func (is *InvoiceService) UpdateInvoice(ctx context.Context, invoice *localModel.InvoiceInfo) error {
	assetJSON, err := json.Marshal(invoice)
	if err != nil {
		return err
	}

	if err = is.invoiceNetwork.SubmitTransaction(UpdateInvoice, string(assetJSON)); err != nil {
		return err
	}

	return nil
}

func (is *InvoiceService) SetInvoiceIdentityGateway(certPem, keyPem, mspId string) error {
	return is.invoiceNetwork.SetIdentityGateway(certPem, keyPem, mspId)
}

func (is *InvoiceService) CloseInvoiceIdentityGateway() error {
	return is.invoiceNetwork.CloseIdentityGateway()
}

func (is *InvoiceService) GeneratePDFInvoice(invoice *localModel.InvoiceInfo) ([]byte, error) {
	c := creator.New()
	c.SetPageMargins(50, 50, 50, 50) // Left, top, right, bottom margins

	// Create a new invoice instance
	pdfInvoice := c.NewInvoice()

	// Set invoice metadata :cite[1]:cite[4]
	pdfInvoice.SetNumber(strconv.Itoa(int(invoice.ID)))
	pdfInvoice.SetDate(invoice.Date)
	pdfInvoice.SetDueDate(invoice.DueDate)
	pdfInvoice.AddInfo("Status", invoice.Status)

	// Set addresses :cite[4]
	pdfInvoice.SetSellerAddress(&creator.InvoiceAddress{
		Name:    invoice.SellerAddress.Name,
		Street:  invoice.SellerAddress.Street,
		City:    invoice.SellerAddress.City,
		Zip:     invoice.SellerAddress.Zip,
		Country: invoice.SellerAddress.Country,
		Phone:   invoice.SellerAddress.Phone,
		Email:   invoice.SellerAddress.Email,
	})

	pdfInvoice.SetBuyerAddress(&creator.InvoiceAddress{
		Name:    invoice.BuyerAddress.Name,
		Street:  invoice.BuyerAddress.Street,
		City:    invoice.BuyerAddress.City,
		Zip:     invoice.BuyerAddress.Zip,
		Country: invoice.BuyerAddress.Country,
		Phone:   invoice.BuyerAddress.Phone,
		Email:   invoice.BuyerAddress.Email,
	})

	// Add line items :cite[1]
	for _, item := range invoice.Items {
		pdfInvoice.AddLine(
			item.Description,
			fmt.Sprintf("%d", item.Qty),
			fmt.Sprintf("%.2f", item.Rate),
		)
	}

	// Set totals :cite[1]:cite[4]
	pdfInvoice.SetSubtotal(fmt.Sprintf("%.2f", invoice.Subtotal))

	pdfInvoice.SetTotal(fmt.Sprintf("%.2f", invoice.Total))

	// Add notes and terms if available
	if invoice.Notes != "" {
		pdfInvoice.SetNotes("Notes", invoice.Notes)
	}

	// Customize invoice styling :cite[4]
	customizeInvoiceStyle2(pdfInvoice)

	// Draw invoice to creator
	if err := c.Draw(pdfInvoice); err != nil {
		return nil, fmt.Errorf("failed to draw invoice: %v", err)
	}

	// Write to buffer instead of file
	var buf bytes.Buffer
	if err := c.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %v", err)
	}

	return buf.Bytes(), nil
}
