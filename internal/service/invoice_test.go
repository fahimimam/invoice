package service

import (
	"fmt"
	"github.com/fahimimam/invoice/internal/network"
	"github.com/fahimimam/invoice/logger"
	localModel "github.com/fahimimam/invoice/model"
	"github.com/unidoc/unipdf/v4/common/license"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestInvoiceService_GeneratePDFInvoice(t *testing.T) {
	type fields struct {
		invoiceNetwork network.NetworkInterface
		lgr            logger.StructLogger
	}
	type args struct {
		invoice *localModel.InvoiceInfo
	}
	var items []localModel.Item
	items = append(items, localModel.Item{
		InvoiceItemId: "2",
		ItemId:        "3",
		Taxable:       "yes",
		Rate:          30,
		ItemName:      "Item 1",
		Description:   "This is a good Item",
		Qty:           3,
		Tags:          "Raw Item",
	})

	// Make sure to load your metered License API key prior to using the library.
	// If you need a key, you can sign up and create a free one at https://cloud.unidoc.io
	err := license.SetMeteredKey("dae528f86ed66188eec903e853b1f31dd43ce4a926c82a4b41c2339a83223cf3")
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				invoiceNetwork: nil,
				lgr:            logger.DefaultOutStructLogger,
			},

			args: args{
				invoice: &localModel.InvoiceInfo{
					Model: &gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					UserId:         "1",
					ClientId:       "1",
					InvoiceDate:    "2002-01-02",
					InvoiceId:      1,
					DueDate:        "2002-01-05",
					Date:           "2002-01-05",
					Total:          50,
					Subtotal:       60,
					Vat:            "10",
					Paid:           "60",
					Status:         "Paid",
					IsPaid:         "Yes",
					Comment:        "NO",
					CardAcceptable: "Yes",
					SellerAddress: localModel.InvoiceAddress{
						Name:    "Fahim",
						Street:  "3",
						Zip:     "7800",
						City:    "Dhaka",
						Country: "Bangladesh",
						Phone:   "017635368111",
						Email:   "fhaim@gmail.com",
					},
					BuyerAddress: localModel.InvoiceAddress{
						Name:    "Sagor",
						Street:  "2",
						Zip:     "1212",
						City:    "Dhaka",
						Country: "Bangladesh",
						Phone:   "0192748234",
						Email:   "Sagor@gmail.com",
					},
					Notes: "No notes",
					Items: items,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &InvoiceService{
				invoiceNetwork: tt.fields.invoiceNetwork,
				lgr:            tt.fields.lgr,
			}
			bytes, err := is.GeneratePDFInvoice(tt.args.invoice)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePDFInvoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			fmt.Println(len(bytes))
		})
	}
}
