package model

import "gorm.io/gorm"

type GoogleSSOClaims struct {
	Email string `json:"email"`
}

type CreateInvoicePayload struct {
	InvoiceInfo *InvoiceInfo `json:"invoice_info"`
}

// InvoiceInfo define user details
type InvoiceInfo struct {
	*gorm.Model
	UserId         string         `json:"userId"`
	ClientId       string         `json:"clientId"`
	InvoiceDate    string         `json:"invoiceDate"`
	InvoiceId      int            `json:"invoiceId"`
	DueDate        string         `json:"dueDate"`
	Date           string         `json:"date"`
	Total          float64        `json:"total"`
	Subtotal       float64        `json:"subtotal"`
	Vat            string         `json:"vat"`
	Paid           string         `json:"paid"`
	Status         string         `json:"status"`
	IsPaid         string         `json:"isPaid"`
	Comment        string         `json:"comment"`
	CardAcceptable string         `json:"cardAcceptable"`
	SellerAddress  InvoiceAddress `json:"sellerAddress"`
	BuyerAddress   InvoiceAddress `json:"buyerAddress"`
	Notes          string         `json:"notes"`
	Items          []Item         `json:"items"`
}

type Item struct {
	InvoiceItemId string `json:"invoiceItemId"`
	ItemId        string `json:"itemId"`
	Taxable       string `json:"taxable"`
	Rate          int    `json:"rate"`
	ItemName      string `json:"itemName"`
	Description   string `json:"description"`
	Qty           int    `json:"qty"`
	Tags          string `json:"tags"`
}

type InvoiceAddress struct {
	Name    string
	Street  string
	Zip     string
	City    string
	Country string
	Phone   string
	Email   string
}
