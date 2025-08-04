package service

import (
	"bytes"
	"fmt"
	localModel "github.com/fahimimam/invoice/model"
	"github.com/unidoc/unipdf/v4/creator"
	"github.com/unidoc/unipdf/v4/model"
	"strconv"
)

func generatePDFInvoice(invoice *localModel.InvoiceInfo) ([]byte, error) {
	// Initialize PDF creator
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

func customizeInvoiceStyle(invoice *creator.Invoice) {
	// Load fonts :cite[4]
	fontHelvetica, _ := model.NewStandard14Font("Helvetica")
	fontHelveticaBold, _ := model.NewStandard14Font("Helvetica-Bold")
	//fontHelvetica := model.NewStandard14FontMustCompile(model.NewStandard14Font("Helvetica"))
	//fontHelveticaBold := model.NewStandard14FontMustCompile(fonts.HelveticaBoldName)

	// Create colors
	primaryColor := creator.ColorRGBFrom8bit(33, 150, 243) // Blue
	//accentColor := creator.ColorRGBFrom8bit(255, 87, 34)   // Orange
	lightGray := creator.ColorRGBFrom8bit(245, 245, 245)

	// Set title style
	invoice.SetTitleStyle(creator.TextStyle{
		Font:     fontHelveticaBold,
		FontSize: 24,
		Color:    primaryColor,
	})

	// Set address heading style
	invoice.SetAddressHeadingStyle(creator.TextStyle{
		Font:     fontHelveticaBold,
		FontSize: 12,
		Color:    primaryColor,
	})

	// Style columns
	for _, col := range invoice.Columns() {
		col.BackgroundColor = lightGray
		col.BorderColor = lightGray
		col.TextStyle.Font = fontHelveticaBold
		col.TextStyle.FontSize = 10
		col.Alignment = creator.CellHorizontalAlignmentCenter
	}

	// Style lines
	for _, line := range invoice.Lines() {
		for _, cell := range line {
			cell.BorderColor = lightGray
			cell.TextStyle.Font = fontHelvetica
			cell.TextStyle.FontSize = 9
		}
	}

	// Style totals
	titleCell, contentCell := invoice.Total()
	titleCell.BackgroundColor = lightGray
	titleCell.BorderColor = lightGray
	titleCell.TextStyle.Font = fontHelveticaBold
	contentCell.BackgroundColor = lightGray
	contentCell.BorderColor = lightGray
	contentCell.TextStyle.Font = fontHelveticaBold
}
