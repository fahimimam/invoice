package model

import (
	"github.com/lib/pq"
	"github.com/triapex/auth/utils"
	"gorm.io/gorm"
)

type GoogleSSOClaims struct {
	Email string `json:"email"`
}

type CreateInvoicePayload struct {
	InvoiceInfo *InvoiceInfo `json:"invoice_info"`
}

// InvoiceInfo define user details
type InvoiceInfo struct {
	gorm.Model
	FirstName string         `json:"first_name" gorm:"column:first_name"`
	LastName  string         `json:"last_name" gorm:"column:last_name"`
	Phone     string         `json:"phone" gorm:"column:phone;unique"`
	Email     string         `json:"email" gorm:"column:email;unique"`
	Password  string         `json:"password" gorm:"column:password"`
	Type      utils.UserType `json:"type" gorm:"column:type"`
	Roles     pq.StringArray `json:"roles" gorm:"type:text[];column:roles"`
	Verified  bool           `json:"verified" gorm:"column:verified"`
}
