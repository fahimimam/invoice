package service

import (
	"github.com/unidoc/unipdf/v4/creator"
	"github.com/unidoc/unipdf/v4/model"
)

func customizeInvoiceStyle(invoice *creator.Invoice) {
	// Load fonts :cite[4]
	fontHelvetica, _ := model.NewStandard14Font("Helvetica")
	fontHelveticaBold, _ := model.NewStandard14Font("Helvetica-Bold")

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
